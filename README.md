# Sistema de Gestión de Barbería

> **Frontend moderno:** el directorio `frontend/` contiene la app activa (React 19 + Vite + TypeScript + Tailwind), con wizard de reservas y panel de administración. El sitio estático de la raíz (HTML/CSS/JS vanilla) queda como referencia histórica.
>
> 📚 **Documentación:** [README del frontend](frontend/README.md) · [README del backend](backend/README.md) · [Índice de docs](docs/README.md) · `AGENTS.md` (arquitectura y comandos para agentes).

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

### Suite automatizada (backend)

El backend tiene **64 tests automatizados** (controllers con MockMvc, servicios, repositorios sobre H2 e integración de Argon2):

```bash
cd backend && ./mvnw test
```

En cada push/PR que toque `backend/**` corren en CI (`.github/workflows/backend-ci.yml`).
El frontend React todavía no tiene tests automatizados ni CI propio.

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
![](public/Diagrama_Barberia-Casos%20de%20Uso.jpg)
### Diagrama de Clases
![](public/Diagrama_Barberia-Clases.jpg)

---

## Demo en Línea
Demo interactiva de la app React, 100 % en el navegador (sin backend): la API se simula con
[MSW](https://mswjs.io) y trae una **visita guiada** que acompaña la reserva de un turno y el panel del barbero.

🔗 **https://mateo-404.github.io/Barberia/**

- Se publica con `.github/workflows/pages.yml` al hacer push a la rama `demo` (Settings → Pages → Source: *GitHub Actions*).
- Build local: `cd frontend && VITE_DEMO=true VITE_BASE=/Barberia/ bun run build`.
- Datos de ejemplo: `frontend/src/mocks/db.ts` (espeja `DemoDataSeedRunner`). Admin demo: `admin@admin.com` / `admin1234`.
- El backend real (Spring Boot) se corre en local con `SPRING_PROFILES_ACTIVE=h2`; ver más abajo.

> **🔐 Admin de prueba:** El primer administrador se crea automáticamente al iniciar el backend
> por primera vez si la base de datos está vacía (`AdminSeedRunner`). Las credenciales se
> configuran vía variables de entorno:
>
> ```
> ADMIN_SEED_EMAIL=admin@barberia.com
> ADMIN_SEED_PASSWORD=miPasswordSegura123
> ```
>
> Consultá `backend/.env.example` para la lista completa de variables.
> En Railway (o tu PaaS), configuralas en el panel de Environment Variables.  

---

## Capturas de Pantalla

### 1. Reserva de Turno (Cliente)
![Reserva de turno](public/screenshots/reserva_turno.png)

### 2. Panel de Administración
![Panel de administración](public/screenshots/panel_admin.png)

### 3. Estadísticas
![Estadísticas](public/screenshots/estadisticas.png)

---

## Organización del Proyecto

- **Repositorio en GitHub:** [Barberia](https://github.com/Mateo-404/Barberia)  
- **Tablero de Tareas:** [GitHub Projects](https://github.com/users/Mateo-404/projects/9)

---

## Instalación

### Prerrequisitos

- **Java 21** (JDK)
- **Maven Wrapper** (incluido en `backend/mvnw`)
- **PostgreSQL** 15+ (opcional: con el perfil `h2` corre sobre H2 en memoria)
- **Bun** (o npm) para el frontend React

### Backend

#### Variables de entorno

El backend lee las siguientes variables de entorno:

| Variable | Descripción | Ejemplo |
|---|---|---|
| `DB_URL` | URL de conexión JDBC | `jdbc:postgresql://localhost:5432/barber` |
| `DB_USER` | Usuario de base de datos | `postgres` |
| `DB_PASSWORD` | Contraseña del usuario | `postgres` |
| `JWT_SECRET` | **Obligatoria.** Secreto HS256 de al menos 32 bytes. Generar con `openssl rand -base64 32` | `kR9/xZ...` |
| `ADMIN_SEED_EMAIL` | Email del admin seed (opcional) | `admin@admin.com` |
| `ADMIN_SEED_PASSWORD` | Contraseña del admin seed (opcional) | `admin1234` |

> ⚠️ Sin `JWT_SECRET` el backend **no arranca** (fail-closed). `ADMIN_SEED_*` tienen defaults.
> Copiar `backend/.env.example` a `backend/.env` y completar los valores. Ver más detalles en [backend/README.md](backend/README.md).

#### Desarrollo sin PostgreSQL (perfil `h2`)

Usar H2 en memoria — ideal para desarrollo rápido sin instalar PostgreSQL. El perfil se llama **`h2`**:

```bash
SPRING_PROFILES_ACTIVE=h2 JWT_SECRET=$(openssl rand -base64 32) ./mvnw spring-boot:run
```

#### Comandos (desde `backend/`)

```bash
# Ejecutar el servidor (Postgres requerido)
./mvnw spring-boot:run

# Ejecutar con perfil H2 (sin Postgres)
SPRING_PROFILES_ACTIVE=h2 ./mvnw spring-boot:run

# Ejecutar tests (64 tests)
./mvnw test

# Build completo (package)
./mvnw clean package
```

El servidor arranca en `http://localhost:8080`.
Documentación interactiva de la API (Swagger UI): [http://localhost:8080/swagger-ui.html](http://localhost:8080/swagger-ui.html) — deshabilitada en el perfil `prod`.

### Frontend (app React)

```bash
cd frontend
bun i                       # instalar dependencias
cp .env.example .env        # VITE_API_BASE_URL (default http://localhost:8080)
bun dev                     # dev server en http://localhost:5173
```

Build de producción: `bun run build` → `frontend/dist/`.
Guía completa (estructura, auth, atajos de teclado): [frontend/README.md](frontend/README.md).

> El backend debe estar corriendo (en `:8080`) para que la app funcione.
> En desarrollo el CORS permite `localhost:5173/5174` (perfil por defecto).

### Ramas

| Rama | Propósito |
|---|---|
| `main` | Desarrollo activo (monorepo: `backend/`, `frontend/`, sitio estático legacy y docs) |
| `backend` | Histórica — despliegue del API en PaaS (Railway) |
| `frontend` | Histórica — sitio estático legacy publicado en GitHub Pages |

---

## Licencia

Distribuido bajo [GNU GPL v3](LICENSE).
