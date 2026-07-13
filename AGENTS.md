# Barbería — AGENTS.md

## Architecture

Two-part project: **Spring Boot backend** + **vanilla HTML/CSS/JS frontend** (no framework, no build step, no npm).

| Directory | Tech | Entrypoint |
|---|---|---|
| `backend/` | Spring Boot 3.4, Java 21, Maven, JPA, PostgreSQL | `BarberBackendApplication.java` |
| `frontend/` | Vanilla HTML/CSS/JS (ES modules), Bootstrap 5 | `index.html` (client), `admin/login.html` (admin) |

## Backend

- **Build & run:** `./mvnw spring-boot:run` (from `backend/`)
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

## Flujo de ramas
- main: SOLO recibe código estable, ya probado. NUNCA se commitea
  directo a main durante desarrollo activo.
- backend / frontend: ramas de trabajo activo para cada área. Todo
  desarrollo (fases, fixes, features) va acá primero.
- Sync a main: solo cuando el trabajo está confirmado como estable,
  vía comando explícito aprobado por el usuario, sincronizando SOLO
  la carpeta correspondiente (backend/ o frontend/), nunca un
  merge completo de rama.

## Convenciones de commit
- Conventional Commits obligatorio: <type>: <description>
  Tipos permitidos: feat, fix, ci, refactor, test, docs, chore
- Mensajes en español, consistente con el resto del proyecto

## Principio de dependencias
- Preferir librerías estándar y mantenidas activamente sobre código
  custom, SIEMPRE que resuelvan un problema genérico (seguridad,
  validación, serialización, mapeo, manejo de errores).
- Ejemplos ya aplicados: spring-security-crypto (Argon2), Bean
  Validation (Jakarta), ProblemDetail nativo de Spring (RFC 7807)
  en vez de formato de error custom.
- NO aplica a lógica de negocio específica del dominio (reglas de
  horario de turnos, validaciones de la barbería) — eso siempre
  queda en código propio.
- Antes de escribir una utilidad o mapper a mano, evaluar si existe
  una librería estándar del ecosistema Spring que ya resuelva el
  problema (ej: MapStruct para mapeo DTO↔Entity en vez de mappers
  manuales).

## Proceso de trabajo por fases
- Cada fase se planea antes de implementar, mostrando diseño +
  archivos afectados + impacto en tests existentes, ANTES de
  escribir código.
- Cambios que afecten firmas de métodos existentes requieren
  verificar tests que dependan de la firma/mensaje actual, con
  evidencia real (mostrar el archivo, no asumir).
- Build completo (./mvnw clean package) + resultado de los 33 tests
  se muestra después de cada fase, no solo al final de una sesión.
- No se avanza a la siguiente fase sin confirmación explícita del
  usuario.
