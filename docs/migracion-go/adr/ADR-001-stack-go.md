# ADR-001 — Stack Go: chi + pgx + sqlc

- **Estado**: aceptado (2026-10-21); enmendado el 2026-10-21 (ver "Enmienda")
- **Contexto**: migrar un monolito Spring Boot a Go manteniendo contrato, seguridad
  y trazabilidad del SQL; preparar base para SaaS.
- **Decisión**: `net/http` + `chi v5.5.0`; `pgx/v5 v5.11.0` + `pgxpool`; `sqlc v1.31.1`
  para acceso tipado con SQL explícito; `golang-migrate v4.20.1`; `golang-jwt v5.3.1`
  (HS256 estricto); `argon2` de `x/crypto`.
- **Alternativas descartadas**: GORM/ent (ORM oculta consultas, peor control);
  echo/gin (runtime propio sobre `net/http`, innecesario); sqlx con reflection
  (menos verificable que sqlc).
- **Consecuencias**: SQL auditable en `db/queries/`; dependencia de tooling
  (sqlc, migrate) fijado en versiones exactas + CI.

## Enmienda (2026-10-21): oapi-codegen descartado

Se decidió inicialmente usar `oapi-codegen` para generar `ServerInterface` y
tipos. Durante la implementación se descartó porque:

1. Los DTOs del backend Java tienen formas muy concretas que hay que respetar
   al milímetro (nulos omitidos vs serializados, claves en camelCase, Spring
   `Page`). Los tipos generados interferían con ese control fino.
2. El acoplamiento al codegen añadía fricción sin garantía extra: el contrato
   ya se verifica con tests de integración E2E contra el OpenAPI spec y con
   los tests de la matriz de compatibilidad.
3. Menos dependencias de build = build más simple y reproducible.

**Resultado**: `api/openapi.json` sigue siendo el contrato único de verdad
(generado desde `backend/openapi-spec.json`, extendido con `/healthz`,
`/readyz` y `429` en login); los DTOs y rutas están escritos a mano en
`internal/httpapi/` y `internal/modules/*/`, y el contrato se valida por
tests en vez de por compilación.
