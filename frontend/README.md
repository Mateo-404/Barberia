# Frontend — App de reservas (React)

Aplicación moderna de reservas de la barbería: wizard de reserva para clientes y panel de administración (JWT). Es la app activa del proyecto; el sitio HTML/CSS/JS vanilla de la raíz del repo queda solo como referencia histórica.

> Guía general del proyecto: [README raíz](../README.md) · [Índice de docs](../docs/README.md) · [Atajos de teclado](../docs/atajos-teclado.md)

## Stack

- **React 19** + **TypeScript 7** + **Vite 8**
- **Tailwind CSS v4** (+ shadcn-style components en `components/ui`)
- **TanStack Query** (estado de servidor) · **React Hook Form + Zod** (formularios/validación)
- **react-router-dom 7** (routing)
- Package manager: **Bun** (hay `bun.lock`; también funciona npm, pero el repo está estandarizado en Bun)

## Requisitos

- [Bun](https://bun.sh) ≥ 1.4
- El backend Go corriendo en `http://localhost:8080` (ver [backend-go/README.md](../backend-go/README.md); lo más rápido: `docker compose up` desde la raíz)

## Arranque rápido

```bash
bun i                       # instalar dependencias
cp .env.example .env        # VITE_API_BASE_URL=http://localhost:8080
bun dev                     # http://localhost:5173
```

## Scripts

| Comando | Qué hace |
|---|---|
| `bun dev` | Dev server de Vite en el puerto 5173 |
| `bun run build` | Typecheck (`tsc -b`) + build a `dist/` |
| `bun run preview` | Sirve el build de producción en local |
| `bun run lint` | Lint con Oxlint |
| `bun run typecheck` | Typecheck (`tsc -b`) |
| `bun run test` | Tests unitarios (Vitest + Testing Library) |
| `bun run test:e2e` | Tests e2e (Playwright sobre `vite preview`, API mockeada) |
| `bun run types` | Regenera `src/types/api.ts` desde `../backend-go/api/openapi.json` |

## Variables de entorno

| Variable | Descripción | Default |
|---|---|---|
| `VITE_API_BASE_URL` | URL base del API backend | `http://localhost:8080` |
| `VITE_TENANT_SLUG` | Tenant usado como clave de caché/tema del frontend | `default` |
| `VITE_CSP_CONNECT_SRC` | Orígenes **extra** para el `connect-src` de la CSP (separados por espacio) | vacío |

Ver `frontend/.env.example`. `.env` está en `.gitignore`.

### CSP (`connect-src`)

`frontend/index.html` declara una CSP con un placeholder `__CSP_CONNECT_SRC__`
que `vite.config.ts` reemplaza en build-time por:

- el **origen de `VITE_API_BASE_URL`** (el mismo que usa `api-client.ts`, así la
  CSP nunca bloquea al backend configurado);
- en **dev**, además `ws://localhost:5173` (HMR de Vite);
- los orígenes extra de `VITE_CSP_CONNECT_SRC`, si los hay.

Para producción basta con `VITE_API_BASE_URL=https://api.mibarberia.com bun run build`.

## Rutas

| Ruta | Vista | Acceso |
|---|---|---|
| `/` | Wizard de reserva (cliente) | Público |
| `/login` | Login de administrador | Público |
| `/admin` | Panel con KPIs y últimos turnos | Protegida (JWT) |
| `/admin/turnos` | Gestión de turnos (lista/calendario) | Protegida (JWT) |
| `/admin/estadisticas` | Estadísticas y gráficos | Protegida (JWT) |
| `/estadisticas/panel` | Redirect legacy → `/admin/estadisticas` | — |

## Estructura

```
src/
├── api/            # funciones de acceso a la API (auth, turnos, servicios, estadísticas)
├── components/
│   ├── admin/      # AdminLayout (nav + atajos), ClienteDetalle, KpiCard
│   ├── ui/         # primitivas de UI (button, card, table…)
│   ├── ProtectedRoute.tsx
│   └── ShortcutsHelp.tsx   # panel de ayuda de atajos (Shift+?)
├── context/        # AuthContext (admin + token en sessionStorage)
├── hooks/          # useHotkeys, useShortcutsRegistry
├── lib/            # api-client (fetch con Bearer), format, utils
│   └── schemas/    # schemas Zod (login, turno)
├── pages/          # Reserva, Login
│   └── admin/      # AdminDashboard, TurnosAdmin, EstadisticasAdmin
└── types/          # api.ts — GENERADO, no editar a mano
```

## Autenticación

1. `POST /administradores/login` devuelve `LoginResponseDTO { id, nombre, apellido, email, token }`.
2. `AuthContext` guarda el admin y el token en `sessionStorage` (claves `admin`, `auth_token`).
3. `lib/api-client.ts` agrega `Authorization: Bearer <token>` a cada request.
4. Ante un `401 /errors/invalid-token` se limpia la sesión y se redirige a `/login`.
5. `ProtectedRoute` envuelve todas las rutas `/admin/*`.

## Tipados generados desde el backend

`src/types/api.ts` lo genera `openapi-typescript` a partir de `backend-go/api/openapi.json` (única fuente de verdad, mantenida junto al backend Go). **No editar a mano.** Si cambia la API, actualizar el spec en el mismo cambio y regenerar:

```bash
bun run types
```

> Nota: el script usa `bunx` con `typescript@5.9` aislado porque `openapi-typescript` todavía no es compatible con el TypeScript 7 del proyecto.

## Atajos de teclado

Hay atajos en todo el sitio (wizard de reserva, panel admin, turnos, estadísticas). Se listan con **`Shift + ?`**. Referencia completa y guía para agregar nuevos: [docs/atajos-teclado.md](../docs/atajos-teclado.md).

## Calidad de código

- Lint: `bun run lint` (Oxlint, config en `.oxlintrc.json`).
- Tests: Vitest (`src/**/__tests__`) y Playwright (`e2e/`).
- CI: `.github/workflows/frontend-ci.yml` corre lint, typecheck, tests, build y e2e en cada cambio de `frontend/`.

## Deploy

Hoy la app React **no tiene despliegue automático**: se ejecuta en local contra el backend. El sitio publicado en GitHub Pages ([demo](https://mateo-404.github.io/Barberia/)) es el sitio estático legacy de la rama `frontend`.
