# Auditoría de seguridad — fixes aplicados y decisiones

Fecha: 2026-10-21. Alcance: backend Go (`backend-go/`), frontend React
(`frontend/`) y dependencias. El sitio legacy vanilla (`index.html`, `js/`,
`admin/`) se audita pero **no se toca** por decisión del usuario (queda como
deuda pendiente documentada).

## Resumen

| Severidad | Encontrados | Aplicados | No aplicados (justificados) |
|---|---|---|---|
| Crítico | 0 | 0 | 0 |
| Alto | 1 | 1 | 0 |
| Medio | 8 | 8 | 0 |
| Bajo | 5 | 4 | 1 |

Backend verificado: `go build`, `go test -race ./...` (8 paquetes),
`go test -tags integration` (17+ tests), `gofmt`, `golangci-lint` (0 issues),
`govulncheck` (0 vulnerabilidades).

Frontend verificado: `bun run test` (20 tests), `tsc -b`, `bun run build`.

## Hallazgos aplicados

### A1 (Alto) — Data race en `tenant.Resolver`

El `map` de caché se escribía desde goroutines concurrentes del servidor HTTP:
Mapa sin sincronización ⇒ "concurrent map writes" y crash del proceso.

**Fix:** `sync.RWMutex` (lectura rápida con `RLock`, escritura con `Lock`).
Test de regresión: `internal/tenant/resolver_race_test.go` con `-race`.

### M6 (Medio) — Consulta por Host en cada request desconocido

Un atacante variando el `Host` forzaba una consulta a `tenants` por request (ni
siquiera se cacheaba el fallo). **Fix:** caché de "misses" con set de Hosts ya
observados como desconocidos: tras el primer fallo, se sirve el tenant por
defecto sin volver a consultar.

> Durante este fix se detectó y corrigió un bug propio: el caché de misses se
> había comido la consulta `GetTenantByHost`, degradando el aislamiento. Se
> restauró la consulta y se añadió `TestAislamientoEscriturasEntreTenants` +
> el test de aislamiento por Host reforzado.

### M1 (Medio) — Cabeceras de seguridad y caché

Faltaban cabeceras defensivas y las respuestas sensibles no evitaban cachés
intermedias.

**Fix:** `X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY`,
`Referrer-Policy: no-referrer`, `Cross-Origin-Resource-Policy: same-site`,
CSP `default-src 'none'; frame-ancestors 'none'`, HSTS si la conexión es TLS, y
`Cache-Control: no-store` en login y respuestas autenticadas.

### M2 (Medio) — Validación en PATCH y topes de longitud

Java solo validaba `POST` (`@Valid`) y el PATCH aceptaba cualquier cosa; además
no había topes de longitud en campos de texto.

**Fix:** validación de los campos *presentes* en PATCH (`email`, `contrasenia
≥ 8`, `precio > 0`, `nombre`/`apellido`/`tipo` no vacíos) y topes de longitud
(`maxNombre`, `maxEmail`, `maxTelefono`, `maxTipoServicio`) en create y PATCH.
Desviación documentada en la matriz de compatibilidad.

### M3 (Medio) — Tope de paginación y overflow

`size` sin límite permitía amplificar el trabajo por request; `page*size` podía
desbordar `int32`. **Fix:** `size > 100` → 500; fuera de int32 → 400;
`page < 0 || size < 1` mantiene el 500 de `PageRequest.of`.

### M4 (Medio) — `MAX_REQUEST_BODY` no se aplicaba

La variable de entorno era config muerta: el handler usaba una constante fija.

**Fix:** `decodificarJSONConLimite` lee `cfg.MaxRequestBody` en cada handler
(manteniendo 1 MiB como default).

### M5 (Medio) — Sin límite en escrituras públicas

`POST /turnos` es público y no tenía control de tasa: cualquiera podía crear
clientes y turnos masivamente.

**Fix:** rate limit por IP (`TURNO_RATE_LIMIT_PER_MINUTE`, default 30) →
429. Se aplica antes de tocar la BD.

### M7 (Medio) — Timeouts de servidor y pool

Sin `Read/Write/IdleTimeout` ni `statement_timeout`/`AcquireTimeout`, una
conexión lenta o una consulta colgada podían agotar recursos.

**Fix:** timeouts en `http.Server` y en el pool pgx (`statement_timeout`
configurable vía `STATEMENT_TIMEOUT`).

### M8 (Medio) — PATCH de turno podía enlazar un servicio ajeno

La FK de Postgres valida el `id` pero no el tenant, así que un tenant podía
apuntar su turno a un servicio de otro tenant y filtrar su tipo/precio.

**Fix:** validación de pertenencia al tenant en `PATCH /turnos`. Ahora un id
inexistente o ajeno → 404 (antes 500 por FK en el caso inexistente).

### B3 (Bajo) — `DATABASE_URL` en el log de arranque

`pgxpool.ParseConfig` puede incluir la cadena (con contraseña) en su error, y
ese error se registraba. **Fix:** mensaje genérico sin la cadena.

### B4 (Bajo) — Ruta reflejada en el `detail` del 404

El `detail` replica el comportamiento de Spring (`No static resource ...`) e
incluye la ruta. **Fix/mitigación:** `nosniff` + CSP `default-src 'none'`, con
`Content-Type: application/problem+json`, hacen inviable el XSS. Se mantiene la
paridad con Spring.

### B5 (Bajo) — Rate limit detrás de proxy

Sin proxy, todos los clientes comparten IP; con proxy, `RemoteAddr` es la del
proxy. **Fix:** `TRUST_PROXY_HEADERS` (opt-in, default `false`) usa el primer
`X-Forwarded-For`. Requiere un proxy que **sobrescriba** la cabecera; no exponer
el backend directo con la opción activa.

### I2 (Informativo, aplicado) — Panic sin traza

`middlewareRecover` ahora registra `debug.Stack()`.

## Hallazgos NO aplicados

### I4 (Informativo) — `Vary: Origin`

La respuesta solo varía por `Origin` cuando el origen es permitido, y en ese
caso ya se añade `Vary: Origin`. Se dejó como está por no aportar beneficio real
en este caso.

### `braces@3.0.3` (CVE-2026-93687, GHSA-vfj7-8cjw-p6xm) — **sin parche upstream**

`bun audit` reporta `braces@3.0.3` con un stack overflow por patrones anidados
(DoS, severidad Alta). Investigado:

- **No existe versión parcheada** ("Patched versions: None" en el advisory; la
  última publicada es 3.0.3). No hay `bun audit fix` posible.
- `braces` llega de forma **transitiva** por `shadcn` → `@shadcn/registry` →
  `fast-glob` → `micromatch` → `braces`.
- `shadcn` es una **CLI de desarrollo** (genera componentes); **no aparece en el
  bundle** desplegado (verificado: 0 referencias en `dist/assets/*.js`) ni se
  ejecuta en runtime. El vector de ataque exigiría alimentar `micromatch` con un
  patrón controlado por un atacante, cosa que no ocurre en build ni en runtime.

**Decisión (del usuario): documentar y no tocar.** Sin fix upstream no hay acción
útil; se vigila el advisory. No se regeneró `bun.lock`.

## Deuda pendiente — sitio legacy (`index.html`, `js/`, `admin/`)

Por decisión del usuario **no se toca**. Hallazgos a resolver cuando se retire
o se endurezca:

- **C1 (Crítico) — XSS almacenado:** `js/panel-turnos.js` y
  `js/panel-estadisticas.js` escriben datos del servidor (nombre, teléfono,
  email de clientes) con `innerHTML`. Debe pasar a `textContent`/DOM seguro.
- **Console con PII:** `console.log` con datos de clientes en los paneles.
- **SRI:** los `<script>`/`<link>` de CDN (Bootstrap) no usan `integrity`/`crossorigin`.
- **Login legacy sin JWT:** `admin/login.js` mantiene un flujo sin token Bearer.
- **CSP:** `index.html` no declara las cabeceras de seguridad del backend.

## Checklist de despliegue (CSP del frontend)

`frontend/index.html` declara `connect-src 'self' __CSP_CONNECT_SRC__`; el
placeholder lo sustituye `vite.config.ts` en build-time por el **origen de
`VITE_API_BASE_URL`** (la misma variable que usa el cliente HTTP), más
`ws://localhost:5173` en dev y los orígenes extra de `VITE_CSP_CONNECT_SRC`.

> Actualización 2026-10-23: antes el default de producción era
> `https://mateo-404.github.io`, lo que bloqueaba cualquier API en otro origen
> (el e2e de Playwright lo detectó). Para producción basta con
> `VITE_API_BASE_URL=https://api.<dominio> bun run build`.

## Frontend React — fixes aplicados

- **`.max()` en los schemas Zod** (`login-schema.ts`, `turno-schema.ts`):
  alineados con los topes del backend (email 254, contraseña 128, nombre/
  apellido 120, teléfono 32).
- **`exp` obligatorio** en `AuthContext.isTokenExpired`: un token sin `exp` se
  trata como expirado (antes se aceptaba ⇒ bypass). El backend exige expiración.
- **Errores 5xx genéricos** en `api-client.ts`: no se muestra el `detail`
  interno del servidor (podía filtrar trazas/SQL); se conserva en 4xx.
- **`Authorization` solo en rutas protegidas**: no se envía el Bearer a `/ping`,
  `/healthz`, `/readyz`, `/servicios`, `/turnos/findDateTimes`,
  `/administradores/login` ni `POST /turnos`.
- **`EstadisticasAdmin`**: estado de error con `role="alert"` y botón
  "Reintentar" (antes mostraba ceros como si fueran datos reales).
- **Tests nuevos:** `lib/__tests__/api-client.test.ts` (Bearer por ruta, 5xx sin
  detalle) y `lib/__tests__/auth-token.test.ts` (política de `exp`).
- **`@hookform/resolvers` 5.9.2 → 6.0.0:** major de bajo riesgo. Respecto a
  5.9.2, los peer deps solo cambian `effect` (`^3` → `^4`, no usado); `zod`
  (`^3.25.0 || ^4.0.0`) y `react-hook-form` (`^7.55.0`) se mantienen. El único
  uso es `zodResolver(...)` en `Reserva.tsx` y `Login.tsx`; typecheck, lint,
  tests (20) y build OK tras la subida.
