# Documentación del proyecto

Índice de la documentación del sistema de gestión de barbería.

## Guías principales

| Documento | Contenido |
|---|---|
| [README](../README.md) | Alcance, requerimientos, demo, instalación y licencia |
| [backend-go/README.md](../backend-go/README.md) | Backend Go: stack, estructura, comandos, tests, multi-tenancy y seguridad |
| [frontend/README.md](../frontend/README.md) | App React: setup, estructura, auth, CSP y comandos |
| [DEPLOY_NOTES.md](DEPLOY_NOTES.md) | Despliegue a producción: roles de BD, migraciones, variables |
| [AGENTS.md](../AGENTS.md) | Arquitectura y comandos de referencia para agentes/asistentes |

## Especificación de marca por tenant, RLS y reservas

| Documento | Contenido |
|---|---|
| [SPEC.md](SPEC.md) | Fuente de verdad técnica: tenancy con RLS, settings con versiones, tema sin parpadeo, integridad de reservas |
| [ACCEPTANCE.md](ACCEPTANCE.md) | Criterios de aceptación y el test que cubre cada uno |
| [OPEN_QUESTIONS.md](OPEN_QUESTIONS.md) | Decisiones tomadas donde la SPEC y el repo divergían |

## Referencia

| Documento | Contenido |
|---|---|
| [atajos-teclado.md](atajos-teclado.md) | Todos los atajos de teclado de la app y cómo agregar nuevos |
| [frontend/](frontend/) | Auditoría, arquitectura objetivo, plan y contrato de configuración del frontend |
| [migracion-go/](migracion-go/) | Migración del backend Java a Go: auditoría, arquitectura, matriz de compatibilidad, plan y ADRs |
| [assets/](assets/) | Diagramas UML y capturas usadas en el README |

### Migración a Go (`migracion-go/`)

| Documento | Contenido |
|---|---|
| [00-auditoria.md](migracion-go/00-auditoria.md) | Auditoría del backend Java y hallazgos |
| [01-arquitectura-objetivo.md](migracion-go/01-arquitectura-objetivo.md) | Arquitectura del monolito modular Go |
| [02-matriz-compatibilidad.md](migracion-go/02-matriz-compatibilidad.md) | Contrato Java→Go, desviaciones y endurecimiento de seguridad |
| [03-plan-migracion.md](migracion-go/03-plan-migracion.md) | Plan por etapas, migración de datos y criterios de aceptación (activación completada) |
| [04-auditoria-seguridad.md](migracion-go/04-auditoria-seguridad.md) | Auditoría de seguridad posterior: fixes aplicados, decisiones y deuda |
| [adr/](migracion-go/adr/) | Decisiones de arquitectura (ADR-001..006) |

## Histórico

Material de etapas anteriores en [`historico/`](historico/). Queda como referencia; rutas y nombres pueden no reflejar la estructura actual (p. ej. `backend/` Java ya no existe).

- [informe-sesion-completo.md](historico/informe-sesion-completo.md) — integración del frontend React, JWT y CI.
- [auditoria-seguridad-2026-07-18.md](historico/auditoria-seguridad-2026-07-18.md) — auditoría de seguridad del manejo de sesión.
- [reporte-pulido-2026-07-17.md](historico/reporte-pulido-2026-07-17.md) — pulido visual tras login, reserva y theming.
- [tech-debt-backend-java.md](historico/tech-debt-backend-java.md) — deuda técnica del backend Java (resuelta o descartada con la migración).
- [discovery-settings-tenancy.md](historico/discovery-settings-tenancy.md) — relevamiento previo a la implementación de settings/RLS.
- [kit-referencia/](historico/kit-referencia/) — prototipos de partida de la SPEC (onboarding HTML, settings schema-per-tenant, migraciones `tstzrange` para un futuro con duraciones variables).

## Fuentes de verdad

- **API/contratos:** `backend-go/api/openapi.json`. Los tipos del frontend (`frontend/src/types/api.ts`) se regeneran con `bun run types` desde `frontend/`.
- **Esquema de DB:** migraciones versionadas en `backend-go/db/migrations/` (golang-migrate, embebidas en el binario); tests SQL en `db/tests/`.
