# Barbería — AGENTS.md

## Architecture

Proyecto en dos partes: **Spring Boot backend** + **frontends**. Hay dos frontends que conviven:

| Directory | Tech | Entrypoint |
|---|---|---|
| `backend/` | Spring Boot 3.4, Java 21, Maven, JPA, PostgreSQL | `BarberBackendApplication.java` |
| `frontend-react/` | **React 19 + Vite 8 + TypeScript 6 + Tailwind v4 + TanStack Query + RHF + Zod**. App moderna de reservas + panel admin. | `src/main.tsx` → `src/App.tsx` |
| root `index.html` | Vanilla HTML/CSS/JS (ES modules), Bootstrap 5 — sitio público legacy | `index.html` (client), `admin/login.html` (admin) |

> El frontend React es la app activa. El estático vanilla del root queda solo como referencia/history.

## Backend

- **Build & run:** `./mvnw spring-boot:run` (desde `backend/`)
- **Test:** `./mvnw test` (64 tests: controller slices + Argon2 integration)
- **Puerto:** 8080
- **DB:** PostgreSQL en `localhost:5432/barber` (user/pass `postgres`/`postgres`). H2 en classpath.
- **JPA:** `ddl-auto=update`
- **CORS (dos perfiles, bean `barberCorsConfigurationSource` via @Qualifier):**
  - `@Profile("!prod")` → `allowedOriginPatterns` `localhost:5173/5174` (Vite) con credenciales
  - `@Profile("prod")` → `allowedOrigin("https://mateo-404.github.io")` sin credenciales
- **API base (frontend React):** `VITE_API_BASE_URL` (default `http://localhost:8080`)
- **Endpoints clave:** `/turnos`, `/turnos/findDateTimes` (ocupados), `/servicios`, `/administradores/login` (JWT), `/estadisticas/panel`, `/clientes`.
- **Auth:** Login devuelve `LoginResponseDTO { id, nombre, apellido, email, token }`. El frontend guarda admin + token en `sessionStorage` y envía `Authorization: Bearer <token>`. Ante `401 /errors/invalid-token` limpia sesión y redirige a `/login`.
- **Seed:** `AdminSeedRunner` crea admin con `admin.seed.email` / `admin.seed.password` (defaults `admin@admin.com` / `admin1234`).
- **Seguridad:** `jwt.secret=${JWT_SECRET}` (obligatorio). Contraseñas con Argon2PasswordEncoder (OWASP 2025).

## Frontend React

- **Dev server:** `bun i && bun dev` (Vite, puerto 5173).
- **Build:** `bun run build` → `frontend-react/dist/`.
- **API config:** `VITE_API_BASE_URL` en `frontend-react/.env` (ver `.env.example`).
- **Auth:** token Bearer en `sessionStorage`; redirect a `/login` ante 401.
- **Tipados:** `src/types/api.ts` es GENERADO por `openapi-typescript` desde `backend/openapi-spec.json` — no editarlo a mano; regenerar con `bun run types`.
