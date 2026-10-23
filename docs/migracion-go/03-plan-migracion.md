# 03 — Plan de migración, riesgos y criterios de aceptación

## Etapas

| Etapa | Contenido | Estado |
|---|---|---|
| A | Auditoría (`00`), arquitectura (`01`), matriz (`02`), plan (`03`), ADRs | ✅ |
| B | `backend-go/` en paralelo: config, HTTP, auth, tenant, módulos, tests HTTP | ✅ |
| C | Migraciones SQL + sqlc + integración PostgreSQL real (constraints, concurrencia) | ✅ |
| D | Compatibilidad funcional verificada por suite Go + comparación con respuestas Java capturadas | ✅ |
| E | Docker, CI, docs, `.env.example`, flujo de tipos TS | ✅ |
| F | Activación: Go es el único backend en `main`; `backend/` (Java), su CI y el sitio estático legacy fueron eliminados (2026-10-23) | ✅ (despliegue productivo: ver `docs/DEPLOY_NOTES.md`) |

## Etapa F — activación

> **Estado (2026-10-23):** el Java se retiró de `main`. El texto siguiente queda
> como registro del procedimiento de conmutación; el rollback al Java requiere
> la rama `backend` o el historial de git. Nota: hoy los tokens sin claim
> `tenant` se **rechazan** con 401 (no hay fallback por `sub`).

Procedimiento original:

1. Levantar `backend-go` contra una copia de la BD de producción y validar.
2. Apuntar el proxy/dominio al Go (mismo puerto 8080; el frontend no cambia nada).
3. Rollback: re-apuntar al Java; los tokens emitidos por Go siguen siendo
   válidos para Java (HS256, mismos claims; el claim extra `tenant` es ignorado
   por jjwt) y viceversa los de Java son aceptados por Go (fallback de tenant
   por `sub`).

## Migración de datos (BD existente creada por Hibernate `ddl-auto=update`)

Migraciones en `backend-go/db/migrations/`:

1. `000001_baseline` — crea el esquema canónico (tenant + tablas + constraints).
   Para BD vacías (dev/demo).
2. `000002_adopción_hibernate` — para BD existentes del Java: agrega `tenant_id`,
   crea tenant por defecto (`DEFAULT_TENANT_SLUG`, slug `default`), backfill de
   todas las filas, `SET NOT NULL`, índices/únicas. Idempotente y reversible
   (down: elimina tenant y columnas agregadas). **Antes de ejecutarla en
   producción: backup (`pg_dump`) y ejecución en staging con un dump real.**
   Este entorno no tiene acceso a una BD de producción: la migración está probada
   contra el esquema equivalente generado por Hibernate (verificado con
   `spring-boot:run` sobre H2 y con el DDL de Hibernate 6 para PostgreSQL);
   la validación final contra el dump real queda como paso de despliegue.
3. Secuencias: `turno.id` usa `hibernate_sequence` en Java; la migración crea
   `turno_id_seq` y la deja en `MAX(id)+1` (nextval seguro con datos existentes).

Prohibido: migraciones destructivas sobre datos reales; `TRUNCATE`; borrar tablas
con datos.

## Riesgos

| Riesgo | Mitigation |
|---|---|
| Formato de hash Argon2 inesperado en BD | Verificación por prefijo + tests con hash real de Spring (`$argon2id$v=19$m=19456,t=2,p=1$…`, muestra extraída de la BD demo) |
| Tokens Java aceptados por Go con claims distintos | Validación estricta `HS256` + exp; `sub` parseable a int64; fallback tenant por BD |
| Esquema Hibernate real difiere del supuesto | `000002` es idempotente (`IF NOT EXISTS`), se ejecuta primero en staging; documentado |
| Diferencias numéricas JSON (`5000.0` vs `5000`) | Irrelevante para `JSON.parse`/TS; documentado |
| `/turnos/all` corregido cambia comportamiento | Ruta no usada por el frontend; era un 500; documentado (ADR-005) |
| Índice único rechaza duplicados pasados (Java los permitía) | Es un estado inválido; documentado (ADR-006); el backfill de `000002` falla ante duplicados reales (deliberado) |
| Rate limit afecta legítimos | 5/min por IP solo en login, configurable, 429 documentado |

## Criterios de aceptación

- [x] `backend-go` compila con las versiones documentadas (`go build ./...`).
- [x] 36 operaciones del contrato implementadas; matriz completada.
- [x] Frontend React compila sin cambios (no se modificó el código del
      frontend; el script `types` apunta ahora a `backend-go/api/openapi.json`
      y los tipos se regeneran con `bun run types`).
- [x] PostgreSQL se inicializa con las migraciones (verificado contra PG 18 real).
- [x] Datos existentes conservables vía `000002` (procedimiento documentado;
      validación final pendiente de dump real — §migración de datos).
- [x] Argon2id compatible con hashes de Spring; JWT compatible bidireccionalmente.
- [x] Tests de aislamiento de tenant pasan (integration build tag).
- [x] Concurrencia de reservas garantizada por UNIQUE + test de carrera.
- [x] Unitarias + HTTP + integración pasan en entorno limpio (`make test`).
- [x] Dockerfile multi-stage non-root + `docker-compose.yml` con healthchecks.
      (Build de imagen no ejecutado en este entorno: el daemon Docker no está
      disponible; documento como pendiente de verificación.)
- [x] CI nueva `.github/workflows/backend-go-ci.yml` (fmt, lint, build, test,
      integración con service container postgres, govulncheck).
- [x] Sin secretos ni diagnósticos expuestos; `JWT_SECRET` fail-closed.
- [x] Activación como backend activo y remoción del Java del repositorio
      (despliegue productivo según `docs/DEPLOY_NOTES.md`).

## Comandos

```bash
# Backend Go — desarrollo
cd backend-go
cp .env.example .env            # completar JWT_SECRET (openssl rand -base64 32)
make migrate-up                 # aplica migraciones a DATABASE_URL
make run                        # build + ejecuta bin/api
make test                       # unitarias + HTTP (-race)
make test-integration           # requiere PostgreSQL (TEST_DATABASE_URL_BASE)
make lint                       # golangci-lint
make vuln                       # govulncheck
make sqlc                       # regenera código de acceso a BD
```
