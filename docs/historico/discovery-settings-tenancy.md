# DISCOVERY.md (Fase 0 - solo lectura)

## 1) Módulo, versión Go, estructura, router
- Módulo: `github.com/Mateo-404/Barberia/backend-go` (backend-go/go.mod:1-3)
- Go: `go 1.27.0` (backend-go/go.mod:3)
- Router/framework HTTP: `github.com/go-chi/chi/v5 v5.5.0` (go.mod:8). Router montado en `backend-go/internal/httpapi/server.go:205-305` con chi.NewRouter(), middlewares y rutas bajo `/administradores`, `/clientes`, `/servicios`, `/turnos`, `/estadisticas/panel`.
- Estructura relevante:
  - `backend-go/cmd/api/main.go` (entrypoint)
  - `backend-go/internal/httpapi/` (server, middleware, handlers, respuestas)
  - `backend-go/internal/auth/` (JWT, login limiter)
  - `backend-go/internal/tenant/` (resolver por host)
  - `backend-go/internal/modules/{administradores,clientes,servicios,turnos,estadisticas}/`
  - `backend-go/internal/platform/database/` (pool, migraciones)
  - `backend-go/internal/platform/database/sqlc/` (sqlc generado)
  - `backend-go/db/migrations/` (migraciones existentes en el repo)
  - `backend-go/api/`, `backend-go/public/` (otros directorios)

## 2) Acceso a BD, transacciones, migraciones
- Driver/pool: `github.com/jackc/pgx/v5 v5.11.0` con `pgxpool` (`internal/platform/database/pool.go` vía uso en main). Pool creado con opciones en `cmd/api/main.go:72-82`.
- Transacciones: módulos usan `*pgxpool.Pool`/`*sqlc.Queries` según caso (p.ej. `administradores.New` toma pool+q, `estadisticas.New` toma q). No hay helper `WithTenantTx` visible aún.
- Migraciones: `github.com/golang-migrate/migrate/v4 v4.20.1` + driver postgres (`go.mod:10,28`). Ejecutadas con `database.RunMigrations(ctx, cfg.DatabaseURL, logger)` si `cfg.AutoMigrate` (main:85-89). Numeración actual en `backend-go/db/migrations/`: ver archivos (ej. `000001_*.up.sql`, etc., revisar descubrimiento completo).
- sqlc: `sqlc.yaml` presente, queries en `internal/platform/database/queries/` y código generado en `sqlc/`.

## 3) Auth: JWT, claims, roles, login
- JWT: `github.com/golang-jwt/jwt/v5` (`go.mod:9`). Manager en `internal/auth/jwt.go`. Claims incluyen `sub`, `role` (y código existente parsea). En middleware `tenantDeClaims` (middleware.go:185-199): si `c.TenantID != 0` usa claim `tenant` (tokens nuevos); si no, consulta admin por `sub` para obtener tenant (tokens legados Java).
- Roles existentes: en login/handlers se manejan roles (`administradores`), migraciones/queries muestran roles. En migraciones de referencia del kit aparecen `owner/admin/staff` para `users`.
- Login: ruta `POST /administradores/login` expuesta en router (server.go:267-276) y marcada como pública en `esPublica` (middleware.go:97-120). Handler `hAdminLogin` en `internal/httpapi/handlers_administradores.go` (llama a servicio). Servicio `administradores.Service` crea token vía JWT manager.

## 4) Esquema actual vs migraciones de referencia (tabla de mapeo)
**Migraciones existentes en repo (backend-go/db/migrations/):** revisar archivos concretos para mapeo preciso.
**Migraciones de referencia (kit):** `0001_tenancy.up.sql` crea `tenants`, `tenant_domains`, `users` con RLS y función `enable_tenant_rls`; `0002_settings.up.sql`, `0003_bookings.up.sql`.

**Hallazgos iniciales del esquema actual (desde código):**
- Existe lógica multi-tenant parcial por host: `tenant.Resolver` consulta por host/default slug vía `sqlc` (`GetTenantByHost`, `GetTenantBySlug`). Main resuelve default para seed (main:104-129). Middleware resuelve tenant por host cuando no hay token, o desde claims cuando hay token (middleware.go:162-183).
- No se ve uso generalizado de `set_config('app.tenant_id', ...)` ni helper `WithTenantTx` aún.
- Roles/claims JWT: hay soporte para claim `tenant` en middleware (tenantDeClaims) pero implementación completa según SPEC puede faltar.
- Tablas/columnas reales hay que confirmar leyendo migraciones sqlc y esquema actual. **Requiere lectura detallada de `backend-go/db/migrations/*.sql` y `internal/platform/database/sqlc/*`**.

## 5) Cómo se sirve el front hoy
- **No hay embed evidente en main** leyendo cmd/api/main.go: no se ve `embed`, `http.FileServer`, ni servir archivos estáticos. Servidor solo monta router API (srv.Handler()). 
- En repo raíz hay muchos archivos estáticos HTML/JS/CSS (`index.html`, `admin/`, `js/`, `css/`) y `frontend/` (React/Vite según AGENTS.md raíz). 
- En backend-go hay `public/` y `api/` dirs; revisar si se sirve desde ahí.
- **Según SPEC §11 default:** "Go sirve `frontend/` (embed o `http.FileServer`) y renderiza solo la página pública con plantilla". Necesita confirmación inspeccionando httpapi/server.go completo y búsqueda de FileServer/embed.

## 6) Config/env, tests existentes, CI, make verify
- Config: `internal/config/config.go`, `.env.example` en backend-go. Variables típicas DB/JWT/CORS/tenancy.
- Tests existentes: buscar `*_test.go` en backend-go. También hay tests SQL en kit (`db/tests/*.sql`) y kit tiene Makefile.
- CI: buscar `.github/workflows/` en repo.
- `make verify` en backend-go/Makefile (leerlo). Kit también tiene Makefile. **Evaluar viabilidad**: puede requerir DB, herramientas (sqlc, migrate), etc.

## 7) Riesgos / contradicciones con SPEC
- SPEC exige: RLS con `app_user`, `set_config('app.tenant_id', ...)` por transacción vía helper `WithTenantTx(ctx,fn)`, claim `tenant_id` en JWT, middleware según orden §4 (Host exacto en tenant_domains verificado, subdominio BASE_DOMAIN, header X-Tenant-Slug solo dev), 404 genérico para tenant desconocido/suspendido, JWT tenant != host → 403, login dentro del tenant resuelto.
- Estado actual: resolver existe por host/default; auth middleware tiene lógica para tenant desde claims o host; pero falta implementación completa de políticas RLS, helper WithTenantTx, claim tenant_id obligatorio según nuevo esquema, etc. Coincide parcialmente con dirección pero hay que adaptar al esquema real.

## Actualización detallada (hallazgos clave)

### Esquema real vs kit
- **Tenants:** tabla `tenants` con `BIGSERIAL id`, `slug`, `name`, `host`, `created_at`. No tiene `plan/status` como en kit; tiene `host`. Resolver usa `GetTenantByHost`/`GetTenantBySlug` sobre este esquema.
- **Tenant_id:** `BIGINT` (no UUID). Agregado por migración 000002 con DEFAULT `default_tenant()` y backfill.
- **PK compuesta clientes:** `(tenant_id, telefono)`. FK compuesta `turnos(tenant_id, cliente_id) -> clientes(tenant_id, telefono)`.
- **RLS:** **NO implementado** aún en migraciones actuales. No hay `ENABLE/FORCE ROW LEVEL SECURITY`, no hay función `enable_tenant_rls`, no hay políticas. No hay rol `app_user` creado/forzado.
- **Settings/theme:** **NO existe** `tenant_settings`. 
- **Bookings/range:** tabla `turnos` tiene `fecha_hora TIMESTAMP` (sin `tstzrange`/`EXCLUDE`). Unicidad por `(tenant_id, fecha_hora)` vía índice único (no constraint EXCLUDE).
- **Usuarios/auth tabla:** existe `administradores` (con `contrasenia`), no tabla `users` como en kit. JWT claims/legado ya manejan tenant lookup por admin id.
- **Tenant domains:** **NO existe** tabla `tenant_domains`. Resolución actual solo por `tenants.host` o default slug.

### DB/tx/migraciones
- Migraciones con `golang-migrate`. Numeración `000001_*`, `000002_*`. Necesitará adaptar migraciones del kit a este esquema (BIGINT, tablas existentes, naming). También crear `.down.sql` para nuevas migraciones.
- Sin `WithTenantTx` helper. Sin `set_config('app.tenant_id', ...)` uso.
- Pool `pgxpool` via `database.NewPool`.

### Auth/JWT
- Claims: estructura en `internal/auth/claims.go` (ver). Middleware soporta `c.TenantID` (claim `tenant`) o lookup legado por admin. Hay que asegurar claim `tenant_id` según SPEC.
- Login en `/administradores/login`, servicio usa JWT manager.

### Frontend serving
- **No encontrado** en backend-go código para servir archivos estáticos/embed. Revisar `internal/httpapi` completo y `public/` dir. Según SPEC default puede servirse `frontend/` - pero frontend/ está en root del repo, no en backend-go. Necesita investigación adicional.

### Config/tests/CI/verify
- `backend-go/Makefile` tiene `test`, `test-integration` (tag integration), `lint` (golangci-lint), etc. No veo target `verify` explícito. Kit pide `make verify` - puede no existir tal cual; habrá que ajustar según hallazgos.
- Tests Go: existen `*_test.go` en varios paquetes. Tests SQL de kit (`db/tests/*.sql`) no están en backend-go - están en `/home/mateo/Descargas/barberia-saas-kit/db/tests/`. Kit dice "tests SQL ya escritos (RLS y reservas). Son la vara; no los debilites" - habrá que integrarlos/ejecutarlos según lo que pida SPEC/ACCEPTANCE.
- CI: revisar `.github/workflows/` en repo raíz.

### Riesgos/contradicciones con SPEC
1. **Esquema incompatible** con kit (UUID vs BIGINT, `administradores` vs `users`, sin `tenant_domains`, sin `tenant_settings`, sin RLS). Hay que mapear y adaptar migraciones, NO duplicar.
2. **RLS/roles**: actualmente sin RLS; pasar a `app_user` con FORCE RLS requiere crear rol, migrar conexión a usarlo, asegurar ninguna ruta use dueño/superuser.
3. **Resolución tenant**: SPEC pide orden Host exacto en `tenant_domains` (verificado), subdominio BASE_DOMAIN, header X-Tenant-Slug solo dev, 404 genérico si suspendido/desconocido, JWT tenant != host → 403. Actual es más simple (host en tenants.host o default). Hay que extender.
4. **WithTenantTx helper**: requerido por SPEC, no existe.
5. **Settings API/theme CSS inline**: no existe. Frontend serving/templating para inyectar CSS inline no visible.
6. **Reservas**: sin `tstzrange` EXCLUDE, usa índice único `(tenant_id, fecha_hora)`. SPEC pide migrar a `during tstzrange` con EXCLUDE si no usa eso. Hay que evaluar.
7. **Migraciones kit usan UUID** - incompatible directamente con BIGINT actual. Adaptación crítica.

## Tabla de mapeo completa (esquema real vs kit)
| Concepto | Esquema actual (backend-go) | Kit referencia | Acción |
|---|---|---|---|
| tenants | `BIGSERIAL id`, `slug` UNIQUE, `name`, `host` UNIQUE, `created_at` | UUID id, slug/name/plan/status | **Mantener BIGINT**. Adaptar lógica a estructura existente (no migrar a UUID). Añadir campos si necesarios por SPEC. |
| tenant_domains | **NO existe** | existe (domain PK, tenant_id UUID, verified_at) | Crear si necesario para resolución §4; adaptar a BIGINT. |
| users/administradores | `administradores` (BIGSERIAL, tenant_id BIGINT, nombre/apellido/email/contrasenia) con unique (tenant_id, lower(email)) | `users` (UUID) con roles owner/admin/staff | Usar `administradores`. Mantener unicidad por tenant. Adaptar migraciones/tests a `administradores`. |
| clientes | `(tenant_id, telefono)` PK compuesta, FK compuesta desde turnos | UUID users style | Mantener estructura compuesta. |
| servicios | `BIGSERIAL id`, `tenant_id`, `tipo`, `precio` | referencia diferente | Mantener. Añadir tenant_id/RLS si falta. |
| turnos | `BIGSERIAL id`, `(tenant_id, fecha_hora)` unique index, FK compuestas | puede usar tstzrange EXCLUDE | Mantener estructura; Fase 3 evaluar EXCLUDE con tstzrange. |
| tenant_settings | **NO existe** | existe (draft/published/versioning) | Crear en migración 000003+ para Fase 2. |
| RLS/app_user | **NO existe** | app_user + FORCE RLS + políticas | Crear rol app_user sin SUPERUSER/BYPASSRLS, grants tablas+secuencias+default privs, FORCE RLS, función current_tenant_id() BIGINT. |
| WithTenantTx | **NO existe** | requerido | Implementar con set_config app.tenant_id (BIGINT) desde context. |
| JWT claim tenant | claim name `tenant`, valor **int64/BIGINT** | puede variar | **Mantener nombre `tenant`**. Tokens sin claim → **401 (re-login), sin fallback**. |
