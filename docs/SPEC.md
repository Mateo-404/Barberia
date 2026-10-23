# SPEC — Marca por tenant, aislamiento y reservas íntegras

Fuente de verdad técnica. Si el repo contradice esta SPEC, **no improvises**: anotalo en `docs/OPEN_QUESTIONS.md`.

## 1. Alcance de esta entrega
1. Multi-tenancy con **tabla compartida + `tenant_id` + RLS** (decisión tomada; no re-discutir).
2. Configuración de marca por tenant (colores, tipografía, logo, estilo, datos del negocio) con **borrador, publicación, versiones y rollback**.
3. Entrega del tema **sin parpadeo** (CSS inyectado desde el servidor).
4. Integridad de reservas en la base (`EXCLUDE`) con respuesta HTTP 409.
5. Wizard de primer login y editor de marca en el panel admin.

## 2. Decisiones de arquitectura (ADR resumidos)
| # | Decisión | Motivo |
|---|---|---|
| A1 | RLS en vez de schema-por-tenant | Muchos tenants chicos; migraciones una sola vez; sin problemas de `search_path` con pooling. |
| A2 | `set_config('app.tenant_id', $1, true)` al inicio de cada transacción | `is_local=true`: no se filtra a otras requests del pool. |
| A3 | `FORCE ROW LEVEL SECURITY` + rol `app_user` sin BYPASSRLS | El owner también queda sujeto a la política. |
| A4 | Tema renderizado en el servidor | Evita FOUC (flash de estilos por defecto). |
| A5 | Fuentes por **allowlist** | Evita inyección CSS y deja listo el self-hosting (fuera de alcance ahora). |
| A6 | Publicación versionada, historial append-only | Rollback seguro; el cliente edita sin romper su sitio público. |

## 3. Modelo de datos
Ver `db/migrations/0001..0003`. Son **referencia**: si ya existen tablas equivalentes (turnos, servicios, usuarios…),
**mapealas y adaptá** las migraciones en vez de duplicarlas. Para datos existentes: una migración aparte, reversible,
que cree el tenant `legacy` y complete `tenant_id` antes de volver `NOT NULL` y activar RLS.

## 4. Resolución de tenant (middleware)
Orden de resolución:
1. `Host` exacto en `tenant_domains` con `verified_at IS NOT NULL`.
2. Subdominio de `BASE_DOMAIN` (env) → `slug`.
3. Solo con `APP_ENV=dev`: header `X-Tenant-Slug`.

Reglas:
- Slugs reservados (no asignables): `www`, `api`, `admin`, `app`, `static`, `assets`, `internal`, `mail`.
- Tenant desconocido o `status='suspended'` → `404` genérico (no revelar si existe).
- Requests autenticados: el `tenant_id` del JWT debe coincidir con el resuelto por host; si no → `403`.
- El `login` se hace **dentro** del tenant resuelto (la tabla `users` tiene RLS).
- El tenant resuelto se guarda en el `context`; el helper de DB lo usa y **no acepta** el tenant por parámetro libre.

## 5. Settings v1 (JSON)
```json
{
  "name": "Barbería Don Mateo", "tagline": "Corte clásico, estilo propio", "logo": "",
  "c": {"primary":"#c9a24b","accent":"#e8d9a8","bg":"#161616","surface":"#222222","text":"#f4f0e8"},
  "hf": "Playfair Display", "bf": "Inter", "r": 8, "btn": "solid",
  "biz": {"whatsapp":"+54 9 3492 123456","instagram":"@donmateo","address":"","currency":"ARS",
          "slot":30,"notice":1,"open":"09:00","close":"20:00","tz":"America/Argentina/Buenos_Aires"}
}
```
Es el mismo formato que envía `docs/historico/kit-referencia/onboarding.html`. `tz` es opcional (default `America/Argentina/Buenos_Aires`).
Se decodifica con `DisallowUnknownFields`.

### Validación (todos los errores se acumulan en `details`, no se corta en el primero)
| Campo | Regla |
|---|---|
| `name` | 2–60 caracteres (contar runas, no bytes) |
| `tagline` | ≤ 120 |
| `logo` | vacío, o `https://…`, o `data:image/(png|jpeg|webp);base64,…`. **SVG prohibido.** ≤ 400 000 caracteres |
| `c.*` | `^#[0-9a-fA-F]{6}$`; se normaliza a minúsculas |
| `hf`, `bf` | pertenecer a la allowlist (§6) |
| `r` | entero 0–24 |
| `btn` | `solid` \| `outline` |
| `biz.whatsapp` | vacío o `^\+?[0-9 ()-]{6,20}$` |
| `biz.instagram` | vacío o `^@?[A-Za-z0-9._]{1,30}$` |
| `biz.currency` | `ARS`,`USD`,`EUR`,`UYU`,`CLP` |
| `biz.slot` | entero 10–240 |
| `biz.notice` | entero 0–168 |
| `biz.open` / `biz.close` | `HH:MM` 24 h y `open < close` |
| `biz.tz` | zona IANA válida (`time.LoadLocation`) |

### Contraste (WCAG 2.x, luminancia relativa)
- `text` sobre `bg`: **< 3:1 → error**; entre 3:1 y 4.5:1 → **warning** (`warnings[]`), se permite guardar.
- `text` sobre `surface` y `primary` sobre `bg`: **< 3:1 → warning**.
- `on-primary` (color del texto sobre el botón principal) **se calcula**: el de mayor contraste entre `#111111` y `#ffffff`
  (empate → `#111111`). Por barrido, ese ratio es siempre ≥ 4.3, así que no necesita validación.
- Vectores de referencia: `backend-go/testdata/contrast_vectors.json`.

## 6. Allowlist de fuentes (v1) y fallback genérico
Serif: Playfair Display, Lora, Merriweather, Cinzel, Cormorant Garamond, Libre Baskerville · 
Sans: Inter, DM Sans, Montserrat, Poppins, Raleway, Roboto, Rubik, Work Sans, Nunito, Josefin Sans, Space Grotesk · 
Display/condensed: Bebas Neue, Oswald, Barlow Condensed, Archivo Black, Anton.
Cada una con su genérico (`serif` / `sans-serif`) para el `font-family`. Una constante `map[string]Font` en un solo lugar.

## 7. Render del tema
`RenderCSS(settings) string` es una **función pura y determinista** (orden fijo de variables, sin timestamps):
```css
:root{--color-primary:#c9a24b;--color-accent:#e8d9a8;--color-bg:#161616;--color-surface:#222222;--color-text:#f4f0e8;--color-on-primary:#111111;--radius:8px;--font-heading:"Playfair Display",serif;--font-body:"Inter",sans-serif}
```
- `hash = hex(sha256(css))[:12]`.
- La página pública incluye el CSS **inline** en el `<head>` (vía `html/template`; el CSS se inyecta como `template.CSS` solo
  después de pasar la validación, nunca desde input crudo). Además `<link rel="preload">` de las fuentes cuando existan.
- Los textos del usuario (`name`, `tagline`, `address`…) salen **escapados** (`html/template`, nunca `text/template`).

## 8. API
Todas devuelven JSON (salvo `.css`). Roles `owner`/`admin` para `/api/admin/*`.
| Método y ruta | Descripción | Respuestas |
|---|---|---|
| `GET /api/admin/settings` | `{onboarded, draft, published_version, hash}` (defaults si no hay fila) | 200 |
| `PUT /api/admin/settings/draft` | Valida y guarda el borrador. No cambia el sitio público | 200 `{draft, warnings[]}` · 422 `{details[]}` |
| `POST /api/admin/settings/publish` | Copia el borrador a `published`, versión N+1, marca `onboarded=true`. Idempotente si el hash no cambió | 200 `{version, hash}` |
| `GET /api/admin/settings/versions` | Últimas 20 `{version, hash, created_at}` | 200 |
| `POST /api/admin/settings/rollback` `{version}` | Crea una **nueva** versión con el contenido de la elegida (no reescribe historia) | 200 · 404 |
| `GET /api/admin/preview/theme.css` | CSS del **borrador**, `Cache-Control: no-store` | 200 |
| `GET /api/public/settings` | Publicado (o defaults), `Cache-Control: public, max-age=60` | 200 |
| `GET /api/public/theme.css?v={hash}` | Si `v` == hash actual: `public, max-age=31536000, immutable`; si no: sirve el actual con `no-cache` + `ETag` | 200 |

Error estándar: `{"error":"…","code":"validation_failed","details":[{"field":"c.primary","message":"…"}]}`.

## 9. Reservas
- Mapear `SQLSTATE 23P01` (exclusion_violation) → **409** `{"code":"slot_taken"}`; `23503` → 422; `23514` → 422.
- La verificación previa en código sigue existiendo **solo** para dar mejor mensaje; la garantía es la base.
- No reescribir el cálculo de disponibilidad en esta entrega.

## 10. Fuera de alcance (NO empezar)
Dominios propios/Caddy (solo existe la tabla `tenant_domains`) · recordatorios por WhatsApp · Mercado Pago y planes/entitlements ·
self-hosting de fuentes · derivar paleta desde un solo color (v2) · migración a React · schema-por-tenant · alta de tenants (signup).

## 11. Preguntas abiertas con default
| Duda | Default si no hay respuesta |
|---|---|
| ¿Cómo se sirve hoy el front? | Go sirve `frontend/` (embed o `http.FileServer`) y renderiza solo la página pública con plantilla |
| ¿Librería de JWT/auth existente? | Reusar la existente; solo agregar el claim `tenant_id` |
| ¿Tabla `users` ya existe? | Adaptar la migración 0001 a la existente |

## 12. Decisiones de implementación (backend Go)
- **Claim de tenant:** se conserva el nombre `tenant` (no se renombra). Su valor es el `tenant_id` como **BIGINT (int64)**. Los tokens sin claim de tenant se rechazan con **401** (re-login); no hay fallback a un tenant por defecto (ADR-002).
- **Claim de rol:** se agrega `rol` (`owner` | `admin` | `staff`) al JWT. `/api/admin/*` exige `owner`/`admin` (403 para `staff`).
- **F-02 (CSS en HTML inicial):** el frontend hace `fetch('/api/public/theme.css')` en `main.tsx` antes de montar React y escribe el resultado en `<style id="tenant-theme">` del `<head>`. El backend no sirve el frontend.
- **Reservas:** `turnos.fecha_hora` se mantiene `TIMESTAMP` (sin zona); la zona horaria se interpreta en la capa de aplicación con `biz.tz`. La unicidad es `(tenant_id, fecha_hora)` (índice único). No hay entidad barbero; un slot por tenant. Ver `docs/OPEN_QUESTIONS.md`.
- **Mapeo de errores de reserva:** 23505 (unique) y 23P01 (exclusion) → 409 con `code: "slot_taken"`; 23503/23514 → 422.
- **RLS en producción:** toda lectura/escritura sobre tablas de tenant pasa por `database.WithTenantTx` (set_config transaction-local). Los handlers no reciben el pool para tablas de tenant.
