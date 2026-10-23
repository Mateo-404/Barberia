# Plan priorizado — frontend Barbería/Matecito

Fases independientes y revisables. Estado: ⬜ pendiente · 🟡 en curso · ✅ implementado (con verificación) · 🔗 bloqueado por contrato externo.

| # | Fase | Estado | Dónde |
|---|---|---|---|
| 1 | Auditoría, baseline, riesgos | ✅ | `01-auditoria.md` |
| 2 | Tests + CI | 🟡 | `frontend/src/**/__tests__`, `.github/workflows/frontend-ci.yml` |
| 3 | Cliente HTTP y contrato tipado | ⬜ | `src/lib/api-client.ts` |
| 4 | Flujo de reservas + preparación multi-tenant | ⬜ | `Reserva.tsx`, `business-config.ts`, `query-keys.ts` |
| 5 | Seguridad de sesión (diseño, coord. con Go) | 🔗 | `05-sesion-cookies.md` |
| 6 | UX y accesibilidad | ⬜ | diálogos, labels, live regions |
| 7 | Rendimiento medido | ⬜ | bundle baseline en `01-auditoria.md` §1 |
| 8 | SEO público y documentación | ⬜ | `06-seo-publico.md` |

## Fase 2 — Tests y CI

- Dependencias nuevas (avisadas): `vitest`, `jsdom`, `@testing-library/react`, `@testing-library/user-event`, `@testing-library/jest-dom`, `axe-core`, `@playwright/test` (E2E).
- Scripts: `test`, `test:watch`, `test:e2e`, `typecheck`, `lint`, `build`.
- Cobertura mínima exigida (misión §4): reserva pública (carga/estados de servicios, validaciones, selección fecha/hora, envío, doble clic, timeout, error de servidor, 409 de concurrencia, reset, recuperación) y acceso admin (login válido/inválido, token ausente/expirado, protección de rutas, logout con limpieza de caché, 401, acciones admin).
- CI reproducible: `bun install --frozen-lockfile` → lint → typecheck → test → build → regenerar tipos OpenAPI y verificar que no hay diffs.

## Fase 3 — Cliente HTTP

- `AbortSignal.any` respetando al caller; kinds distinguishibles (`http/network/timeout/cancelled`).
- 401 ⇒ evento único ⇒ AuthContext limpia sesión **y caché** ⇒ router navega (sin `window.location`).
- No reintentar 4xx en TanStack Query.

## Fase 4 — Reservas configurables y multi-tenant

- Eliminar `TIME_SLOTS` duplicados y la regla comercial del schema (la autoridad es el backend); centralizar en `lib/business-config.ts` mientras no exista el endpoint (contrato en `04-...`).
- `query-keys.ts` jerárquico con scope de negocio; invalidar `disponibilidad` tras reserva; `queryClient.clear()` en logout/cambio de tenant.
- Quitar branding hardcodeado de componentes reutilizables (configuración de marca por tenant).

## Fase 5 — Sesión (coordinación con backend Go)

- 🔗 Requiere contrato: cookie `HttpOnly`/`Secure`/`SameSite` o BFF, CORS con credenciales, CSRF, renovación. Diseño y tests preparados sin romper el JWT actual → `05-sesion-cookies.md`.

## Fase 6 — UX y accesibilidad (WCAG 2.2 AA)

- `Dialog` accesible (role/aria-modal/foco/Escape) reutilizable; labels asociados (`htmlFor`/`id`); `aria-live` para errores y cambio de estado asíncrono; nombres de acciones distinguibles; eliminar controles muertos ("Filtros Avanzados", "Recordarme", `href="#"`).
- axe en tests de componentes; checklist manual de teclado documentado.

## Fase 7 — Rendimiento medido

- Línea base: `01-auditoria.md` §1.
- Separar recharts de TurnosAdmin (chunk de 108 kB gzip que no usa esa vista).
- Lighthouse en `vite preview` antes/después; sin datos de campo no se afirman Core Web Vitals.

## Fase 8 — SEO y docs

- `06-seo-publico.md`: estrategia por superficie, metadata actual, robots/sitemap cuando existan páginas públicas.

## Deuda técnica explícita pendiente

1. Paginación server-side en TurnosAdmin (H-04) — requiere acordar filtros con backend Go.
2. Cookies de sesión (H-10) — bloqueada por backend.
3. Zona horaria del negocio (H-08) — bloqueada por `04-contrato-configuracion-negocio.md`.
4. React Compiler — evaluar recién con baseline y tests (lint ya emite warnings).
5. Datos de campo (RUM) — no existen; sólo métricas de laboratorio.
