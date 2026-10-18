# Documentación del proyecto

Índice de la documentación del sistema de gestión de barbería.

## Guías principales

| Documento | Contenido |
|---|---|
| [README](../README.md) | Alcance, requerimientos, demo, instalación y licencia |
| [backend/README.md](../backend/README.md) | API REST: endpoints, autenticación JWT, perfiles y tests |
| [frontend/README.md](../frontend/README.md) | App React: setup, estructura, auth y comandos |
| [AGENTS.md](../AGENTS.md) | Arquitectura y comandos de referencia para agentes/asistentes |

## Referencia

| Documento | Contenido |
|---|---|
| [atajos-teclado.md](atajos-teclado.md) | Todos los atajos de teclado de la app y cómo agregar nuevos |
| [tech-debt-barberia-fase3b.md](tech-debt-barberia-fase3b.md) | Deuda técnica: lo resuelto y lo pendiente |

## Histórico

Reportes de sesiones de desarrollo anteriores en [`historico/`](historico/) — quedan como referencia; los nombres de archivos y rutas pueden no reflejar la estructura actual (`frontend-react/` → `frontend/`).

- [informe-sesion-completo.md](historico/informe-sesion-completo.md) — integración del frontend React, JWT y CI.
- [auditoria-seguridad-2026-07-18.md](historico/auditoria-seguridad-2026-07-18.md) — auditoría de seguridad del manejo de sesión.
- [reporte-pulido-2026-07-17.md](historico/reporte-pulido-2026-07-17.md) — pulido visual tras login, reserva y theming.

## Fuentes de verdad

- **API/contratos:** `backend/openapi-spec.json` (exportado desde Swagger `/v3/api-docs`). Los tipos del frontend (`frontend/src/types/api.ts`) se regeneran con `bun run types` desde `frontend/`.
- **Esquema de DB:** derivado de JPA (`ddl-auto=update`); no hay migraciones versionadas todavía.
