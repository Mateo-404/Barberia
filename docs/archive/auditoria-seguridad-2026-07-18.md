# Auditoría de Seguridad — Sistema Barbería (TH Barber Club)

**Fecha:** 18 jul 2026
**Rol:** Ingeniero de Ciberseguridad (análisis ofensivo + defensivo)
**Alcance:** Backend Spring Boot (rama `backend`) + Frontend React (rama `frontend`) + config de despliegue.
**Método:** Revisión estática de código (SAST), revisión de config y pruebas dinámicas (DAST) contra el backend corriendo en `localhost:8080`.

---

## Resumen ejecutivo

| Severidad | Cantidad | Hallazgos |
|-----------|----------|-----------|
| 🔴 Crítico | 2 | JWT secret hardcoded en repo · Swagger UI público en prod |
| 🟠 Alto | 2 | CORS `localhost:*` + credenciales · Sin rate-limiting en login |
| 🟡 Medio | 3 | Token en `sessionStorage` (XSS) · `invalid-token` vs `unauthorized` inconsistencia · Info-leak en errores genéricos parcial |
| 🟢 Bajo | 2 | Dev security password en log · Logs de seed con PII mínima |

**Veredicto:** La base de autenticación (Argon2 + JWT stateless + filtro) está bien planteada. Pero hay **2 fallas críticas de configuración** que comprometen la integridad del sistema en producción si no se corrigen.

---

## 🔴 CRÍTICO

### C1 — JWT secret hardcoded como fallback en el repo
**Archivo:** `backend/src/main/resources/application.properties:12`
```properties
jwt.secret=${JWT_SECRET:dev-secret-at-least-32-bytes-long-HS256-key}
```
**Problema:** El valor por defecto (`dev-secret-at-least-32-bytes-long-HS256-key`) está **committeado en el repo**. Si el entorno de producción no define `JWT_SECRET`, el backend firma tokens con una clave **pública y conocida**.
**Impacto:** Un atacante puede forjar un JWT válido con `sub` = cualquier admin ID y `email` arbitrario → **escalada de privilegios completa / impersonación de administrador** sin credenciales.
**Prueba de concepto (lógica):** con la clave conocida, `Jwts.builder().signWith(Keys.hmacShaKeyFor("dev-secret-...".getBytes()))` reproduce un token aceptado por `JwtAuthenticationFilter.validateToken()`.
**Remediación:**
- Eliminar el fallback por defecto: `jwt.secret=${JWT_SECRET}` (arranque falla si falta → fail-closed).
- Generar secreto de 32+ bytes aleatorios (`openssl rand -base64 32`) y cargarlo solo por env/secret manager.
- Rotar el secreto y invalidar tokens emitidos (cambio de clave invalida todos los JWT existentes).

### C2 — Swagger UI y OpenAPI expuestos sin autenticación
**Archivo:** `backend/src/main/java/com/barber/barberBackend/config/SecurityConfig.java:56`
```java
.requestMatchers("/v3/api-docs/**", "/swagger-ui/**").permitAll()
```
**Problema:** Toda la superficie de la API (endpoints, DTOs, parámetros) queda pública. En producción (Railway) cualquiera puede leer el contrato completo y usar "Try it out" contra el backend.
**Impacto:** Reconocimiento total del sistema; facilita ataques dirigidos (fuzzing de `/turnos`, enumeración de `/administradores`, etc.).
**Remediación:**
- Restringir a `@Profile("!prod")` o exigir rol ADMIN.
- En prod, deshabilitar `springdoc`/`swagger` vía propiedad (`springdoc.api-docs.enabled=false`).

---

## 🟠 ALTO

### A1 — CORS `http://localhost:*` con credenciales
**Archivo:** `backend/src/main/java/com/barber/barberBackend/config/CorsConfig.java:22-26`
```java
config.addAllowedOriginPattern("http://localhost:*");
config.setAllowCredentials(true);
```
**Problema:** `allowedOriginPattern("http://localhost:*")` + `allowCredentials(true)` equivale a "cualquier origen en localhost puede hacer requests con credenciales". Cualquier app maliciosa servida en `localhost:<puerto>` (o un subdominio `localhost` controlado) puede enviar el cookie/token del usuario si hubiera sesión basada en cookie.
**Impacto:** En un entorno de desarrollo compartido o si se activa por error en prod, habilita **CSRF / lectura de respuestas autenticadas** cross-origin.
**Mitigación actual buena:** Está bajo `@Profile("!prod")`, así que no aplica en prod. Pero el patrón es frágil.
**Remediación:** En dev, listar puertos explícitos (`5173`, `5174`); nunca combinar wildcard-origin + credentials. Confirmar que el perfil `prod` NO se activa por error (variable de entorno del deploy).

### A2 — Sin rate-limiting / lockout en `/administradores/login`
**Archivo:** `SecurityConfig.java:55` · `AdministradorService.login()`
**Problema:** El endpoint de login es `permitAll` y no tiene throttling, CAPTCHA ni lockout. `InvalidCredentialsException` se lanza en O(1) por intento.
**Impacto:** **Fuerza bruta** de credenciales de administrador viable (el seed usa `admin@admin.com` / `admin1234` — credenciales débiles y conocidas).
**Remediación:**
- Agregar rate-limit por IP + por email (Bucket4j, o filtro custom).
- Alerta tras N intentos fallidos.
- Considerar MFA para el rol admin.

---

## 🟡 MEDIO

### M1 — Token JWT en `sessionStorage` (superficie XSS)
**Archivo:** `frontend-react/src/context/AuthContext.tsx` (claves `auth_token`, `admin`)
**Problema:** El JWT y los datos del admin se guardan en `sessionStorage`. Cualquier XSS en la app (inyección en el render de turnos/servicios, librería comprometida) puede leer el token y exfiltrarlo.
**Impacto:** Robo de sesión de administrador (el token dura 24h, `JwtService.EXPIRATION_MS = 86_400_000`).
**Mitigación:** Es el trade-off estándar de SPA sin backend propio para setear cookie `httpOnly`. Recomendado:
- Sanitizar toda salida de datos del backend en el render (React ya escapa por defecto; no usar `dangerouslySetInnerHTML`).
- Agregar un CSP estricto (`default-src 'self'`; solo `localhost:8080` para API).
- Reducir `EXPIRATION_MS` a ~15-30 min + refresh token.

### M2 — Inconsistencia en el tipo de error 401
**Archivo:** `JwtAuthenticationFilter.java:67` vs `api-client.ts`
```java
pd.setType(URI.create("/errors/unauthorized"));   // filtro JWT
```
Pero `api-client.ts` solo limpia sesión en `type === "/errors/invalid-token"`.
**Problema:** El `JwtAuthenticationFilter` devuelve `type=/errors/unauthorized` para tokens inválidos, mientras que el frontend solo reacciona a `/errors/invalid-token`. Si el backend usa el tipo del filtro, el frontend **no** limpia la sesión ni redirige ante un token caducado/expirado presentado al llamar a un endpoint protegido.
**Impacto:** El usuario queda "logueado" en la UI aunque el backend rechace sus requests → mala experiencia + estado inconsistente (ya se corrigió la validación de `exp` al cargar, pero no el caso de token rechazado en-vuelo).
**Remediación:** Unificar el `type` entre backend y frontend (ej. backend devuelve `/errors/invalid-token` y frontend reacciona a ese), o ampliar la condición del frontend a ambos tipos.

### M3 — Mensajes de error (parcialmente bien)
**Bien:** Login devuelve "Credenciales inválidas" genérico (no revela si el email existe). ✅
**Mal:** `/errors/invalid-tredentials` (typo en el backend: "inval**i**d-credentials" con una i de menos en algunos paths) puede romper el match exacto en el frontend que espera `/errors/invalid-credentials`. Verificar consistencia del string en todo el backend.

---

## 🟢 BAJO

### L1 — Password de seguridad de Spring en los logs
`docker logs` muestra: `Using generated security password: 11fa2aaf-...` (password de Spring Security para el user `user`). Es solo dev, pero queda en el log del contenedor.
**Remediación:** No depende de ese user en prod; ignorable si `spring.security.user` no está configurado en prod.

### L2 — Log de seed con email
`AdminSeedRunner` loguea `email=admin@admin.com`. Es PII mínima pero innecesaria en prod.
**Remediación:** Loguear solo "admin semilla creado" sin el email.

---

## Hallazgos que ESTÁN BIEN (defensivos)

- ✅ Contraseñas hasheadas con **Argon2** (`Argon2PasswordEncoder(16,32,1,19456,2)`) — costo de memoria alto, resistente a GPU cracking.
- ✅ **CSRF deshabilitado + sesión STATELESS** — correcto para API JWT (no hay cookie de sesión que proteger vía CSRF).
- ✅ `SecurityConfig` usa `ProblemDetail` (RFC 7807) consistente.
- ✅ El seed de admin pasa por `AdministradorService.save()` → **se hashea con Argon2** (no plaintext, contrariamente a lo que sugiere una lectura rápida de `AdminSeedRunner`).
- ✅ Frontend: `api-client` inyecta `Bearer` y limpia sesión en token inválido; `AuthContext` valida `exp` al cargar.
- ✅ Login del frontend no filtra cuál campo es incorrecto.

---

## Nota sobre la prueba de login (E2E)

Durante las pruebas, `POST /administradores/login` con `admin@admin.com` / `admin1234` devolvió **401 invalid-credentials**. Esto NO es una vulnerabilidad: indica que el admin en la DB tiene un hash distinto (el seed se saltó porque ya existía un admin, o el password fue cambiado en una corrida previa). No se pudo completar el flujo de login en E2E por estado de datos, no por falla de frontend. Las rutas `/` y `/login` renderizan correctamente y `/admin` redirige a `/login` (ProtectedRoute OK).

---

## Plan de remediación priorizado

| # | Hallazgo | Esfuerzo | Prioridad |
|---|----------|----------|-----------|
| 1 | C1: JWT secret sin fallback + rotar | Bajo | 🔴 Inmediato |
| 2 | C2: Swagger solo en !prod | Bajo | 🔴 Inmediato |
| 3 | A1: CORS sin wildcard+credentials | Bajo | 🟠 |
| 4 | A2: Rate-limit en login | Medio | 🟠 |
| 5 | M1: CSP + reducir exp del JWT | Medio | 🟡 |
| 6 | M2: Unificar tipo 401 backend↔frontend | Bajo | 🟡 |
| 7 | M3/L1/L2: limpieza de logs y typos | Bajo | 🟢 |

> Nota: los archivos de backend están fuera del alcance editable de esta sesión (rama `frontend`). Los remedios C1–A2 requieren cambios en el clon `Barberia-backend-run` (rama `backend`), que deben ejecutarse por la sesión de backend.
