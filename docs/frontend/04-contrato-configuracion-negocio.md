# Contrato requerido: configuración pública del negocio

Estado: **propuesta, aún no implementada por el backend** (verificado 2026-06-24 sobre `backend/openapi-spec.json`: no existe ningún endpoint de configuración/horarios de negocio).

El frontend **no inventó** este endpoint ni lo simula: mientras no exista, los valores viven centralizados en `frontend/src/lib/business-config.ts` con marcas de reemplazo. Al aparecer el endpoint, ese módulo pasa a consumir el contrato de abajo y se elimina la configuración local.

## Endpoint propuesto

```
GET /publico/negocios/{slug}/configuracion
```

- `slug`: identificador público del negocio en la plataforma (resolución de tenant). **No es autorización**: la API debe validar igual el acceso a los datos.
- Sin autenticación. Cacheable corto (`Cache-Control: public, max-age=60`).

### Respuesta (OpenAPI sketch)

```yaml
type: object
required: [slug, nombre, zonaHoraria, diasAtencion, intervaloMinutos, servicios, marca]
properties:
  slug: { type: string }
  nombre: { type: string }                    # "TH Barber Club"
  zonaHoraria: { type: string }               # IANA, ej. "America/Argentina/Buenos_Aires"
  diasAtencion:                              # días laborables del negocio
    type: array
    items:
      type: object
      required: [dia, desde, hasta]
      properties:
        dia: { type: string, enum: [mon,tue,wed,thu,fri,sat,sun] }
        desde: { type: string, format: time, example: "13:00:00" }
        hasta: { type: string, format: time, example: "20:00:00" }
        cerrado: { type: boolean }
  intervaloMinutos: { type: integer, example: 30 }
  servicios:                                 # precios y duración, ya vienen de /servicios
    type: array
    items:
      type: object
      required: [id, tipo, precio, duracionMinutos]
      properties:
        id: { type: integer }
        tipo: { type: string }
        precio: { type: number }
        duracionMinutos: { type: integer }
  marca:
    type: object
    properties:
      logoUrl: { type: string }
      instagramUrl: { type: string }
      colorPrimario: { type: string }         # opcional, branding por tenant
```

### Disponibilidad (complemento necesario)

```
GET /publico/negocios/{slug}/disponibilidad?fecha=YYYY-MM-DD&servicioId=12
→ { fecha, zonaHoraria, turnosLibres: ["13:00","13:30", ...] }
```

Hoy sólo existe `GET /turnos/findDateTimes` (devuelve **todos** los `fechaHora` ocupados, sin horarios de apertura, sin filtro por fecha/servicio, sin zona horaria). Con el contrato de arriba el cliente:

- deja de construir la grilla con `TIME_SLOTS` hardcodeados;
- recalcula disponibilidad al cambiar fecha **o** servicio (si la duración afecta el encaje);
- envía `fechaHora` con offset o en la zona declarada por el negocio;
- valida sólo forma y fecha futura; **el backend sigue siendo la autoridad** (ante un turno tomado concurrently responde 409 y la UI lo muestra recuperándose).

## Reglas comerciales en el cliente (a eliminar con el contrato)

Hoy codificadas y que deben salir del frontend:

| Dónde | Qué dice | Con el contrato |
|---|---|---|
| `Reserva.tsx` / `TurnosAdmin.tsx` | lista `TIME_SLOTS` 13:00–19:30 | se reemplaza por `diasAtencion` + `intervaloMinutos` del backend |
| `turno-schema.ts` | `h >= 13 && h < 20 && min ∈ {0,30}` | se elimina; sólo queda forma + fecha futura |
| `turno-schema.ts` | "La fecha debe ser futura" vs `new Date()` local | se compara contra la zona del negocio |

## Coordinación

1. Backend Go expone `GET /publico/negocios/{slug}/configuracion` y la disponibilidad filtrada (o devuelve la config y el front arma la grilla con los datos reales).
2. Se actualiza `backend/openapi-spec.json` (fuente de verdad) → `bun run types`.
3. Se reemplaza el contenido de `lib/business-config.ts` por el hook `useConfiguracionNegocio(slug)` con query key `["configuracion", slug]`.
4. Tests: el wizard renderiza horarios distintos según la config recibida (prueba con dos tenants distintos).
