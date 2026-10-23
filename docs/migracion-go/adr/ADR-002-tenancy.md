# ADR-002 — Multi-tenancy: shared schema con `tenant_id`

- **Estado**: aceptado (2026-10-21)
- **Contexto**: Barbería será SaaS en el ecosistema Matecito; hoy es instalación
  única. Se necesita un límite de seguridad por tenant sin sobre-ingeniería.
- **Decisión**: columna `tenant_id` (FK → `tenants`) en `clientes`, `servicios`,
  `turnos`, `administradores`. Índice `(tenant_id, …)` en las consultas calientes
  y `UNIQUE (tenant_id, fecha_hora)` en turnos.
  - **Ámbito autenticado**: el JWT emite claim `tenant` desde la fila del admin en
    el login; toda query filtra por ese valor. Un `tenant_id` enviado por el
    cliente jamás autoriza (no existe header confiable `X-Tenant-ID`).
  - **Rutas públicas** (`POST /turnos`, `GET /turnos/findDateTimes`,
    `GET /servicios*`): el tenant se resuelve por `Host` contra `tenants.host`
    (mapeo DNS controlado por operador), con fallback `DEFAULT_TENANT_SLUG` para
    despliegue single-tenant (modo actual, comportamiento idéntico al Java).
- **Alternativas descartadas**: esquema/base por cliente (operación cara, pools por
  cliente, migraciones N veces); RLS de PostgreSQL (requiere `SET LOCAL app.tenant`
  por transacción y disciplina con el pool — superficie de error mayor sin beneficio
  hoy; se reevaluará si hay un requisito de aislamiento regulatorio).
- **Consecuencias**: toda query nueva debe llevar `tenant_id` (reviwed por convención
  y por los tests de aislamiento `internal/integration`); el "registro de nuevos
  negocios del SaaS" queda fuera de alcance (sin flujo comercial aún) — crear
  tenants es una operación administrativa documentada (INSERT en `tenants`).
- **Datos existentes**: `000002` crea el tenant `default` y asigna todo.
