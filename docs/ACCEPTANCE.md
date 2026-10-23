# ACCEPTANCE — criterios de aceptación

Un ítem está "hecho" cuando existe un test automatizado que falla sin la implementación y pasa con ella.
Los ítems `T-*` y `B-*` ya tienen test SQL en `db/tests/`. Los demás los escribe el agente en Go (`REQUIRE_DB=1`: no se saltean).

## Aislamiento (Fase 1)
| ID | Criterio | Cobertura |
|---|---|---|
| T-01 | Sin tenant seteado, `SELECT` devuelve 0 filas e `INSERT` falla | `rls_isolation.sql` |
| T-02 | El tenant A solo ve sus filas; `UPDATE`/`DELETE` sobre filas de B afectan 0 | `rls_isolation.sql` |
| T-03 | No se puede insertar ni "mover" filas a otro tenant (`WITH CHECK`) | `rls_isolation.sql` |
| T-04 | `app_user` no es superusuario ni tiene BYPASSRLS | `rls_isolation.sql` |
| T-05 | El tenant no se filtra a la siguiente transacción de la misma conexión | `rls_isolation.sql` + **Go**: pool de 1 conexión, dos requests de tenants distintos en serie |
| T-06 | Tenant vacío → 0 filas; tenant inválido → error (nunca "sin filtro") | `rls_isolation.sql` |
| T-07 | Toda tabla con `tenant_id` (salvo `tenant_domains`) tiene RLS habilitada y forzada | `rls_isolation.sql` |
| T-08 | Resolución de tenant: host verificado, subdominio, header solo en dev, reservados, suspendido → 404 | **Go**, tabla de casos |
| T-09 | JWT con tenant ≠ tenant del host → 403 | **Go** |
| T-10 | El helper de DB no ofrece una API que reciba el tenant como string libre desde handlers | revisión de código en el informe |

## Marca / tema (Fase 2)
| ID | Criterio |
|---|---|
| S-01 | Todo caso de `backend-go/testdata/settings_valid.json` (base + variantes) valida OK |
| S-02 | Todo caso de `backend-go/testdata/settings_invalid.json` falla con 422 y `details` menciona `expect_field` |
| S-03 | Logo > 400 000 caracteres rechazado (test generado en código) |
| S-04 | Ratios de `contrast_vectors.json` coinciden (±0.01) y `on_primary` coincide exacto |
| S-05 | Propiedad: para 10 000 colores aleatorios, ratio de `on-primary` ≥ 4.3 |
| S-06 | `RenderCSS` es determinista (misma entrada → mismo string y hash) y solo contiene valores ya validados (hex, enteros, fuentes de la allowlist) |
| S-07 | `name` = `<script>alert(1)</script>` se guarda y se renderiza **escapado** en la página pública |
| S-08 | Draft no altera `/api/public/*`; `publish` sí; publicar dos veces lo mismo no crea versión nueva |
| S-09 | `rollback` crea versión N+1 con el contenido viejo; el historial no se modifica |
| S-10 | `/api/public/theme.css?v=<hash actual>` → header `immutable`; hash viejo → `no-cache` + `ETag` |
| S-11 | Warnings de contraste (3:1–4.5:1) no bloquean el guardado; < 3:1 en texto/fondo sí |
| S-12 | Tenant A no puede leer ni publicar la config de B (test de integración end-to-end vía HTTP) |
| S-13 | Un `staff` recibe 403 en `/api/admin/settings*` |

## Reservas (Fase 3)
| ID | Criterio | Cobertura |
|---|---|---|
| B-01…B-09 | Solapes, contiguos, otro barbero, cancelación libera, reactivar choca, rango `[)`, FK entre tenants, >12 h, sin fin | `booking_exclusion.sql` |
| B-10 | **Concurrencia**: 20 goroutines reservan el mismo barbero y horario → exactamente 1 éxito (201) y 19 conflictos (409), con `-race` | **Go** |
| B-11 | La respuesta 409 tiene `code: "slot_taken"` y no expone detalles internos de SQL | **Go** |

## Frontend (Fase 4)
| ID | Criterio |
|---|---|
| F-01 | Con `onboarded=false`, el primer login del admin redirige al wizard; con `true`, al panel |
| F-02 | La página pública trae `:root{--color-primary` **en el HTML inicial** (sin esperar JS) |
| F-03 | El editor permite: editar borrador, ver vista previa (iframe con `preview/theme.css`), publicar, ver versiones, volver atrás |
| F-04 | Sin frameworks nuevos; sin `localStorage` para datos de marca; sin `innerHTML` con input del usuario sin escapar |
| F-05 | Los textos del wizard están en español es-AR y el layout funciona en 360 px de ancho |

## Seguridad transversal
| ID | Criterio |
|---|---|
| SEC-01 | `make lint-sql` en verde (sin SQL armado con `Sprintf`) |
| SEC-02 | Ningún log contiene contraseñas, JWT ni el valor completo de `logo` (data URI) |
| SEC-03 | Ninguna ruta `/api/admin/*` funciona sin autenticación |
