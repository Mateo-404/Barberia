# OPEN_QUESTIONS.md

## Fase 3 - Reservas  (RESUELTAS)

- **Tipo y zona horaria de fecha_hora:** Se mantiene `TIMESTAMP` (sin zona) en BD, tal como el esquema base. La interpretación de zona horaria se hace en la capa de aplicación usando `biz.tz` del settings del tenant; la BD almacena la hora local del negocio. Motivo: los turnos son slots fijos de 30 min y el esquema existente (Java) usaba `TIMESTAMP`; migrar a `timestamptz` exigiría reescribir datos y no aporta valor para slots fijos. Si en el futuro se agregan duraciones variables, se migra a `tstzrange` + EXCLUDE (migración reversible preparada, ver `docs/historico/kit-referencia/db/migrations`).

- **Entidad barbero/profesional:** No existe tabla `barberos`/`profesionales`. La unicidad es `(tenant_id, fecha_hora)` (índice único `uq_turno_tenant_fecha`), es decir, un slot por tenant. No se requiere EXCLUDE por recurso. Si se agregan barberos, se crea migración `000007` con `tstzrange` + EXCLUDE por `(tenant_id, barbero_id)`.

## Consecuencias
- B-01..B-09 del kit (solapes de rangos) no aplican: con slots fijos, "solape" = misma `fecha_hora`. B-11 (409 `slot_taken`) se cumple vía unique violation (23505) → `KindConflict` → `conflictSlotTaken`.
- El mapeo 23P01→409 se mantiene por compatibilidad futura.
