# Barbería — AGENTS.md

## Architecture

Two-part project: **Spring Boot backend** + **frontend**. Hay dos frontends que conviven:

| Directory | Tech | Entrypoint |
|---|---|---|
| `barberBackend/` (rama `backend`) | Spring Boot 3.4, Java 21, Maven, JPA, PostgreSQL | `BarberBackendApplication.java` |
| `frontend/` | Vanilla HTML/CSS/JS (ES modules), Bootstrap 5 — sitio público legacy | `index.html` (client), `admin/login.html` (admin) |
| `frontend-react/` (rama `frontend`) | **React 19 + Vite 8 + TypeScript 6 + Tailwind v4 + TanStack Query + RHF + Zod**. App moderna de reservas + panel admin. | `src/main.tsx` → `src/App.tsx` |

> El frontend React es la app activa. `frontend/` (vanilla) queda solo como referencia/history.

## Backend

- **Ubicación clon independiente:** `~/Documentos/repositorios/Barberia-backend-run/` (rama `backend`). El frontend React NO edita archivos de backend.
- **Build & run:** `./mvnw spring-boot:run` (desde `backend/`)
- **Test:** `./mvnw test`
- **Puerto:** 8080
- **DB:** PostgreSQL en `localhost:5432/barber` (user/pass `postgres`/`postgres`). H2 en classpath.
- **JPA:** `ddl-auto=update`
- **CORS (dos perfiles):**
  - `@Profile("!prod")` → `allowedOriginPatterns("http://localhost:*")` con credenciales
  - `@Profile("prod")` → `allowedOrigin("https://mateo-404.github.io")` sin credenciales
- **API base (frontend React):** `VITE_API_BASE_URL` (default `http://localhost:8080`)
- **Endpoints clave:** `/turnos`, `/turnos/findDateTimes` (ocupados), `/servicios`, `/administradores/login` (JWT), `/estadisticas/panel`, `/clientes`.
- **Auth:** Login devuelve `LoginResponseDTO { id, nombre, apellido, email, token }`. El frontend guarda admin + token en `sessionStorage` y envía `Authorization: Bearer <token>`. Ante `401 /errors/invalid-token` limpia sesión y redirige a `/login`.
- **Seed:** `AdminSeedRunner` crea admin con `admin.seed.email` / `admin.seed.password` (defaults `admin@admin.com` / `admin1234`).

## Frontend React (`frontend-react/`)

- **Stack:** Vite 8, React 19, TypeScript 6 (strict), Tailwind v4, `@tanstack/react-query` v5, `react-hook-form` v7 + `zod` v4, `react-router-dom` v7, `class-variance-authority`, `lucide-react`.
- **Comandos (Bun):** `bun run dev`, `bun run build` (`tsc -b && vite build`), `bun run lint` (oxlint), `bun run preview`.
- **Alias:** `@/` → `src/`.
- **Sistema de color:** Todo color vive en variables CSS en `src/index.css` (`:root`). Tailwind resuelve utilidades semánticas (`bg-primary`, `text-muted-foreground`, `border-border`, `ring-focus-ring`, etc.) a esas variables. **No hay hex hardcodeados en componentes.**
  - Paleta: bg `#1c1c1c`, card `#2c2c2c`, primary `#ff6600` (hover `#e65500`), muted `#999`, success `#28a745`, warning `#ffc107`, destructive `#dc3545`.
  - Variables útiles: `--nav-height` (4.5rem), `--focus-ring`, animación `.page-enter`.
- **Rutas:** `/` (Reserva wizard 3 pasos), `/login`, `/admin` (protegida → `AdminDashboard` con KPIs + tabla de turnos + ranking de servicios).
- **API:** `src/lib/api-client.ts` (fetch wrapper con Bearer token, timeout 15s vía AbortController, manejo de `401 invalid-token`). Hooks en `src/api/`.
- **Tipos:** `src/types/api.ts` generado con `openapi-typescript` desde el spec del backend.

## Deployment

- **`frontend` branch** → GitHub Pages (raíz del repo). La app React se sirve desde `frontend-react/` vía build estático.
- **`backend` branch** → Railway (o PaaS similar).

## Notable quirks

- `.gitignore` raíz combina reglas de Java/Maven e IDE.
- El backend y el frontend React viven en clones/distintas ramas; el session de frontend solo toca `frontend-react/` y archivos de infra raíz.
- No hay tests E2E corriendo en CI todavía (handlers MSW presentes en `src/tests/`).
- Commits: Conventional Commits (`feat`, `fix`, `ci`, `refactor`, `test`, `docs`, `chore`).

## Commits

Commits MUST follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:
`<type>: <description>`

Allowed types: `feat`, `fix`, `ci`, `refactor`, `test`, `docs`, `chore`.
