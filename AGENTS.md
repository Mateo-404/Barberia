# Barbería — AGENTS.md

## Architecture

Two-part project: **Spring Boot backend** + **vanilla HTML/CSS/JS frontend** (no framework, no build step, no npm).

| Directory | Tech | Entrypoint |
|---|---|---|
| `barberBackend/` | Spring Boot 3.4, Java 21, Maven, JPA, PostgreSQL | `BarberBackendApplication.java` |
| `frontend/` | Vanilla HTML/CSS/JS (ES modules), Bootstrap 5 | `index.html` (client), `admin/login.html` (admin) |

## Backend

- **Build & run:** `./mvnw spring-boot:run` (from `barberBackend/`)
- **Test:** `./mvnw test` (single test class, context-loads only)
- **Port:** 8080 (default Spring Boot)
- **DB:** PostgreSQL at `localhost:5432/barber` (user/pass: `postgres`/`postgres`). H2 also on classpath as runtime dep.
- **JPA:** `ddl-auto=update` — no migration tooling.
- **CORS:** whitelists `localhost:3000` and `127.0.0.1:3000` only.
- **API base:** `http://localhost:8080`
- **Endpoints:** `/turnos`, `/clientes`, `/servicios`, `/administradores`, `/estadisticas`
- **Custom endpoint:** `GET /turnos/findDateTimes` returns `List<String>` of occupied `LocalDateTime` in ISO format.
- **Pattern:** Generic REST controller with `GenericController<T, ID, S>` — see `generics/` package. Custom controllers extend it.
- **Models:** Lombok (`@Getter @Setter @NoArgsConstructor @AllArgsConstructor`). `Turno` has `@ManyToOne` to `Cliente` and `Servicio`.

## Frontend

- **No build step** — served directly as static files.
- **JS modules:** use `import`/`export` (ES module syntax in `config.js`).
- **API config:** `API_BASE_URL` in `frontend/js/config.js` — change here for different backend origin.
- **CORS note:** when developing, frontend must be served on `localhost:3000` (the CORS-allowed origin). Use e.g. `python3 -m http.server 3000` from `frontend/`.
- **Admin login:** `admin/login.html` POSTs to `/administradores/login` with `{email, contrasenia}`. Uses temporary `setTimeout(1000)` wrapper around fetch.
- **Reserva flow:** 3-step wizard (service → date/time → personal data), submits `POST /turnos`.

## Deployment

- **`frontend` branch** → GitHub Pages (serves `frontend/` as root)
- **`backend` branch** → Railway (or similar PaaS)

## Notable quirks

- `.gitignore` at repo root combines rules for both Java/Maven and IDE artifacts.
- The `public/` directory at root contains diagrams and screenshots only — not served by any server.
- No CI, no linter, no formatter config.

## Commits

Commits MUST follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:
`<type>: <description>`

Allowed types: `feat`, `fix`, `ci`, `refactor`, `test`, `docs`, `chore`.
