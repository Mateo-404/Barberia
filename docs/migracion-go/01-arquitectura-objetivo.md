# 01 — Arquitectura objetivo (Go)

## 1. Decisiones de stack (versiones verificadas)

| Herramienta | Versión elegida | Fuente oficial verificada | Motivo |
|---|---|---|---|
| Go | **1.27.3** (stable, 2026-10) | https://go.dev/dl/?mode=json (`go1.27.3`, `stable: true`) | Último release estable; la local (1.22.2) no alcanza a las dependencias actuales (pgx ≥1.25, golangci-lint v2). Instalado en `~/.local/share/go-toolchains/go1.27.3` (aislado, sin tocar el sistema). |
| chi | **v5.5.0** | github.com/go-chi/chi latest release | Routing/middleware sobre `net/http`, minimal, sin runtime reflectivo. Evaluado contra `net/http` puro (Go 1.22+ tiene routing): chi aporta middleware chain, URL params y wildcard de ruta consistentes con menor ruido. |
| pgx | **v5.11.0** (pgx + pgxpool) | github.com/jackc/pgx latest release | Driver nativo PostgreSQL, soporte COPY, listen/notify, tipado. |
| sqlc | **v1.31.1** | github.com/sqlc-dev/sqlc latest release | Código Go tipado generado desde SQL explícito: sin ORM, consultas auditables. |
| golang-migrate | **v4.20.1** | github.com/golang-migrate/migrate latest release | Migraciones SQL versionadas con driver postgres; adoptado vía librería (embebida) + CLI en scripts. |
| oapi-codegen | **v2.8.0** | github.com/oapi-codegen/oapi-codegen latest release | Genera tipos Go + `ServerInterface` (chi) desde `api/openapi.yaml`: el spec es la fuente de verdad y el compilador detecta desvíos. |
| golangci-lint | **v2.15.0** | github.com/golangci/golangci-lint latest release | Linter meta (govet, errcheck, staticcheck…). |
| golang-jwt | **v5.3.1** | github.com/golang-jwt/jwt latest release | JWT estándar, `SigningMethodHS256` estricto. |
| Go Cryptography / argon2 | **golang.org/x/crypto** (versión fijada en go.sum vía el toolchain) | pkg.go.dev/golang.org/x/crypto | `argon2.IDKey` — compatible con el formato `$argon2id$v=19$m=19456,t=2,p=1$…` de Spring Security. |
| govulncheck | **x/vuln v1.9.0** (CLI instalada) | `go install golang.org/x/vuln/cmd/govulncheck` | Análisis de vulnerabilidades. |

Todas las versiones son releases estables (no RC/beta). Fijadas en `go.mod`
(vía `tool` directive y require), `go.sum`, `backend-go/tools.go` y CI.

## 2. Estructura (`backend-go/`)

```
backend-go/
├── cmd/api/main.go            # composición: config → db → migraciones → router → server
├── api/openapi.yaml           # contrato OpenAPI 3.1 — fuente única de verdad
├── db/
│   ├── migrations/            # SQL versionado (golang-migrate)
│   └── queries/               # SQL para sqlc
├── internal/
│   ├── config/                # carga y validación de env (fail-closed)
│   ├── httpapi/               # router chi, middleware, problem.go (ProblemDetail)
│   ├── auth/                  # JWT (HS256 estricto), Argon2id, middleware de auth
│   ├── tenant/                # resolución y ámbito de tenant
│   ├── clientes/              # handler + servicio del módulo
│   ├── turnos/
│   ├── servicios/
│   ├── administradores/
│   ├── estadisticas/
│   └── platform/
│       ├── database/          # pgxpool, migraciones embebidas
│       └── store/sqlc/        # código generado por sqlc (no editar a mano)
├── internal/integration/      # tests contra PostgreSQL real
├── Dockerfile                 # multi-stage, non-root
└── sqlc.yaml / .golangci.yml
```

Principios: `main` pequeño y explícito; dependencias inyectadas por constructor;
interfaces solo donde hay una necesidad real (p. ej. `store` como interfaz para
unitarias HTTP); sin capas "por costumbre". Los handlers no contienen SQL: toda
persistencia pasa por `internal/platform/store/sqlc` con SQL explícito.

## 3. Contrato OpenAPI

- `api/openapi.yaml` (OpenAPI 3.1) es la **única fuente de verdad**.
  `oapi-codegen` genera `internal/httpapi/openapi.gen.go` (tipos + `ServerInterface`);
  los handlers implementan la interfaz → cualquier desvío de rutas/firmas rompe la
  compilación.
- El frontend sigue generando sus tipos con `openapi-typescript` desde el mismo
  archivo (`bun run types` apuntará a `../backend-go/api/openapi.yaml`).
- El viejo `backend/openapi-spec.json` queda como snapshot histórico del backend Java
  hasta la Etapa E (rollback); luego se elimina o marca deprecado.
- Diferencias deliberadas respecto del spec Java (documentadas en
  `02-matriz-compatibilidad.md`): `401` documentado en todas las rutas protegidas,
  corrección de `LoginResponseDTO` (campos requeridos), omisión del hash de
  `contrasenia`, y `PATCH /administradores` que hashea la contraseña.

## 4. HTTP y runtime

- `net/http` + chi. `http.Server` con `ReadHeaderTimeout`, `ReadTimeout`,
  `WriteTimeout`, `IdleTimeout` explícitos y `MaxHeaderBytes`.
- `context.Context` propagado a pgx en todas las operaciones.
- Límite de tamaño de body (1 MiB, configurable) y `http.MaxBytesReader`.
- Logging estructurado con `log/slog` (JSON en prod, texto en dev). Nunca se
  registran tokens, contraseñas ni hashes.
- Manejo centralizado de errores: `problem.go` emite RFC 9457 `application/problem+json`
  replicando los `type` y `detail` del backend Java (ver matriz).
- Recuperación de panics (`recover` middleware) → 500 sin filtrar stack.
- Graceful shutdown con signal + `Server.Shutdown` + cierre del pool.
- `GET /healthz` (liveness, sin DB) y `GET /readyz` (ping a DB) — separados de
  `/ping` que replica el contrato Java.

## 5. Persistencia

- `pgxpool` con `MaxConns`, `MinConns`, `MaxConnLifetime`, `HealthCheckPeriod` y
  `ConnConfig.RuntimeParams` (`statement_timeout` configurable) para backpressure.
- Esquema explícito vía migraciones; **sin** `ddl-auto`.
- sqlc para todas las consultas. Transacciones donde el dominio lo requiere
  (creación de turno: cliente + turno; batch endpoints).
- Invariantes en la DB:
  - `turnos`: `UNIQUE (tenant_id, fecha_hora)` → cierra la carrera de doble reserva
    (S3). Ante duplicado concurrente se traduce al mismo 400 del contrato Java.
  - FKs con `ON DELETE RESTRICT` (equivalente al comportamiento actual: borrar un
    cliente/servicio referenciado falla; en Java era un 500, en Go también).
  - `administradores.email` → `UNIQUE (tenant_id, lower(email))` para que el 409
    del contrato sea atómico.
- Fechas: `timestamp` (sin zona) en DB, igual que Hibernate `LocalDateTime`.
  Serialización JSON idéntica a Jackson (`2006-01-02T15:04:05`, con fracción
  solo si existe). `findDateTimes` replica `LocalDateTime.toString()`
  (`2006-01-02T15:04` sin segundos).

## 6. Multi-tenancy (fase 1 de SaaS)

- Modelo compartido con `tenant_id` (FK a `tenants`), **no** RLS ni esquemas por
  cliente (ver ADR-002).
- Ámbito autenticado: el JWT lleva claim `tenant` (id numérico); las queries
  siempre filtran por `tenant_id` del claim. El claim no llega del cliente:
  se emite en el login desde la fila del admin.
- Rutas públicas (reserva): el tenant se resuelve por `Host` → `tenants.host`
  (mapeo controlado por DNS/hosts, no por header de cliente). Fallback
  configurable (`DEFAULT_TENANT_SLUG`) para despliegue single-tenant y compatibilidad.
- Datos existentes: migración que crea el tenant por defecto y asigna todos los
  registros (`000002_tenant_backfill`). Ver `03-plan-migracion.md`.
- Pruebas dedicadas de aislamiento (crear tenant B, verificar que no ve ni
  modifica datos de A por ninguna ruta).

## 7. Seguridad

- Argon2id con parámetros de Spring Security (`m=19456 KiB, t=2, p=1, salt=16B,
  hash=32B`) y **verificación automática del prefijo** para poder migrar
  parámetros futuros sin romper cuentas (ver ADR-003). Login existente en Java
  sigue funcionando sin resetear contraseñas.
- JWT: solo `HS256` aceptado (`WithValidMethods`), `sub` + `email` + `tenant`,
  expiración 24 h, secret ≥ 32 bytes obligatorio (fail-closed al arrancar).
- Rate limit en login por IP (en memoria, 5 intentos/min, configurable).
- CORS: dev limitado a `localhost:5173/5174` con credenciales; prod
  `https://mateo-404.github.io` sin credenciales (idéntico al Java).
- Sin endpoints de diagnóstico expuestos: sin swagger runtime, sin pprof,
  sin actuator. `golangci-lint`, `govulncheck` y tests en CI.

## 8. Testing

- Unitarias: reglas de negocio (turnos), Argon2, JWT, config.
- HTTP (`httptest`): contrato completo de cada endpoint (status, shape JSON,
  headers, mensajes de error) contra una interfaz de store simulada.
- Integración (`internal/integration`, build tag `integration`): PostgreSQL real
  (Docker o cluster local), migraciones completas, queries sqlc, constraints,
  concurrencia (doble POST al mismo slot), aislamiento de tenant.
- Contrato: los handlers implementan `ServerInterface` (compile-time) + test de
  regeneración de tipos TS.
