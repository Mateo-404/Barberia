# Barbería — AGENTS.md

## Architecture

Monorepo: **backend Go** + **frontend React** + SQL de base de datos.

| Directory | Tech | Entrypoint |
|---|---|---|
| `backend-go/` | Go 1.27, chi, pgx/v5, sqlc, golang-migrate, PostgreSQL (RLS multi-tenant) | `cmd/api/main.go` (API), `cmd/migrate/main.go` (migraciones) |
| `frontend/` | React 19 + Vite 8 + TypeScript 7 + Tailwind v4 + TanStack Query + RHF + Zod | `src/main.tsx` → `src/App.tsx` |
| `db/` | `init/` (rol `app_user` de dev) y `tests/` (tests SQL de RLS y doble reserva) | — |
| `docs/` | Índice en `docs/README.md`; migración Java→Go en `docs/migracion-go/` (ADRs) | — |

> El backend Spring Boot (`backend/`) y el sitio estático vanilla de la raíz fueron retirados de `main`; siguen en el historial de git y en las ramas `backend`/`frontend`.

## Backend Go (`backend-go/`)

- **Run:** `make run` (requiere `DATABASE_URL` y `JWT_SECRET`; ver `.env.example`). O `docker compose up --build` desde la raíz.
- **Tests:** `make test` (unitarias + HTTP, `-race`), `make test-integration` (PostgreSQL real vía `TEST_DATABASE_URL_BASE`), `make test-db` desde la raíz (SQL de RLS).
- **Calidad:** `make lint` (golangci-lint v2.15.0), `make vuln` (govulncheck). Ambos deben dar 0 hallazgos.
- **Puerto:** 8080.
- **Config fail-closed:** `JWT_SECRET` ≥32 bytes obligatoria; `APP_ENV` default `prod` → aborta si la conexión es superusuario/BYPASSRLS. En local usar `APP_ENV=dev`.
- **Migraciones:** `db/migrations/` embebidas en el binario. `AUTO_MIGRATE=true` en dev; en prod `go run ./cmd/migrate` con `MIGRATE_DATABASE_URL` (rol dueño). Nunca editar una migración ya aplicada: crear la siguiente (`00000N_*.up/down.sql`).
- **SQL:** queries en `db/queries/*.sql` → código generado con `make sqlc` en `internal/platform/database/sqlc/` (no editar a mano). Todas las queries filtran por `tenant_id`; las transacciones de negocio usan `database.WithTenantTx` (setea `app.tenant_id` para RLS).
- **Contrato:** `api/openapi.json` es la fuente de verdad del API; el frontend genera sus tipos desde ahí.
- **Endpoints clave:** `/turnos`, `/turnos/findDateTimes`, `/servicios`, `/clientes`, `/administradores/login` (JWT), `/estadisticas/panel`, `/api/admin/settings/*`, `/api/public/settings`, `/api/public/theme.css`, `/healthz`, `/readyz`.
- **Auth:** login devuelve `{ id, nombre, apellido, email, token, rol }`; JWT HS256 con claims `sub`, `tenant`, `rol`. Tokens sin `tenant` → 401. Roles `owner`/`staff` (settings solo `owner`).
- **Tenant:** rutas públicas lo resuelven por `Host` (fallback `DEFAULT_TENANT_SLUG`); rutas autenticadas por el claim del JWT. Nunca desde una cabecera del cliente (salvo `X-Tenant-Slug` con `APP_ENV=dev`).
- **Contraseñas:** Argon2id `m=19456,t=2,p=1` (compatible con los hashes que dejó Spring).
- **Fixtures:** `backend-go/testdata/` (settings válidos/inválidos, vectores de contraste).

## Frontend React (`frontend/`)

- **Dev server:** `bun i && bun dev` (Vite, puerto 5173).
- **Checks:** `bun run lint` (oxlint), `bun run typecheck`, `bun run test` (Vitest), `bun run test:e2e` (Playwright, mockea el API), `bun run build` → `dist/`.
- **API config:** `VITE_API_BASE_URL` en `frontend/.env` (ver `.env.example`). La CSP `connect-src` se deriva de esa misma variable en `vite.config.ts`.
- **Auth:** token Bearer en `sessionStorage`; ante `401 /errors/invalid-token` limpia sesión y redirige a `/login`.
- **Tipados:** `src/types/api.ts` es GENERADO desde `backend-go/api/openapi.json` — no editarlo a mano; regenerar con `bun run types`.
- **Atajos de teclado:** `useHotkeys` + registro por scope (`global`/`cliente`/`admin`) con panel de ayuda (`Shift+?`). Ver `docs/atajos-teclado.md`.

## Repo-wide

- `make up` / `make down` (PostgreSQL de docker-compose), `make verify` (todo: fmt, vet, lint-sql, migraciones, tests Go, SQL y frontend).
- CI: `.github/workflows/backend-go-ci.yml`, `.github/workflows/frontend-ci.yml`.
- Docs: `README.md`, `backend-go/README.md`, `frontend/README.md`, `docs/README.md` (índice), `docs/DEPLOY_NOTES.md` (producción).
