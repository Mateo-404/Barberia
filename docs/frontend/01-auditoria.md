# Auditoría frontend — Barbería/Matecito

Fecha: 2026-06-24 · Rama: `main` (commit `87548a1`) · Estado local: solo `.gitignore` modificado (corrección de rutas `frontend-react/` → `frontend/`, cambio del usuario, no tocado).

## 1. Baseline verificado

Comandos ejecutados en `frontend/` **antes** de modificar nada:

| Comando | Resultado |
|---|---|
| `bun run lint` (oxlint 1.x) | ✅ pasa, **5 warnings** (ver §2) |
| `bun run build` (`tsc -b && vite build`) | ✅ pasa |
| `npx tsc -b --force` con `strict: true` agregado | ✅ pasa sin errores |
| Tests | ❌ **no existen** (ni scripts ni deps) |
| CI de frontend | ❌ **no existe** (solo `.github/workflows/backend-ci.yml`) |

### Bundle de producción (línea base, antes de cambios)

| Chunk | Raw | Gzip |
|---|---|---|
| `index-t7BtOQkR.js` (entry: React, Router, RHF, Zod, Query, UI) | 436.68 kB | 134.35 kB |
| `KpiCard-*.js` (chunk compartido que arrastra **recharts**) | 376.29 kB | 108.08 kB |
| `TurnosAdmin-*.js` | 43.47 kB | 9.89 kB |
| `EstadisticasAdmin-*.js` | 31.56 kB | 9.36 kB |
| `AdminDashboard-*.js` | 5.24 kB | 1.63 kB |
| CSS | 55.70 kB | 10.11 kB |

No hay medición de Lighthouse todavía (ver `03-plan-priorizado.md`, fase 7).

## 2. Warnings de lint existentes (previos, no regresiones)

- `src/components/ui/button.tsx:54,58` — `react/only-export-components` (variantes exportadas junto al componente).
- `src/context/AuthContext.tsx:107` — `react/only-export-components` (`useAuth` en el mismo archivo del provider).
- `src/pages/Reserva.tsx:76` — `react/incompatible-library` sobre `watch()` de RHF (avisos de React Compiler).
- `src/pages/admin/TurnosAdmin.tsx:367` — `react-hooks/exhaustive-deps`: `turnosDelDia` se recalcula en cada render y es dependencia de un `useMemo` (línea 282, no memoizado).

## 3. Hallazgos

Ordenados por severidad. `Archivos` indica dónde está demostrado el problema.

### Críticos

**H-01 · Reglas de negocio codificadas en el cliente (y divergentes entre sí)**
- `src/pages/Reserva.tsx:16-21` — `TIME_SLOTS` hardcodeado (13:00–19:30, 30 min).
- `src/pages/admin/TurnosAdmin.tsx:25-29` — **la misma lista duplicada**.
- `src/lib/schemas/turno-schema.ts:19-26` — regla comercial en Zod: `h >= 13 && h < 20 && (m === 0 || m === 30)`, con mensaje "entre 13:00 y 20:00".
- Divergencia real: el schema acepta `19:30` como máximo del rango `< 20`, pero el mensaje dice "hasta 20:00" y la lista de slots termina en 19:30. Ningún componente ni schema consulta al backend.
- El backend **no tiene** endpoint de configuración de negocio (verificado sobre `backend/openapi-spec.json`: sólo CRUD + `/turnos/findDateTimes` + `/estadisticas/panel` + paginaciones + `/ping`). Contrato requerido documentado en `04-contrato-configuracion-negocio.md`.

**H-02 · `api-client` descarta el `AbortSignal` del llamador y confunde timeout con cancelación**
- `src/lib/api-client.ts:40-56`: crea su propio `AbortController` y hace `...init, signal: controller.signal`, pisando cualquier `signal` del caller (React Query cancela onUnmount y se pierde).
- Todo `AbortError` se traduce en "Tiempo de espera agotado" aunque lo haya cancelado el caller.
- `src/lib/api-client.ts:66-69`: ante un 401 hace `window.location.href = "/login"` → recarga completa de la SPA, **no limpia la caché de React Query**, y no hay forma de que `AuthContext` se entere. Si la caché contiene turnos/clientes, quedan en memoria.
- No distingue error de red (fetch rechazado) de error de servidor; no hay tipo `network`/`cancelled`/`timeout`.

**H-03 · La caché de datos sensibles no se limpia al cerrar sesión**
- `src/context/AuthContext.tsx:85-89`: `logout()` borra `sessionStorage` pero **no** `queryClient.clear()`.
- `App.tsx:15-21`: `QueryClient` global con `staleTime: 60_000`; tras logout y nuevo login (otro u same), turnos/clientes/estadísticas siguen en memoria hasta 60 s y pueden renderizarse (`TurnosAdmin` usa `useTodosTurnos`).

**H-04 · El admin carga TODOS los turnos y calcula KPIs/filtros en el cliente**
- `src/api/turnos.ts:25-30`: `GET /turnos` sin paginar; `TurnosAdmin.tsx:121-154` filtra, agrupa por día y calcula KPIs sobre el historial completo. El backend sí expone `GET /turnos/{page}/{size}`. Escala mal y hace agua con multi-tenant.

### Altos

**H-05 · Identidad de negocio hardcodeada en componentes reutilizables**
- `Reserva.tsx:222-227`, `AdminLayout.tsx:62-68`, `Login.tsx:74,150`: `alt="TH Barber Club"`, link `https://instagram.com` (genérico, no del negocio), `© 2026 TH Barber Club`. Imposible de rebrandear por tenant.

**H-06 · Query keys planas, sin identidad de negocio ni invalidación por tenant**
- `["servicios"]`, `["fechasOcupadas"]`, `["turnos","todos"]`, `["estadisticas","panel"]` (`src/api/*.ts`). Nada invalida `["fechasOcupadas"]` tras una reserva exitosa → otro usuario ve slots ya tomados (el backend rechaza, pero la UX muestra disponibilidad vieja).

**H-07 · Accesibilidad de los flujos críticos**
- `Reserva.tsx:378-405`: `<label>` sin `htmlFor` ni `id` asociado a los `Input` (labels no asociados).
- Sin `aria-live` para errores de submit (`errors.root`) ni para el cambio de paso.
- `TurnosAdmin.tsx:911-975`: modales "Eliminar turno" y "Mover turno" **sin** `role="dialog"` ni `aria-modal` (el de Editar sí lo tiene, línea 1019); sin foco inicial ni retorno de foco; sin `aria-labelledby`.
- `aria-label="Editar turno"` / `"Eliminar turno"` repetidos idénticos en cada fila (nombre accesible no distinguible para lectores de pantalla).
- `TurnosAdmin.tsx:509-512`: botón "Filtros Avanzados" **muerto** (sin onClick).
- `Login.tsx:130`: checkbox "Recordarme" sin handler (hace nada); `href="#"` en "Términos"/"Privacidad"/"¿Olvidaste tu contraseña?" (enlaces vacíos).
- Tablas sin `<caption>`/scope explícito (menor).

**H-08 · Fechas y zona horaria ambiguas**
- `Reserva.tsx:100`: construye `${fecha}T${hora}:00` **sin offset**; `turno-schema.ts:14` lo compara con `new Date()` (zona del navegador); `format.ts:11-26` formatea con la zona local del usuario. Si el negocio tiene otra zona horaria, el wizard muestra y envía horas desplazadas. Ver `04-contrato-configuracion-negocio.md` (requiere `timeZone` del negocio).

**H-09 · Chunk de recharts descargado por módulos que no lo usan**
- `components/admin/ClienteDetalle.tsx:8` importa `recharts` y se importa estáticamente desde `TurnosAdmin` → chunk compartido de 376 kB (108 kB gzip) viaja con la vista de turnos (línea base §1).

### Medios

**H-10 · Sesión JWT en `sessionStorage` accesible por JS**
- `AuthContext` + `api-client` leen `auth_token` desde `sessionStorage`. Aceptable hoy (backend actual), pero la migración a cookies `HttpOnly`/`SameSite` debe definirse con el backend Go (fase 5 del plan). La expiración se chequea decodificando el payload (`isTokenExpired`), no hay renovación.

**H-11 · Manejo de errores inconsistente en la UI**
- No hay manejo de 409/400 de concurrencia distinto al resto (el mensaje crudo del backend se muestra tal cual en `Reserva.tsx:203`); no hay retries diferenciados (default de TanStack Query = 3 reintentos incluso en 4xx, salvo `useServicios` que pone `retry: 1`).

**H-12 · Duplicación de KPI card**: `AdminDashboard.tsx:8-26` define su propio `KpiCard` (memo) mientras existe `components/admin/KpiCard.tsx` con otra API.

**H-13 · Sin `typecheck` ni `test` como scripts independientes**: `build` hace `tsc -b` implícitamente; no hay forma de correr sólo tipos ni tests en CI por separado.

## 4. Riesgos de seguridad (frontend)

1. Token en `sessionStorage` + redirección por `window.location` sin limpieza de caché (H-02/H-03/H-10).
2. CSP presente en `frontend/index.html` (`default-src 'self'`,
   `connect-src 'self' %VITE_CSP_CONNECT_SRC%`; el placeholder lo resuelve
   `vite.config.ts` en build-time vía `VITE_CSP_CONNECT_SRC`); ver que el
   despliegue final mantenga `https` y agregue headers de servidor
   (`X-Frame-Options`, `Referrer-Policy`) — depende del host, documentado como
   pendiente.
3. No se detectaron secretos ni credenciales en el código del frontend; `.env` fuera del repo; `JWT_SECRET` sólo en backend.
4. Datos de turnos/clientes renderizados vía JSX (React escapa); no se encontró `dangerouslySetInnerHTML`.
5. La autorización real es del backend: `ProtectedRoute` sólo UX (correcto, se documenta).

## 5. Lo que está bien (y se conserva)

- React 19 + Vite 8 + TS estricto (habilitado y verificado limpio en esta auditoría) + Tailwind 4 + TanStack Query + RHF/Zod: sin razones para migrar.
- Code splitting de páginas admin con `lazy()` y `Suspense`.
- CSP meta en `index.html`, `rel="noopener noreferrer"` en externos.
- Atajos de teclado bien encapsulados (`useHotkeys` ignora inputs y diálogos).
- Tipos generados desde OpenAPI con script determinista (`bun run types`).
- Build rápido (1.1 s) y CSS contenido.

## 6. Limitaciones de esta auditoría

- **No** se ejecutó Lighthouse ni profiling de navegador (pendiente fase 7).
- **No** se levantó el backend Java en esta sesión: los tests de flujo usan `fetch` simulado (MSW-free, stub directo) y no validan el contrato vivo.
- **No** se revisó accesibilidad con lectores de pantalla reales; sólo inspección de código + axe (automatizada, fase 6).
