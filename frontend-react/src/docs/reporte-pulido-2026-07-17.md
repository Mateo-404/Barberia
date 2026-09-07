# Reporte de Pulido — Frontend React (TH Barber Club)

**Fecha:** 17 jul 2026
**Alcance:** Pulido máximo de `frontend-react/` tras las fases de login, reserva y theming.
**Estado:** typecheck ✅ · build ✅ (0 errores) · lint ✅ (solo warnings pre-existentes de fast-refresh)

---

## 1. Funcionales corregidos

### F1 — Dashboard de administración real
Antes era un placeholder ("próximamente"). Ahora `src/pages/admin/AdminDashboard.tsx`
consume los endpoints reales del backend:

- `GET /estadisticas/panel` → KPIs (turnos hoy, ingresos del mes, total clientes, turnos registrados)
- `GET /turnos` → tabla de últimos 10 turnos con badge de estado (Hoy / Agendado)

Incluye:
- 4 tarjetas KPI con valores formateados en ARS (`src/lib/format.ts`)
- Ranking de servicios más solicitados con barras de progreso relativas
- Tabla responsive (scroll horizontal en móvil) con fecha/hora formateadas en `es-AR`
- Estados de loading (skeletons `animate-pulse`) y error dedicados

### F2 — Validación de expiración de JWT al cargar
`AuthContext.tsx` ahora decodifica el claim `exp` del token en `loadAdmin`/`loadToken`
y limpia la sesión si está expirado. Antes un token muerto en `sessionStorage` dejaba
al usuario "logueado" en pantalla aunque el backend lo rechazara.

### F3 — Copy específico de error de login
`Login.tsx` distingue `401 /errors/invalid-credentials` → *"Email o contraseña incorrectos."*
y timeout de red (`status 0`) → mensaje de conexión. Antes mostraba el `detail` crudo.

### F4 — Estado de error en la carga de servicios
`Reserva.tsx` ahora muestra un bloque de error si `useServicios` falla, en lugar de
confundirlo con "No hay servicios disponibles". Se agregó `retry: 1` al query.

### F5 — Timeout y aborto de requests
`api-client.ts` usa `AbortController` con timeout de 15s. Un backend colgado ya no
cuelga la UI para siempre: lanza `ApiError(0, "Tiempo de espera agotado", ...)`.

---

## 2. Diseño / UX pulido

| ID | Cambio | Archivo |
|----|--------|---------|
| D1 | `index.html`: `lang="es"`, `<title>` real, `theme-color` + meta description | `index.html` |
| D2 | Toggle mostrar/ocultar contraseña con `aria-label` y foco visible | `Login.tsx` |
| D3 | `focus-visible:ring-2 ring-focus-ring` en flecha-atrás, links y service-cards | `AdminLayout.tsx`, `Reserva.tsx`, `Login.tsx` |
| D4 | Eliminados `DataTable.tsx` y `table.tsx` (dead code, no usados) + dep `@tanstack/react-table` | `package.json` |
| D6 | Empty-states con microcopy en dashboard (servicios/turnos vacíos) | `AdminDashboard.tsx` |

---

## 3. Sistema de variables (reciclables)

Todo color vive en `:root` (`src/index.css`). Tailwind resuelve utilidades semánticas
(`bg-primary`, `text-muted-foreground`, `border-border`, `ring-focus-ring`, etc.) a esas
variables. **Cero hex hardcodeados en componentes.** Variables agregadas en esta pasada:

- `--nav-height: 4.5rem` — altura unificada de navbar (reemplaza `10vh`)
- `--focus-ring` — anillo de foco naranja translúcido
- `.page-enter` — animación de entrada (fade + slide-up, respeta `prefers-reduced-motion`)

---

## 4. Archivos nuevos / modificados

**Nuevos**
- `src/pages/admin/AdminDashboard.tsx`
- `src/api/estadisticas.ts`
- `src/lib/format.ts`

**Modificados**
- `src/App.tsx` — usa `AdminDashboard` real
- `src/context/AuthContext.tsx` — validación de exp
- `src/lib/api-client.ts` — AbortController + timeout
- `src/pages/Login.tsx` — toggle password, error copy, focus rings, meta
- `src/pages/Reserva.tsx` — estado error servicios, focus rings
- `src/components/admin/AdminLayout.tsx` — focus ring flecha
- `src/api/servicios.ts` — `retry: 1`
- `src/index.html` — lang/title/meta
- `src/index.css` — `--nav-height`, `--focus-ring`, `.page-enter`
- `package.json` — removido `@tanstack/react-table`

**Eliminados**
- `src/components/ui/DataTable.tsx`
- `src/components/ui/table.tsx`

---

## 5. Pendiente / fuera de alcance

- F6: mapear error 409 "slot tomado" a copy amigable (requiere confirmar el `type` del backend).
- Wire de "Recordarme" / "¿Olvidaste tu contraseña?" (acordado como placeholders).
- Tests E2E (MSW) — existen los handlers pero no corren en CI todavía.
