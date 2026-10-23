# Arquitectura objetivo — frontend Barbería/Matecito

Principio rector: **evolucionar sin reescribir**. Se conservan React 19, Vite 8, TypeScript estricto, Tailwind 4, TanStack Query, React Hook Form + Zod, OpenAPI como contrato y Bun como package manager. No se agregan Redux/Zustand, ni un segundo cliente HTTP, ni una segunda librería de formularios.

## 1. Estructura modular por funcionalidad

```
src/
├── app/                  # QueryClient, rutas, providers (hoy: App.tsx)
├── features/
│   ├── reserva/          # wizard público: páginas, hooks, schemas, tests
│   ├── auth/             # AuthContext, ProtectedRoute, login
│   └── admin/            # turnos, estadísticas, dashboard, componentes admin
├── components/ui/        # primitivas genéricas (button, input, dialog…)
├── lib/
│   ├── api-client.ts     # ÚNICO punto de acceso HTTP
│   ├── query-keys.ts     # claves jerárquicas tipadas
│   ├── business-config.ts# configuración del negocio (interina, ver §4)
│   └── format.ts
└── types/api.ts          # GENERADO desde OpenAPI (nunca a mano)
```

Migración incremental: se mueve lo que se toca; no se hacen grandes reubicaciones de un solo golpe (regla del repo: explicar el plan antes de reorganizar).

## 2. Cliente HTTP único

`lib/api-client.ts` queda como única estrategia de acceso, con:

- `ApiError` extendido con `kind: "http" | "network" | "timeout" | "cancelled"`, `status`, `title`, `detail`, `type` (RFC 9457 del backend).
- Composición de `AbortSignal`: `AbortSignal.any([caller, timeout])` → el caller (React Query) puede cancelar y **no** se reporta como timeout.
- Manejo de 204/vacío/no-JSON sin romper.
- 401 ⇒ emite un evento (`auth:unauthorized`) que `AuthContext` consume: limpia sesión **y** `queryClient.clear()`; la navegación la hace React Router (sin `window.location`).
- Sin logs con datos personales; mensajes aptos para usuario, `type`+`status` para diagnóstico.
- Los tipos de request/response salen de `components["schemas"]` (OpenAPI). Los genéricos no validan runtime: la validación runtime de payloads críticos se hace con Zod en el boundary si hace falta.

## 3. Estado remoto: claves jerárquicas y tenant-aware

```ts
// lib/query-keys.ts
export const queryKeys = {
  servicios: (tenant: string) => ["servicios", tenant] as const,
  disponibilidad: (tenant: string) => ["disponibilidad", tenant] as const,
  turnos: {
    all:      (tenant: string) => ["turnos", tenant] as const,
    list:     (tenant: string, filtros: FiltrosTurno) => ["turnos", tenant, "lista", filtros] as const,
    pagina:   (tenant: string, page: number, size: number) => [..., "página", page, size] as const,
  },
  estadisticas: {
    panel: (tenant: string) => ["estadisticas", tenant, "panel"] as const,
  },
}
```

- `tenant` es el **slug/identificador de negocio** que resuelve el contexto público; es un *scope de datos*, **nunca** una autorización (la API Go debe imponer el aislamiento).
- Al logout o cambio de tenant: `queryClient.clear()` (o `removeQueries({ prefix })`) antes de que la nueva sesión pinte algo.
- Invalidaciones: `["turnos", tenant]` tras mutaciones; `["disponibilidad", tenant]` tras reserva exitosa (evita slots obsoletos).
- Retries: sólo en errores transitorios (red/timeout/5xx); **no** en 4xx. `staleTime` por recurso (disponibilidad corta, catálogo más largo).
- Paginación server-side en el panel (`GET /turnos/{page}/{size}`): el front no descarga historiales ilimitados.

## 4. Configuración del negocio (preparación multi-tenancy)

Interina (hoy): un único módulo `lib/business-config.ts` con la ventana de atención, intervalo y branding, **una sola vez**, consumido por reserva y admin (elimina la duplicación H-01) y con TODO marcado como "reemplazar cuando exista el endpoint". Ver `04-contrato-configuracion-negocio.md`.

Objetivo: `GET /negocios/{slug}/configuracion` devuelve horarios, servicios, zona horaria y branding; el frontend renderiza lo que devuelva y no guarda reglas comerciales propias. El backend sigue siendo la autoridad en la confirmación de reservas.

Identidad de producto (Matecito) vs identidad del negocio (barbería) se separan: los componentes de UI reutilizables no conocen logotipos ni redes de ningún cliente.

## 5. Sesión

- Hoy: JWT en `sessionStorage` (compatible con el backend Java actual).
- Diseño preparado para cookies `HttpOnly` + `Secure` + `SameSite=Lax`/BFF: `api-client` no lee storage directamente en más de un punto (`getToken()` único), y el manejo de 401 pasa por un solo evento ⇒ cambiar el transporte no toca las páginas. Requiere contrato con Go (CORS con credenciales, CSRF, ciclo de vida). Documentado en `05-sesion-cookies.md`.

## 6. Diseño y accesibilidad

- Tokens existentes en `index.css` se mantienen como fuente única (color/tipografía/espacio/radio/sombra/estado).
- Componentes uniformes: `Dialog` con `role="dialog"`, `aria-modal`, foco atrapado y `Escape`; `ErrorSummary` accesible; estados loading/empty/error consistentes.
- WCAG 2.2 AA como objetivo verificable: axe en CI + revisión manual de teclado en los flujos de reserva y login.
- Mobile-first ya parcialmente presente (tabla ↔ tarjetas); se conserva.

## 7. Rendimiento

- Mantener `lazy()` por ruta; **separar recharts** de los módulos que no grafican (dynamic import de `ClienteDetalle` o mover gráficos detrás de `EstadisticasAdmin`).
- Presupuesto objetivo inicial: entry < 150 kB gzip (línea base: 134 kB entry + 108 kB de recharts arrastrados).
- React Compiler: evaluar **después** de tener tests y baseline (los warnings `react/incompatible-library` de `watch()` ya aparecen en lint).
- Medición: Lighthouse (lab) en `vite preview`, registrando antes/después. Sin datos de campo, no se afirman Core Web Vitals.

## 8. SEO / público

- El panel NO migra a Astro ni recibe SSR.
- Estrategia por superficie (detallada en `06-seo-publico.md`): landing Matecito y páginas de negocio ⇒ contenido indexable (evaluar Astro sólo si aparecen páginas públicas reales); app admin autenticada ⇒ `noindex`.
- Hoy el sitio público es un SPA de un solo wizard: se documenta qué aplica (title/description/canonical/robots/sitemap cuando existan páginas) sin infraestructura nueva sin necesidad.

## 9. CI

- `.github/workflows/frontend-ci.yml`: install reproducible desde `bun.lock` (`bun install --frozen-lockfile`) → lint → typecheck → tests (unit/component + a11y con axe) → build → generación y diff del contrato OpenAPI → E2E Playwright en job separado/condicionado → `bun audit` de dependencias.
