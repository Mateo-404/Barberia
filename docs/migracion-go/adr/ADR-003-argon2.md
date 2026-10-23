# ADR-003 — Argon2id compatible con Spring Security

- **Estado**: aceptado (2026-10-21)
- **Contexto**: `AdministradorService` usa `Argon2PasswordEncoder(16, 32, 1, 19456, 2)`
  → Argon2id v1.9, `m=19456` KiB, `t=2`, `p=1`, salt 16 B, hash 32 B, formato
  `$argon2id$v=19$m=19456,t=2,p=1$<salt-b64>$<hash-b64>` (verificado en BD).
- **Decisión**: implementar `Hash`/`Verify` con `golang.org/x/crypto/argon2`
  reproduciendo exactamente ese formato y parámetros. `Verify` **parsea los
  parámetros del propio hash** (no de una constante): permite migrar parámetros
  en el futuro sin invalidar cuentas y acepta hashes existentes de Java sin reset.
- **Regeneración de hash**: no se regenera en cada login (comportamiento Java);
  opcionalmente se re-hashea si los parámetros del hash difieren de los actuales
  (deferred, fuera de alcance inicial).
- **Alternativas descartadas**: bcrypt (no es el formato existente); re-hashear
  todo al arrancar (riesgo innecesario).
