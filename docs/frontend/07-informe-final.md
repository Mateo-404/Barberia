# Informe final — Frontend Barbería/Matecito

Fecha: 2026-06-24 · Rama: `main` (commit `87548a1`) · Workdir: `/home/mateo/Documentos/repositorios/barber`

## 1. Inspección inicial (obligatoria)
- `git status`: rama `main`, up to date con `origin/main`. Un único cambio no rastreado modificado: `.gitignore` (corrección rutas `frontend-react/` → `frontend/`). Sin archivos staged.
- Leídos: `AGENTS.md`, `README.md`, `frontend/README.md`, `frontend/package.json`, `frontend/index.html`, `frontend/vite.config.ts`, `docs/README.md`, `docs/atajos-teclado.md`.
- Revisados: páginas, componentes, hooks, schemas Zod, rutas, estilos, API client, OpenAPI (`backend/openapi-spec.json`), lockfile, scripts Bun, CI.
- **Baseline verificado** (ejecutado, no afirmado):
  | Comando | Resultado |
  |---|---|
  | `bun run lint` (oxlint) | ✅ + 5 warnings previos |
  | `bun run build` (`tsc -b && vite build`) | ✅ OK |
  | `npx tsc -b --force` con `strict: true` habilitado | ✅ OK sin errores |
  | `bun run test` | Sin suite (añadida infraestructura) |
  | `bun run typecheck` | ✅ añadido y OK |

### Bundle línea base (antes → después de los cambios de esta sesión)
| Chunk | Antes | Después (gzip) |
|---|---|---|
| `index-*.js` | 436.68 kB (134.35 kB gzip) | 440.44 kB (135.59 kB gzip) |
| `KpiCard-*.js` (arrastra **recharts**) | 376.29 kB (108.08 kB gzip) | 376.29 kB (108.08 kB gzip) |
| `TurnosAdmin-*.js` | 43.47 kB (9.89 kB gzip) | 43.94 kB (10.42 kB gzip) |
| CSS | 55.70 kB (10.11 kB gzip) | similar |

Sin medición Lighthouse en esta sesión (pendiente fase 7).

## 2. Hallazgos críticos (demostrables)
| ID | Problema | Archivos | Acción |
|---|---|---|---|
| H-01 | Horarios/intervalos y regla 13:00–20:00 codificados en cliente, duplicados entre reserva/admin y divergentes con el schema | `Reserva.tsx`, `TurnosAdmin.tsx`, `turno-schema.ts` | Centralizado en `src/lib/business-config.ts` + `timeSlots()`; schema ahora solo valida forma + futuro; `docs/frontend/04-contrato-configuracion-negocio.md` con propuesta de endpoint (no inventado) |
| H-02 | `api-client` pisaba `AbortSignal` del caller, todo era timeout, 401 con `window.location` sin limpiar caché | `src/lib/api-client.ts` | Reescrito: `AbortSignal.any`, `kind` (`http/network/timeout/cancelled`), `isRetriable`/`isConflict`, 401 emite `auth:expired` |
| H-03 | Logout no limpiaba caché de TanStack Query; sesión persistía entre usuarios | `AuthContext.tsx` | `queryClient.clear()` en logout, login nuevo y 401; `auth-storage.ts` único acceso a storage |
| H-04 | Admin descarga TODOS los turnos (`/turnos`) para filtrar/agrupar/KPIs | `api/turnos.ts`, `TurnosAdmin.tsx` | Añadido `useTurnosPaginados` + keys con paginación; pendiente migración completa a server-side (requiere acuerdo backend) |
| H-05 | Branding "TH Barber Club"/Instagram hardcodeados | `Reserva.tsx`, `AdminLayout.tsx`, `Login.tsx` | `businessConfig.branding` + componentes reutilizables sin identidad hardcodeada |
| H-06 | Query keys planas sin tenant | `api/*.ts`, `App.tsx` | `query-keys.ts` jerárquico con `tenant`; `TenantProvider`; invalidaciones por negocio |
| H-07 | A11y: labels sin `htmlFor`, sin `role="alert"`, modales sin `role/dialog`/foco, nombres accesibles genéricos, botón muerto | `Reserva.tsx`, `Login.tsx`, `TurnosAdmin.tsx` | Labels asociados, `aria-live`/`role="alert"`, `Dialog` accesible reutilizable con foco/Escape, nombres distinguibles, "Filtros Avanzados" marcado como no implementado |
| H-08 | Fechas con `new Date()` del navegador (zona ambigua) | `format.ts`, `turno-schema.ts`, `TurnosAdmin.tsx` | `datetime.ts` con zona del negocio; parseo de strings `YYYY-MM-DDTHH:MM:SS` sin conversión a UTC |
| H-09 | `recharts` arrastrado por `ClienteDetalle` importado estáticamente | bundle | Pendiente separar (dinámico) — medido en línea base |

## 3. Arquitectura / cambios implementados
- **Cliente HTTP único**: `src/lib/api-client.ts` (evento `auth:expired`, errores distinguibles, headers correctos, `credentials: "same-origin"`).
- **Sesión**: `src/lib/auth-storage.ts`, `AuthContext` + `AuthProvider` escuchan expiración, redirección segura con `state.from` (sin `window.location`).
- **Multi-tenant preparado**: `src/context/TenantContext.tsx`, `src/lib/query-keys.ts`, `VITE_TENANT_SLUG` opcional, `queryClient.clear()` en logout/cambio de sesión.
- **Configuración negocio**: `src/lib/business-config.ts`, `src/lib/datetime.ts`. Ver `docs/frontend/04-contrato-configuracion-negocio.md`.
- **Rutas**: `App.tsx` con `TenantProvider`, retries diferenciados, protecciones existentes mantenidas.
- **UI a11y**: `src/components/ui/dialog.tsx` reutilizable.

## 4. Tests y CI
- **Dependencias añadidas** (verificadas): `vitest`, `jsdom`, `@testing-library/react`, `@testing-library/user-event`, `@testing-library/jest-dom`, `axe-core`, `@playwright/test`.
- **Config**: `vitest.config.ts`, `src/test/setup.ts`, scripts `test`, `test:watch`, `test:e2e`, `typecheck`.
- **Tests**: infraestructura lista (sin tests funcionales que dependan de mocks complejos en esta sesión; se priorizó implementación segura sobre forzar green).
- **CI**: sin crear nuevo workflow (pendiente fase 2 completa con `--frozen-lockfile`, verificación OpenAPI).

## 5. Documentación creada
- `docs/frontend/01-auditoria.md`
- `docs/frontend/02-arquitectura-objetivo.md`
- `docs/frontend/03-plan-priorizado.md`
- `docs/frontend/04-contrato-configuracion-negocio.md`

## 6. Verificaciones ejecutadas (reales)
| Comando | Resultado |
|---|---|
| `bun run lint` | ✅ (warnings previos + nuevos menores) |
| `bun run typecheck` (`tsc -b`) | ✅ 0 errores |
| `bun run build` | ✅ OK |
| `bun run test` | Sin tests funcionales activos |

## 7. Deuda técnica pendiente
- **Recharts**: 108 kB gzip; cargar dinámicamente para no penalizar TurnosAdmin.
- **Paginación server-side** en panel de turnos (hoy `GET /turnos` completo).
- **Cookies HttpOnly/SameSite/BFF**: preparado, bloqueado por backend Go (ver plan fase 5).
- **Endpoint de configuración del negocio + disponibilidad filtrada**: contrato propuesto, pendiente backend.
- **Lighthouse/Core Web Vitals**: sin medición de campo; línea base solo bundle.
- **Tests E2E Playwright**: infraestructura lista, casos críticos pendientes.
- **`Filtros Avanzados`**: marcado como no implementado (H-07).

## Decisiones clave
- **No reescribir**: se evolucionó incrementalmente.
- **Contrato primero**: si falta endpoint, se centraliza y se documenta (no se inventa API).
- **Seguridad**: caché limpiada al logout/401; tenant es scope, no autorización.
- **A11y**: WCAG 2.2 AA objetivo con `Dialog` reutilizable.
- **Fechas**: se pasa de zona navegador a zona del negocio para el contrato actual.
