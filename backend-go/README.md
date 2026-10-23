# Barbería — Backend Go

Backend activo de la barbería: reescritura en Go del antiguo backend Spring
Boot, compatible con el mismo contrato de API que consume el frontend React.
El backend Java fue retirado de `main` (historial en `docs/migracion-go/` y en
la rama `backend`).

## Stack

| Pieza | Versión | Uso |
|---|---|---|
| Go | 1.27 (toolchain go1.27.3 fijado en `go.mod`) | lenguaje |
| chi | v5.5.0 | router HTTP |
| pgx/v5 | v5.11.0 | driver PostgreSQL + pool |
| sqlc | v1.31.1 | acceso a BD tipado (código generado en `internal/platform/database/sqlc/`) |
| golang-migrate | v4.20.1 | migraciones (`db/migrations/`, embebidas en el binario) |
| golang-jwt/v5 | v5.3.1 | JWT HS256 |
| x/crypto | v0.53.0 | Argon2id |

Decisiones: `docs/migracion-go/adr/ADR-001..006`. Contrato y desviaciones:
`docs/migracion-go/02-matriz-compatibilidad.md`.

## Arranque rápido

```bash
# 1. PostgreSQL + API con Docker (desde la raíz del repo):
JWT_SECRET=$(openssl rand -base64 32) docker compose up --build

# 2. O local (PostgreSQL propio o `make up` desde la raíz):
cp .env.example .env        # completar JWT_SECRET (obligatoria, ≥32 bytes)
set -a; source .env; set +a
make run                    # compila y arranca en :8080
```

Seed: si el tenant por defecto no tiene administradores se crea
`ADMIN_SEED_EMAIL` con `ADMIN_SEED_PASSWORD` (hasheada con Argon2id).

## Estructura

```
backend-go/
├── api/openapi.json            # contrato del API (fuente de los tipos TS del frontend)
├── cmd/api/main.go             # arranque, chequeo de rol, seed, shutdown graceful
├── cmd/migrate/main.go         # aplica migraciones con MIGRATE_DATABASE_URL (rol dueño)
├── testdata/                   # fixtures de settings y vectores de contraste
├── db/
│   ├── migrations/             # 000001 baseline … 000009 (tenancy, RLS, settings, roles, FKs, inglés, theme v2, auth Google/TOTP)
│   ├── queries/                # queries sqlc (todas tenant-scoped)
│   └── embed.go
└── internal/
    ├── apperr/                 # errores de dominio → ProblemDetail
    ├── auth/                   # Argon2id (compatible Spring), JWT HS256 (con scope 2FA), rate limit, TOTP, Google OAuth
    ├── config/                 # env con validación fail-closed
    ├── httpapi/                # router, middleware, handlers, shapes de error
    ├── modules/                # turnos, clientes, servicios, administradores, estadisticas
    ├── pgutil/                 # clasificación SQLSTATE
    ├── platform/database/      # pool, migrador, WithTenantTx (RLS) + código sqlc
    ├── settings/               # validación, contraste WCAG y CSS del branding por tenant
    ├── tenant/                 # resolución Host → tenants.host (fallback default)
    └── timeutil/               # semántica LocalDateTime de Java (formatos, meses)
```

## Comandos

```bash
make test               # unitarias + HTTP con -race (sin BD)
make test-integration   # end-to-end contra PostgreSQL
make lint               # golangci-lint v2.15.0 (0 issues)
make vuln               # govulncheck (0 vulnerabilidades alcanzables)
make migrate-up         # aplica migraciones (MIGRATE_DATABASE_URL o DATABASE_URL)
make sqlc               # regenerar código de BD tras editar db/queries
```

## Tests de integración

`internal/integration/` (build tag `integration`) cubre: migración fresca,
adopción de una base creada por Hibernate, flujo HTTP completo (login Argon2
con hash real de Spring, reglas de turno, PATCH sin refrescar, panel),
concurrencia (doble reserva → exactamente un 201), aislamiento de tenant,
roles owner/staff y publicación/caché de settings. Los tests SQL de RLS viven
en `db/tests/` (`make test-db` desde la raíz).

Requiere un PostgreSQL accesible. Default:
`postgres://postgres@127.0.0.1:55432/postgres?sslmode=disable`
(sobrescribir con `TEST_DATABASE_URL_BASE`; con `make up`:
`postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable`).
Cluster descartable sin Docker:

```bash
PG=/usr/lib/postgresql/18/bin; D=$(mktemp -d)
$PG/initdb -D $D/data -U postgres -A trust
$PG/pg_ctl -D $D/data -o "-p 55432 -k '' -c listen_addresses=127.0.0.1" -w start
```

## Multi-tenancy (ADR-002)

- Columna `tenant_id` en todas las tablas de negocio; queries sqlc siempre
  filtran por tenant.
- Rutas **públicas** (p. ej. `POST /turnos`, `GET /servicios`): el tenant sale
  del `Host` contra `tenants.host`; host desconocido → `DEFAULT_TENANT_SLUG`.
- Rutas **autenticadas**: el tenant sale del claim `tenant` del JWT. Tokens sin
  ese claim (emitidos por el backend Java) → 401 y re-login.
- El tenant nunca se toma de una cabecera enviada por el cliente (excepción:
  `X-Tenant-Slug` solo con `APP_ENV=dev`).
- Defensa en profundidad: políticas RLS (`000003`) con `app.tenant_id` seteado
  por `database.WithTenantTx`; la app debe conectarse con un rol sin
  SUPERUSER/BYPASSRLS (ver `docs/DEPLOY_NOTES.md`).

## Seguridad

- `JWT_SECRET` obligatoria y ≥32 bytes (arranque bloqueado si no).
- `APP_ENV` default `prod`: el arranque aborta si la conexión usa un
  superusuario o un rol con BYPASSRLS. Para desarrollo local, `APP_ENV=dev`.
- Contraseñas Argon2id `m=19456,t=2,p=1` — verifica hashes existentes de Spring
  sin resetear contraseñas; los nuevos se generan igual.
- `GET /administradores/{page}/{size}` **no** expone el hash (el Java sí).
- `PATCH /administradores/{id}` hashea `contrasenia` (el Java guardaba texto plano).
- Rate limit de login por IP (`LOGIN_RATE_LIMIT_PER_MINUTE`, por proceso).
- Rate limit de reservas públicas `POST /turnos` (`TURNO_RATE_LIMIT_PER_MINUTE`).
- CORS por origen exacto configurable (`CORS_ALLOW_ORIGINS`).
- Cabeceras de seguridad (`nosniff`, `X-Frame-Options`, `Referrer-Policy`,
  `Cross-Origin-Resource-Policy`, HSTS si TLS) y `Cache-Control: no-store` en
  login y respuestas autenticadas.
- `TRUST_PROXY_HEADERS=true` usa `X-Forwarded-For` para el rate limit (solo
  detrás de un proxy de confianza que sobrescriba la cabecera).
- Timeouts de servidor y de pool (`STATEMENT_TIMEOUT`).
- Bodies limitados a `MAX_REQUEST_BODY` (default 1 MiB).
- Auditoría y hallazgos: `docs/migracion-go/04-auditoria-seguridad.md`.

## Docker

`Dockerfile` multi-stage (build con `golang:1.27.3-bookworm`, runtime
distroless nonroot, sin CGO). El CI construye la imagen en cada cambio.

## Autenticación

- `POST /api/auth/login` — email + password. Devuelve `{ token, ...admin }` o
  `{ requires2fa: true, tempToken }` si el admin tiene TOTP activo.
- `GET /api/auth/providers` — capacidades disponibles (`{ google: bool }`).
- `POST /api/auth/google/start` / `POST /api/auth/google/callback` — login con Google.
- `POST /api/auth/totp/challenge` — completa el desafío 2FA con `tempToken` + código.
- `POST /api/auth/totp/setup|verify|disable` — gestión del segundo factor (autenticado).

Roles: `owner`, `admin`, `staff`. Los tokens incluyen el claim `tenant` (BIGINT)
y `rol`. Los tokens sin `tenant` se rechazan con 401.
