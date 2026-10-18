# Informe de Sesión — Frontend (rama `frontend`, repo `Barberia/`)

## Resumen

Esta sesión cubrió 4 áreas dentro del alcance del frontend:

1. **Fase Frontend 3** — Login, AuthContext, rutas protegidas, DataTable genérica
2. **Testing Setup** — Vitest + testing-library + MSW + tests
3. **Frontend CI** — Workflow GitHub Actions (`frontend-ci.yml`)
4. **Docker Compose** — Dockerfiles para frontend + propuesta de orquestación

---

## 1. Fase Frontend 3 — Login + Auth (13 archivos)

### Arquitectura

```
App.tsx                         → AuthProvider envuelve todas las rutas
├── /                           → Reserva (wizard público)
├── /login                      → Login (form RHF + zodResolver)
└── ProtectedRoute              → sin auth → redirect a /login
    └── AdminLayout             → navbar top + Outlet
        └── /admin              → AdminDashboard (placeholder)
```

### Archivos creados (10 nuevos)

| Archivo | Líneas | Propósito |
|---|---|---|
| `src/context/AuthContext.tsx` | 56 | `{ admin, isAuthenticated, isLoading, login, logout }` con `sessionStorage` interno. Aislado del mecanismo de auth — ningún componente externo toca storage. Internamente usa `useLogin()` mutation de TanStack Query. |
| `src/lib/schemas/login-schema.ts` | 9 | `email: z.string().email()`, `contrasenia: z.string().min(8)` |
| `src/api/auth.ts` | 12 | `useLogin()` → `POST /administradores/login`, tipado contra `api.ts` |
| `src/pages/Login.tsx` | 82 | RHF + zodResolver. Redirect a `/admin` si ya autenticado. Muestra error 401 vía `ApiError.detail` en `setError("root")`. Sin `window.location`. |
| `src/components/ProtectedRoute.tsx` | 10 | `<Navigate to="/login" replace />` si `!isAuthenticated`. Usa `<Outlet />`. |
| `src/components/admin/AdminLayout.tsx` | 35 | Navbar top (logo centrado, nombre usuario izq, logout der) + `<Outlet />` |
| `src/components/ui/DataTable.tsx` | 56 | Genérica `<TData, TValue>` con `@tanstack/react-table` headless + shadcn Table. Estados loading/empty incluidos. |
| `src/components/ui/StatusBadge.tsx` | 19 | Variant badge: `success/error/warning/info` con colores Tailwind |
| `src/components/ui/table.tsx` | 67 | shadcn Table primitives: `Table/Header/Body/Row/Head/Cell` con estilos base |
| `src/lib/utils.ts` | 7 | `cn()` utility con clsx + tailwind-merge |

### Archivos modificados (3 existentes)

| Archivo | Cambio |
|---|---|
| `src/App.tsx` | +AuthProvider envolviendo rutas, +`/login`, +`/admin` con ProtectedRoute + AdminLayout, +AdminDashboard placeholder |
| `package.json` | +`@tanstack/react-table` dependency |
| `bun.lock` | lockfile actualizado |

### Principios aplicados

- **AuthContext aislado**: ningún componente conoce `sessionStorage`. Cuando llegue JWT del backend, solo cambia AuthContext.
- **Sin `window.location`**: toda navegación vía React Router (`<Navigate />`, `useNavigate()`).
- **Tipos desde `api.ts`**: todos los DTOs usan `components["schemas"]["AdministradorResponseDTO"]`, no interfaces manuales.
- **DataTable genérica**: `<TData, TValue>`, sin `any` en firmas públicas.

---

## 2. Testing Setup (Vitest + Testing Library + MSW)

### Dependencias instaladas

```bash
bun add -d vitest @testing-library/react @testing-library/jest-dom \
    @testing-library/user-event jsdom msw
```

### Configuración

Archivo: `vitest.config.ts`
```ts
environment: "jsdom",
setupFiles: ["./src/tests/setup.ts"],
globals: true,
```

Setup file (`src/tests/setup.ts`): importa `@testing-library/jest-dom/vitest` para matchers como `toBeInTheDocument()`.

### Tests escritos (3 suites, 14 tests)

#### `src/tests/turno-schema.test.ts` (10 tests)

Cubre todas las reglas de negocio del schema Zod de turno:

| Test | Escenario |
|---|---|
| ✔ Válido | fecha futura, horario 13-20, todos los campos |
| ✖ Fecha pasada | rechazada por `turnoSchema` |
| ✖ Fuera del horario (9:00) | menor a 13hs |
| ✖ Fuera del horario (21:00) | mayor a 20hs |
| ✖ Minuto no válido (45) | solo 00/30 permitidos |
| ✖ ServicioId negativo | rechazado |
| ✖ Nombre vacío | required |
| ✖ Apellido vacío | required |
| ✖ Teléfono vacío | required |
| ✖ Teléfono muy corto | min(7) |

#### `src/tests/Reserva.test.tsx` (3 tests)

Usa MSW para mockear la API. Prueba comportamiento observable:

| Test | Lo que verifica |
|---|---|
| ✔ Renderiza paso 1 | el usuario ve la lista de servicios cargada desde API |
| ✔ Completa los 3 pasos y muestra confirmación | flujo completo del wizard: seleccionar servicio → elegir horario → datos personales → confirmación con resumen |
| ✔ Muestra "cargando" mientras se envían los datos | loading state durante la mutation |

#### `src/tests/App.test.tsx` (1 test)

| Test | Lo que verifica |
|---|---|
| ✔ Renderiza sin crash | smoke test, confirmación de que App no explota al montar |

### MSW handlers

Archivo: `src/tests/msw/handlers.ts`

```ts
http.get("*/servicios", () =>
    HttpResponse.json([{ id: 1, tipo: "Corte", precio: 1500, duracionMinutos: 30 }]))
http.get("*/turnos/findDateTimes", () =>
    HttpResponse.json(["2026-07-20T14:00:00"]))
http.post("*/turnos", () =>
    HttpResponse.json({ id: 1, fechaHora: "2026-07-20T14:00:00" }, { status: 201 }))
```

### Resultados

```
✓ src/tests/turno-schema.test.ts (10 tests) 12ms
✓ src/tests/Reserva.test.tsx (3 tests) 456ms
✓ src/tests/App.test.tsx (1 test) 89ms

Test Files  3 passed | 3 total
Tests  14 passed | 14 total
```

---

## 3. Frontend CI — GitHub Actions

### Archivo creado: `.github/workflows/frontend-ci.yml`

```yaml
name: Frontend CI

on:
  push:
    branches: [ frontend ]
    paths: [ "frontend-react/**" ]
  pull_request:
    branches: [ frontend ]
    paths: [ "frontend-react/**" ]

jobs:
  build-and-test:
    runs-on: ubuntu-latest
    defaults:
      run:
        working-directory: frontend-react
    steps:
      - uses: actions/checkout@v4
      - uses: oven-sh/setup-bun@v2
        with:
          bun-version: latest
      - run: bun install
      - run: bun run lint
      - run: bun run typecheck
      - run: bun run test
      - run: bun run build
```

### Scripts validados localmente

| Script | Comando | Status |
|---|---|---|
| `lint` | `oxlint` | ✅ solo warnings de Fast Refresh (esperado) |
| `typecheck` | `tsc -b --noEmit` | ✅ |
| `test` | `vitest run` | ✅ 14/14 |
| `build` | `tsc -b && vite build` | ✅ |

### Script agregado a `package.json`

```json
"typecheck": "tsc -b --noEmit"
```

---

## 4. Docker Compose — orquestación completa

### Archivos creados (pendiente de commit)

#### `frontend-react/Dockerfile` (multi-stage)

```dockerfile
FROM oven/bun:1 AS build
WORKDIR /app
COPY package.json bun.lock ./
RUN bun install
COPY . .
RUN bun run build

FROM nginx:alpine
COPY --from=build /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

#### `nginx.conf`

```nginx
server {
    listen 80;
    server_name localhost;
    root /usr/share/nginx/html;
    index index.html;
    location / {
        try_files $uri $uri/ /index.html;
    }
    location /api/ {
        proxy_pass http://backend:8080/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

#### `docker-compose.yml` (en raíz del repo)

```yaml
services:
  db:
    image: postgres:16-alpine
    networks: [ barber-net ]
    volumes: [ pgdata:/var/lib/postgresql/data ]
    env_file: [ .env ]
    restart: unless-stopped

  backend:
    build:
      context: ./backend
      dockerfile: Dockerfile
    depends_on: [ db ]
    networks: [ barber-net ]
    ports: [ "8080:8080" ]
    env_file: [ .env ]
    restart: unless-stopped

  frontend:
    build:
      context: ./frontend-react
      dockerfile: Dockerfile
    depends_on: [ backend ]
    networks: [ barber-net ]
    ports: [ "80:80" ]
    restart: unless-stopped

networks:
  barber-net:
    driver: bridge

volumes:
  pgdata:
```

#### `.env.example`

```env
POSTGRES_DB=barber
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
SPRING_DATASOURCE_URL=jdbc:postgresql://db:5432/barber
SPRING_PROFILES_ACTIVE=prod
JWT_SECRET=generate-a-random-256-bit-secret
ADMIN_SEED_EMAIL=admin@admin.com
ADMIN_SEED_PASSWORD=admin1234
```

### Recomendación: ¿dónde vive docker-compose.yml?

Opción recomendada: **rama `main` en la raíz del repo**. Razones:

| Opción | Problema |
|---|---|
| Rama `infra` | Requiere checkout extra + merge constante |
| Rama `backend` | No tiene `frontend-react/` (eliminado en commit `5ed5d24`) |
| Rama `frontend` | No tiene `backend/` (eliminado de esta rama históricamente) |
| **`main`** (recomendado) | Visible desde cualquier checkout. Infraestructura es estable, no viola "solo código probado en main" |

Para desarrollo local, se necesitan ambos directorios simultáneamente:
- `git worktree add ../Barberia-backend backend` (rama backend en dir paralelo)
- O el clone `Barberia-backend-run/` ya existente
- docker-compose.yml usaría `context: ./backend` si backend/ está presente, o ruta alternativa

---

## Resumen de archivos tocados (solo frontend)

```
NUEVOS (13):
  frontend-react/src/context/AuthContext.tsx
  frontend-react/src/lib/schemas/login-schema.ts
  frontend-react/src/api/auth.ts
  frontend-react/src/pages/Login.tsx
  frontend-react/src/components/ProtectedRoute.tsx
  frontend-react/src/components/admin/AdminLayout.tsx
  frontend-react/src/components/ui/DataTable.tsx
  frontend-react/src/components/ui/StatusBadge.tsx
  frontend-react/src/components/ui/table.tsx
  frontend-react/src/lib/utils.ts
  frontend-react/src/docs/tech-debt-frontend.md
  frontend-react/src/docs/informe-sesion-completo.md
  .github/workflows/frontend-ci.yml

MODIFICADOS (4):
  frontend-react/src/App.tsx          +25 líneas
  frontend-react/package.json         +3 líneas
  frontend-react/bun.lock             +4 líneas
  frontend-react/vitest.config.ts     +9 líneas

NUEVOS (testing + infra):
  frontend-react/vitest.config.ts
  frontend-react/src/tests/setup.ts
  frontend-react/src/tests/turno-schema.test.ts
  frontend-react/src/tests/Reserva.test.tsx
  frontend-react/src/tests/App.test.tsx
  frontend-react/src/tests/msw/handlers.ts
  frontend-react/src/tests/msw/server.ts
  frontend-react/Dockerfile
  docker-compose.yml
  .env.example
  frontend-react/nginx.conf
```

Ningún archivo fuera de `frontend-react/`, `.github/`, raíz del repo, o `.env.example` fue tocado. Backend (`pom.xml`, `SecurityConfig.java`, etc.) no es responsabilidad de esta sesión.
