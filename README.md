# Sistema de Gestión de Barbería

> **Stack:** backend en **Go** (`backend-go/`: chi + pgx + sqlc + PostgreSQL con RLS multi-tenant) y frontend en **React 19** (`frontend/`: Vite + TypeScript + Tailwind) con wizard de reservas y panel de administración. El backend Spring Boot original y el sitio estático vanilla fueron retirados tras la migración (siguen en el historial de git y en las ramas `backend`/`frontend`).
>
> 📚 **Documentación:** [README del backend](backend-go/README.md) · [README del frontend](frontend/README.md) · [Índice de docs](docs/README.md) · [AGENTS.md](AGENTS.md) (arquitectura y comandos para agentes).

---

## Tabla de Contenidos

1. [Definición del Alcance](#definición-del-alcance)
2. [Breve Descripción del Problema](#breve-descripción-del-problema)
3. [Qué Tiene que Hacer el Sistema](#qué-tiene-que-hacer-el-sistema)
4. [Funciones del Sistema](#funciones-del-sistema)
5. [Requerimientos](#requerimientos)
   - [Requerimientos del Usuario](#requerimientos-del-usuario)
   - [Requerimientos de Sistema](#requerimientos-de-sistema)
   - [Requerimientos No Funcionales](#requerimientos-no-funcionales)
6. [Casos de Uso](#casos-de-uso)
   - [Usuario](#usuario)
   - [Administrador](#administrador)
   - [Sistema](#sistema)
7. [Pruebas de Requerimientos](#pruebas-de-requerimientos-del-sistema)
8. [Diagramas](#diagramas)
9. [Demo en linea](#demo-en-línea)
10. [Capturas de Pantalla](#capturas-de-pantalla)
11. [Organización del Proyecto](#organización-del-proyecto)
12. [Instalación](#instalación)
13. [Licencia](#licencia)

---

## Definición del Alcance
Se desarrollará un sistema para gestionar reservas en una barbería. Permitirá optimizar la experiencia del cliente y la administración del negocio, resolviendo problemas actuales como confusiones en las reservas y falta de estadísticas claras.

El sistema está orientado a pequeñas barberías, pero su estructura permite escalar a otros rubros. El proyecto se desarrolla como parte de la Práctica Profesional Supervisada de la Tecnicatura Universitaria en Programación (UTN - Rafaela).

---

## Breve Descripción del Problema
La barbería enfrenta dificultades con la gestión manual de las reservas, lo que genera superposición de horarios, frustración entre clientes y pérdida de ingresos. El objetivo es implementar un sistema automatizado que simplifique estos procesos.

---

## Qué Tiene que Hacer el Sistema
- Reservas en línea con selección de servicios, fechas y horarios.
- Gestión de datos de clientes.
- Panel de control para el administrador con calendario visual, estadísticas y reportes.

---

## Funciones del Sistema
1. Reserva de turnos en línea.
2. Selección de tipo de servicio.
3. Selección de fecha y hora con disponibilidad en tiempo real.
4. Recopilación segura de datos del cliente.
5. Panel administrativo para gestionar turnos y reportes.

---

## Requerimientos

### Requerimientos del Usuario
1. **Reserva de turnos**: Selección de servicio, fecha y hora.
2. **Gestión de servicios**: Visualización de opciones con descripciones y precios.

### Requerimientos de Sistema
- **Usuario**: Selección de servicio, calendario interactivo, formulario de datos.
- **Administrador**: Panel de control para gestionar reservas, modificar o cancelar citas, y acceder a estadísticas.
- **Sistema**: Recopilación de datos para reportes.

### Requerimientos No Funcionales
- **Compatibilidad y diseño responsivo**: Uso óptimo en cualquier dispositivo.
- **Seguridad**: Contraseñas de administrador con hash Argon2 (no se guardan en texto plano), autenticación con JWT (HS256) y CORS restringido a los orígenes autorizados. El tráfico depende del HTTPS del hosting.

---

## Casos de Uso

### Usuario
1. Selección de servicio.
2. Selección de fecha y hora en un calendario interactivo.
3. Registro de datos personales en un formulario.

### Administrador
1. **Gestión de reservas**: Modificar o cancelar citas a través del panel.
2. **Reportes**: Visualización de estadísticas y gráficos de ganancias, servicios más solicitados, y clientes frecuentes.

### Sistema
1. Generación de reportes.
2. Sincronización de datos en tiempo real.

---

## Pruebas de Requerimientos del Sistema

### Suites automatizadas

| Suite | Qué cubre | Comando |
|---|---|---|
| Backend unitario + HTTP | handlers, contrato JSON, auth (Argon2id/JWT), validación, settings | `make -C backend-go test` |
| Backend integración | PostgreSQL real: migraciones, adopción de esquema Hibernate, aislamiento de tenant, doble reserva concurrente | `make -C backend-go test-integration` |
| SQL | políticas RLS y unicidad de turnos (`db/tests/`) | `make test-db` |
| Frontend | Vitest + Testing Library (unitarios) y Playwright (e2e) | `cd frontend && bun run test && bun run test:e2e` |

CI: `.github/workflows/backend-go-ci.yml` (gofmt, vet, tests con `-race`, integración y SQL contra Postgres, golangci-lint, govulncheck, build de la imagen Docker) y `.github/workflows/frontend-ci.yml` (oxlint, typecheck, tests, build, e2e).

### Casos de validación manual

#### Caso de Prueba 1: Selección de Servicio
- **ID**: CP-001
- **Descripción**: Verificar que el usuario pueda seleccionar un tipo de servicio.
- **Prioridad**: Alta
- **Resultados Esperados**: El servicio seleccionado se almacena correctamente.

#### Caso de Prueba 2: Selección de Fecha y Hora
- **ID**: CP-002
- **Descripción**: Validar la selección de fechas y horarios disponibles.
- **Prioridad**: Media
- **Resultados Esperados**: Fecha y hora seleccionadas correctamente.

#### Caso de Prueba 3: Registro de Datos del Cliente
- **ID**: CP-003
- **Descripción**: Comprobar que los datos del cliente se registren con éxito.
- **Prioridad**: Alta
- **Resultados Esperados**: Los datos se almacenan correctamente.

---

## Diagramas
### Diagrama de Casos de Uso
![](docs/assets/Diagrama_Barberia-Casos%20de%20Uso.jpg)
### Diagrama de Clases
![](docs/assets/Diagrama_Barberia-Clases.jpg)

---

## Demo en Línea
El sitio estático legacy (anterior a la app React) sigue publicado para pruebas:

🔗 **Frontend (Cliente):** [https://mateo-404.github.io/Barberia/](https://mateo-404.github.io/Barberia/)  
🔗 **Frontend (Administrador):** [https://mateo-404.github.io/Barberia/admin/login.html](https://mateo-404.github.io/Barberia/admin/login.html)  

👉 Estado actual del despliegue:
- **GitHub Pages** sirve el sitio estático legacy desde la rama `frontend`.
- La **app React** (`frontend/`) y el **backend Go** todavía no tienen despliegue automático; el procedimiento de producción (roles de BD, migraciones, variables) está en [docs/DEPLOY_NOTES.md](docs/DEPLOY_NOTES.md).

> **🔐 Admin inicial:** si el tenant no tiene administradores, el backend crea uno al arrancar
> con `ADMIN_SEED=true`, usando `ADMIN_SEED_EMAIL` / `ADMIN_SEED_PASSWORD` (hasheada con Argon2id).
> Lista completa de variables: [`backend-go/.env.example`](backend-go/.env.example).

---

## Capturas de Pantalla

### 1. Reserva de Turno (Cliente)
![Reserva de turno](docs/assets/screenshots/reserva_turno.png)

### 2. Panel de Administración
![Panel de administración](docs/assets/screenshots/panel_admin.png)

### 3. Estadísticas
![Estadísticas](docs/assets/screenshots/estadisticas.png)

---

## Organización del Proyecto

- **Repositorio en GitHub:** [Barberia](https://github.com/Mateo-404/Barberia)  
- **Tablero de Tareas:** [GitHub Projects](https://github.com/users/Mateo-404/projects/9)

---

## Instalación

### Prerrequisitos

- **Go** 1.27+ (el toolchain exacto se descarga solo vía `go.mod`)
- **PostgreSQL** 15+ (o Docker para levantarlo con `docker compose`)
- **Bun** 1.4+ para el frontend

### Opción A — todo con Docker (PostgreSQL + backend)

```bash
JWT_SECRET=$(openssl rand -base64 32) docker compose up --build
```

API en `http://localhost:8080` con migraciones automáticas y admin `admin@admin.com` / `admin1234`.

### Opción B — backend local

```bash
cd backend-go
cp .env.example .env        # completar JWT_SECRET (openssl rand -base64 32)
set -a; source .env; set +a
make run                    # compila y arranca en :8080
```

Variables principales (todas en [`backend-go/.env.example`](backend-go/.env.example)):

| Variable | Descripción | Ejemplo |
|---|---|---|
| `DATABASE_URL` | Conexión PostgreSQL de la app | `postgres://postgres:postgres@localhost:5432/barber?sslmode=disable` |
| `JWT_SECRET` | **Obligatoria**, ≥32 bytes | `openssl rand -base64 32` |
| `APP_ENV` | `dev` en local; cualquier otro valor (default `prod`) exige un rol sin SUPERUSER/BYPASSRLS | `dev` |
| `AUTO_MIGRATE` | Aplica las migraciones al arrancar | `true` |
| `CORS_ALLOW_ORIGINS` | Orígenes permitidos, separados por coma | `http://localhost:5173` |
| `ADMIN_SEED*` | Admin inicial | `admin@admin.com` / `admin1234` |

> ⚠️ Sin `JWT_SECRET` (o con un rol privilegiado fuera de `APP_ENV=dev`) el backend **no arranca** (fail-closed).

Más detalle (estructura, tests, multi-tenancy, seguridad): [backend-go/README.md](backend-go/README.md).

### Frontend (app React)

```bash
cd frontend
bun i                       # instalar dependencias
cp .env.example .env        # VITE_API_BASE_URL (default http://localhost:8080)
bun dev                     # dev server en http://localhost:5173
```

Build de producción: `bun run build` → `frontend/dist/`.
Guía completa (estructura, auth, atajos de teclado): [frontend/README.md](frontend/README.md).

### Verificación completa

```bash
make up          # PostgreSQL (docker compose)
make verify      # fmt, vet, lint-sql, migraciones, tests Go, tests SQL y frontend
```

### Ramas

| Rama | Propósito |
|---|---|
| `main` | Desarrollo activo (monorepo: `backend-go/`, `frontend/`, `db/` y docs) |
| `backend` | Histórica — backend Spring Boot (retirado de `main`) |
| `frontend` | Histórica — sitio estático legacy publicado en GitHub Pages |

---

## Licencia

Distribuido bajo [GNU GPL v3](LICENSE).
