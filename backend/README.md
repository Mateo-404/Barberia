# Backend — Spring Boot REST API

API REST para la gestión de barbería: clientes, servicios, turnos y administración (JWT). Expone Swagger UI en desarrollo.

> Guía general: [README raíz](../README.md) · [Índice de docs](../docs/README.md) · [AGENTS.md](../AGENTS.md)

## Stack

- **Spring Boot 3.4** · **Java 21** · **Maven Wrapper**
- **Spring Data JPA** · **PostgreSQL** · **H2** (tests/desarrollo sin Postgres)
- **Spring Security** + **Argon2PasswordEncoder** · **jjwt (api/impl/jackson)**
- **Bean Validation** · **Springdoc OpenAPI (Swagger UI)**
- **Lombok**

## Requisitos

- Java 21 (JDK)
- PostgreSQL 15+ (opcional: usar perfil `h2`)

## Configuración por perfiles

| Perfil | Uso | DB | Notas |
|---|---|---|---|
| `default` (sin especificar) | Dev con Postgres | PostgreSQL vía `DB_URL/DB_USER/DB_PASSWORD` | Usa `ddl-auto=update` |
| `h2` | Dev sin instalar Postgres | H2 en memoria `testdb` | `ddl-auto=create-drop` (ideal para prueba rápida) |
| `test` | Tests | H2 en memoria | Config en `src/test/resources/application-test.properties`. Usa un `jwt.secret` fijo para tests. |
| `prod` | Producción (PaaS/hosting) | PostgreSQL | Deshabilita `springdoc.api-docs` y `springdoc.swagger-ui` para no exponerla. |

## Variables de entorno

| Variable | Requerida | Default | Descripción |
|---|---|---|---|
| `DB_URL` | Solo con perfil default | — | `jdbc:postgresql://localhost:5432/barber` |
| `DB_USER` | Solo con perfil default | — | Usuario Postgres |
| `DB_PASSWORD` | Solo con perfil default | — | Contraseña Postgres |
| `JWT_SECRET` | **Sí** | — | Secreto HS256 (mín. 32 bytes). Fail-closed si falta. |
| `ADMIN_SEED_EMAIL` | No | `admin@admin.com` | Email del admin creado en el primer arranque si no existe |
| `ADMIN_SEED_PASSWORD` | No | `admin1234` | Contraseña del admin creado en el primer arranque si no existe |

Ver `.env.example`. El arranque carga `.env` si lo pasás por variable de entorno o lo inyectás en el runner.

## CORS

Configurado vía `barberCorsConfigurationSource` (bean con `@Qualifier("corsConfigurationSource")`):

- `@Profile("!prod")`: `allowedOriginPatterns = "http://localhost:5173", "http://localhost:5174"` con credenciales permitidas (`Allow-Credentials: true`).
- `@Profile("prod")`: `allowedOrigin = "https://mateo-404.github.io"` **sin credenciales**.

## Comandos

Ejecutar desde `backend/`:

```bash
# H2 (más rápido, sin Postgres)
SPRING_PROFILES_ACTIVE=h2 JWT_SECRET=$(openssl rand -base64 32) ./mvnw spring-boot:run

# Postgres (default)
JWT_SECRET=$(openssl rand -base64 32) DB_URL=jdbc:postgresql://localhost:5432/barber DB_USER=postgres DB_PASSWORD=postgres ./mvnw spring-boot:run

# Tests
./mvnw test        # 64 tests, todos verdes

# Build
./mvnw clean package

# Abrir Swagger UI en dev
open http://localhost:8080/swagger-ui.html
```

El servidor corre en `http://localhost:8080`.

## Endpoints principales

| Ruta | Método | Descripción | Auth |
|---|---|---|---|
| `/turnos` | GET/POST | CRUD de turnos (GENERIC) | Público en create, dependiendo de uso |
| `/turnos/findDateTimes` | GET | Horarios ocupados para una fecha | Público |
| `/servicios` | GET/POST | Servicios | Público |
| `/clientes` | GET/POST | Clientes | — |
| `/administradores/login` | POST | Login → devuelve `{id,nombre,apellido,email,token}` | Público |
| `/administradores/all` | GET | Lista todos | Requiere JWT (según configuración) |
| `/estadisticas/panel` | GET | Panel estadístico (kpis, servicios, horarios) | Público según uso |
| `/ping` | GET | TestController | Público |

> Los endpoints CRUD heredan de `GenericController`. Muchos reads son públicos (consumo desde frontend cliente/admin). Para `/administradores/*` el flujo de acceso se define en `SecurityConfig`.

## Autenticación JWT

- Login POST `/administradores/login` con `{email, contrasenia}` → retorna `LoginResponseDTO` con campo `token` (Bearer).
- El frontend envía `Authorization: Bearer <token>` en cada request a endpoints protegidos.
- En caso de `401 /errors/invalid-token`, el frontend limpia sesión y redirige a `/login`.
- `jwt.secret` es obligatoria (sin valor default). Usá al menos 32 bytes para HS256.

## Seguridad

- Contraseñas: `Argon2PasswordEncoder` (OWASP 2025). No se persisten en texto plano.
- Seed admin: `AdminSeedRunner` crea el primer admin si la tabla está vacía (usa `admin.seed.*`).
- Spring Security: ver `SecurityConfig.java` para el detalle de rutas públicas/protegidas.

## Tests

- 64 tests: `@WebMvcTest` (controller slices), `@DataJpaTest` sobre H2 (repositorios: `ITurnoRepositoryTest`, `IClienteRepositoryTest`, `IAdministradorRepositoryTest`) e integración de Argon2.
- Ejecutar: `./mvnw test`
- CI: `.github/workflows/backend-ci.yml` corre build + tests en push/PR a `main` o `backend` cuando cambian archivos bajo `backend/**`.

## OpenAPI / Contratos

- `backend/openapi-spec.json` es la especificación exportada (se puede regenerar desde Swagger `/v3/api-docs` si el backend está levantado).
- `frontend/src/types/api.ts` se genera a partir de este archivo con `bun run types` desde `frontend/`.
- **Una sola fuente de verdad:** el backend. No editar los tipos a mano.
- En perfil `prod`, la UI de Swagger (`/swagger-ui.html`) y `api-docs` están deshabilitados.

## Notas sobre la arquitectura

- Generic CRUD con mappers: `GenericController` + `GenericService` + DTOs Request/Response para las 4 entidades (Cliente, Turno, Servicio, Administrador). El `update()` genérico acepta entidad JPA por diseño histórico (ver `docs/tech-debt-barberia-fase3b.md`).
- Panel de estadísticas separado: `PanelEstadisticaDTOService` + `PanelEstadisticasControllerDTO`.
- `BarberBackendApplication.java` es el entrypoint.
