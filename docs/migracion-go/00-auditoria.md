# 00 — Auditoría del backend Java (evidencia)

Fecha: 2026-10-21. Todo lo documentado aquí fue verificado contra el código del repositorio
y, cuando se indica, contra el backend Java real ejecutado con el perfil `h2`
(spring-boot:run, puerto 8080, datos demo de `DemoDataSeedRunner`).

## 1. Estado inicial

- Rama: `main` (sincronizada con `origin/main`), único cambio no confirmado: `.gitignore`
  (ruta `frontend-react/` → `frontend/`), preservado sin tocar.
- Baseline Java: `./mvnw test` → **Tests run: 64, Failures: 0, Errors: 0 — BUILD SUCCESS** (Java 21, H2 en memoria).
- No hay Docker ni `docker-compose.yml` en el repo (solo una propuesta histórica en
  `docs/historico/informe-sesion-completo.md`). No hay migraciones versionadas:
  el esquema lo genera `spring.jpa.hibernate.ddl-auto=update`.
- CI existente: `.github/workflows/backend-ci.yml` (Maven + Java 21, sin Docker ni frontend).

## 2. Stack actual

| Área | Tecnología | Versión |
|---|---|---|
| Runtime | Java (Temurin) | 21 |
| Framework | Spring Boot (parent) | 3.4.0 |
| Persistencia | Spring Data JPA (Hibernate) + driver `org.postgresql` | managed |
| Seguridad | Spring Security + `spring-security-crypto` (Argon2) | managed |
| JWT | `io.jsonwebtoken` (jjwt api/impl/jackson) | 0.12.6 |
| OpenAPI | springdoc-openapi-starter-webmvc-ui | 2.7.0 |
| DB dev/test | H2 en memoria (perfil `h2`/`test`) | managed |
| Extras | Lombok, BouncyCastle (`bcprov-jdk18on`, soporte Argon2) | 1.78.1 |

## 3. Modelo de datos (JPA → esquema Hibernate)

Convención de nombres de Spring Boot: tablas en snake_case (`cliente`, `turno`,
`servicio`, `administrador`), columnas snake_case (`fecha_hora`, `cliente_id`,
`servicio_id`).

| Entidad | PK | Campos | Notas |
|---|---|---|---|
| `Cliente` | `telefono` (String, natural, sin autoincremento) | hereda `Persona` (nombre, apellido, email) | sin índice de unicidad adicional |
| `Servicio` | `id` bigserial (`IDENTITY`) | `tipo`, `precio` (`float` → real/float4) | |
| `Administrador` | `id` bigserial (`IDENTITY`) | hereda `Persona` + `contrasenia` (hash Argon2) | índice `idx_admin_email` sobre `email` (no unique) |
| `Turno` | `id` — `GenerationType.AUTO` → en PostgreSQL Hibernate 6 usa la secuencia `hibernate_sequence` | `fecha_hora` NOT NULL, FK `cliente_id`, FK `servicio_id` (ManyToOne sin cascade) | índice `idx_turno_fechahora` sobre `fecha_hora` |

No hay constraints de unicidad sobre `turnos.fecha_hora`: la regla "un turno por
fecha/hora" vive únicamente en el código (`TurnoService.save`) y es **NO atómica**
(ver §6).

`Administrador.contrasenia` guarda el hash Argon2 con parámetros
`Argon2PasswordEncoder(16, 32, 1, 19456, 2)` → formato
`$argon2id$v=19$m=19456,t=2,p=1$<salt-b64>$<hash-b64>` (verificado en BD real:
`GET /administradores/0/1` expone el hash; ver §7).

## 4. Contrato de API (verificado en ejecución)

`backend/openapi-spec.json` (OpenAPI 3.0.1) documenta 24 paths / 36 operaciones.
El frontend React solo consume 9 operaciones; las 27 restantes son API pública
completa (CRUD genérico). Comportamientos medidos con el backend real:

### 4.1 Auth y errores

- `POST /administradores/login` → 200 `LoginResponseDTO{id,nombre,apellido,email,token}`.
  Token JWT HS256, claims exactos: `sub` (id admin como string), `email`, `iat`, `exp` (+24 h).
- 401 sin token (rutas protegidas): `{"type":"/errors/unauthorized","title":"No autorizado","status":401,"detail":"Se requiere autenticación para acceder a este recurso"}` — **sin** campo `instance`.
- 401 con token inválido (cualquier prefijo `Bearer ` malo, incluso en rutas públicas):
  mismo shape, `detail:"Token JWT inválido o expirado"`.
- 401 login fallido: `{"type":"/errors/invalid-credentials","title":"Credenciales inválidas","status":401,"detail":"Credenciales inválidas","instance":"/administradores/login"}`.
- 400 de validación (`@Valid`) y de JSON ilegible: **mismo shape**
  `{"type":"about:blank","title":"Bad Request","status":400,"detail":"Invalid request content.","instance":"<ruta>"}`.
  No hay lista de campos con errores (Spring 6.2).
- 400 regla de negocio: `{"type":"/errors/business-rule-violation","title":"Violación de regla de negocio","status":400,"detail":"<mensaje>","instance":"<ruta>"}`.
- 404 recurso inexistente en GET/PATCH/DELETE/exists por ID genérico: **404 con cuerpo vacío**.
- 404 con cuerpo (`/errors/resource-not-found`) solo cuando lo lanza un servicio:
  servicio inexistente al crear turno (`"No existe un servicio con el id X"`).
- 409 duplicado: `/errors/resource-already-exists` con `detail` específico
  (`"Ya existe un cliente con el teléfono X"` / `"Ya existe un administrador con el email X"`).
- 500: `{"type":"/errors/internal-server-error","title":"Error interno del servidor","status":500,"detail":"Ocurrió un error inesperado. Intente nuevamente más tarde.","instance":"<ruta>"}`.

### 4.2 Formatos de fecha y orden

- `fechaHora` en JSON: ISO-8601 local sin zona (`"2026-10-22T15:00:00"`, con segundos).
- `GET /turnos/findDateTimes` serializa con `LocalDateTime.toString()` → **sin segundos**
  (`"2026-10-22T15:00"`), solo turnos `>= ahora`, orden ascendente observado.
- `GET /{recurso}` y `/{recurso}/{page}/{size}` no tienen `ORDER BY`; el orden observado
  es el de inserción (id ascendente).

### 4.3 Paginación (`Page<T>` de Spring Data)

Shape exacto (medido):
`{"content":[…],"pageable":{"pageNumber":0,"pageSize":2,"sort":{"sorted":false,"unsorted":true,"empty":true},"offset":0,"paged":true,"unpaged":false},"totalElements":81,"totalPages":41,"last":false,"first":true,"size":2,"number":0,"sort":{…},"numberOfElements":2,"empty":false}`

- `page` es 0-based, `size` sin límite.
- `content` contiene **entidades JPA anidadas**, no los DTOs aplanados:
  `Turno` → `{"id","fechaHora","cliente":{"nombre","apellido","email","telefono"},"servicio":{"id","tipo","precio"}}`.

### 4.4 Semántica PATCH (genérico `BeanUtils.copyProperties` sin nulos)

- Copia solo campos no nulos del body sobre la entidad existente.
- `PATCH /turnos/{id}`: **no valida** reglas de negocio (permite hora pasada, minutos
  no múltiplos de 30). `servicio:{"id":N}` se copia como objeto parcial:
  - id válido → la respuesta del PATCH muestra `tipoServicio:null, precioServicio:0.0`
    (objeto sin refrescar); un GET posterior sí muestra los valores reales.
  - id inexistente → **500** (violación FK, sin cascade).
- `PATCH /clientes/{id}`: actualiza campos no nulos (`telefono` es la PK).
- `PATCH /administradores/{id}`: copia campos no nulos **sin hashear `contrasenia`**
  (defecto de seguridad: guardaría contraseña en texto plano; ver §7).

### 4.5 Defectos observados en ejecución

1. `POST /turnos/all` con cliente **nuevo** → **500** (`service.saveAll` salta el
   find-or-create de `TurnoService.save`; FK `cliente_id` inexistente). Funciona (201)
   solo si el cliente ya existe.
2. `GET /administradores/{page}/{size}` expone `contrasenia` (hash Argon2) en `content`.
3. `POST /administradores` + `PATCH` con `contrasenia` la persiste **en texto plano**
   (el override hashea solo en `save()`, no en `update()`).
4. Regla de duplicado de turno no atómica (check-then-insert, carrera entre requests).
5. `PATCH /turnos` permite violar todas las reglas de negocio de creación.
6. `/turnos/` (barra final sin id) → 404 `about:blank` "No static resource turnos."
   (edge case de Spring, no consumido por el frontend).

## 5. Endpoints y permisos (SecurityConfig)

Públicos: `OPTIONS /**`, `GET /ping`, `POST /turnos`, `GET /turnos/findDateTimes`,
`GET /servicios` y `GET /servicios/**`, `POST /administradores/login`,
`/v3/api-docs/**`, `/swagger-ui/**`. Todo lo demás requiere Bearer válido.
`/actuator/health` responde **401** (queda detrás de `anyRequest().authenticated()`).

## 6. Reglas de negocio de turnos

- `POST /turnos`: valida `@Future` + `@NotNull` (fecha), `idServicio` existe (404),
  teléfono/nombre/apellidoNotBlank, email con formato.
- Hora entre 13:00 y 20:00 inclusive; minutos en :00 o :30.
- Duplicado: busca entre turnos `>= fechaHora` pedida; si existe, 400 con mensaje
  `"Ya existe un turno para la fecha y hora especificada. Fecha y hora: <iso>"`.
- Cliente: si el teléfono no existe lo crea; si existe, **reutiliza el registro
  existente** (los datos del request se descartan).
- `GET /turnos/findDateTimes`: turnos `>= ahora` (sin tenant).
- Panel: turnos hoy/ayer (por fecha calendario), ingresos del mes y mes anterior
  (`SUM(s.precio)` de turnos en `[inicio, fin)`), clientes totales, ingresos diarios
  de los últimos ~31 días (bucket de día, hoy primero, incluye hoy parcial),
  servicios del último mes agrupados por servicio ordenado por count desc,
  horarios por hora del día desde el 1 de enero del año en curso.

## 7. Seguridad: hallazgos

| # | Hallazgo | Severidad | Estado en Go |
|---|---|---|---|
| S1 | `contrasenia` (hash) expuesta en `GET /administradores/{page}/{size}` | Alta | Corregida: la página no incluye el hash (campo omitido; forma del objeto idéntica salvo ese campo) |
| S2 | `PATCH /administradores` persiste contraseña en texto plano | Crítica | Corregida: si el body incluye `contrasenia`, se hashea (Argon2id) |
| S3 | Reserva no atómica (doble booking por carrera) | Alta | Corregida: UNIQUE (tenant_id, fecha_hora) + manejo de 23505 |
| S4 | Sin rate limit en login | Media | Corregida: limitador simple por IP (5/min, configurable) |
| S5 | Swagger público fuera de prod (perfil `prod` lo deshabilita) | Baja | En Go no se implementa UI Swagger; el spec es un archivo versionado |
| S6 | Token en `sessionStorage` (frontend) | Media | No cambia (fuera de alcance backend); documentado |
| S7 | CORS dev con credenciales + wildcards de método/headers | Baja | Replicado solo para orígenes localhost:5173/5174; prod solo GitHub Pages |

## 8. Frontend: dependencias del contrato

- Cliente: `fetch` nativo, `VITE_API_BASE_URL` (default `http://localhost:8080`),
  Bearer desde `sessionStorage["auth_token"]`, timeout 15 s, 401 → limpia sesión y
  redirige a `/login` (excepto en el login).
- Operaciones usadas (9): login, `GET /turnos`, `POST /turnos`, `PATCH /turnos/{id}`
  (body parcial `{fechaHora?, servicio?:{id}}`), `DELETE /turnos/{id}`,
  `GET /turnos/findDateTimes`, `GET /turnos/0/{size}` (remapea entidad→DTO),
  `GET /servicios`, `GET /estadisticas/panel`.
- Tipos: `frontend/src/types/api.ts` es generado por `openapi-typescript@7.13.0`
  desde `backend/openapi-spec.json` (`bun run types`).

## 9. Configuración y despliegue actuales

- Env: `DB_URL` (JDBC), `DB_USER`, `DB_PASSWORD`, `JWT_SECRET` (obligatoria),
  `ADMIN_SEED_EMAIL`/`ADMIN_SEED_PASSWORD` (seed si no hay admins), `SPRING_PROFILES_ACTIVE`.
- `prod`: deshabilita springdoc; CORS fijo a `https://mateo-404.github.io` sin credenciales.
- Seed demo solo en perfil `h2`.
