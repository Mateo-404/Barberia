# 02 — Matriz de compatibilidad de API (Java → Go)

Fuente: `backend/openapi-spec.json` + código Java + **mediciones contra el backend
Java real** (perfil `h2`, 2026-10-21; ver `00-auditoria.md` §4). Todo endpoint Java
tiene contraparte Go con la misma ruta y método. Salvo las excepciones marcadas
"⚠ DECISIÓN", la respuesta es byte-compatible salvo el orden de claves JSON
(irrelevante) y la notación numérica de precios (`5000.0` Java vs `5000` Go en JSON;
mismo valor, `type: number`).

## Convenciones del error handler (réplica)

| Caso Java | Status | Body Go (idéntico) |
|---|---|---|
| Sin token en ruta protegida | 401 | `{"type":"/errors/unauthorized","title":"No autorizado","status":401,"detail":"Se requiere autenticación para acceder a este recurso"}` (sin `instance`) |
| Token `Bearer` inválido/expirado (incluso en rutas públicas) | 401 | igual, `detail:"Token JWT inválido o expirado"` |
| JSON ilegible o validación `@Valid` fallida | 400 | `{"type":"about:blank","title":"Bad Request","status":400,"detail":"Invalid request content.","instance":"<path>"}` |
| Regla de negocio | 400 | `{"type":"/errors/business-rule-violation","title":"Violación de regla de negocio","status":400,"detail":"<msg>","instance":"<path>"}` |
| Login inválido | 401 | `{"type":"/errors/invalid-credentials","title":"Credenciales inválidas","status":401,"detail":"Credenciales inválidas","instance":"/administradores/login"}` (o `detail:"El email y la contraseña no pueden ser nulos"`) |
| Duplicado | 409 | `{"type":"/errors/resource-already-exists","title":"El recurso ya existe","status":409,"detail":"Ya existe un cliente con el teléfono X"}` / `...con el email X` |
| Servicio inexistente al crear turno | 404 | `{"type":"/errors/resource-not-found","title":"Recurso no encontrado","status":404,"detail":"No existe un servicio con el id X","instance":"/turnos"}` |
| GET/PATCH/DELETE/exists por ID inexistente | 404 | **cuerpo vacío** |
| Error interno | 500 | `{"type":"/errors/internal-server-error","title":"Error interno del servidor","status":500,"detail":"Ocurrió un error inesperado. Intente nuevamente más tarde.","instance":"<path>"}` |

## /ping

| | Java | Go |
|---|---|---|
| Auth | pública | pública |
| Respuesta | 200 text/plain `Backend funcionando!` | idéntico |

## /administradores

| Método Ruta | Auth | Input | Java | Go |
|---|---|---|---|---|
| GET `/administradores` | Bearer | — | 200 `AdministradorResponseDTO[]` | idéntico |
| POST `/administradores` | Bearer | `AdministradorRequestDTO` (nombre/apellido/emailNotBlank, email válido, contrasenia ≥ 8) | 201 DTO; 400 validación; 409 email duplicado | idéntico (unicidad atómica) |
| POST `/administradores/all` | Bearer | `[AdministradorRequestDTO]` | 201 DTO[]; 409 (primero detectado) | idéntico (todo en transacción) |
| POST `/administradores/login` | pública | `LoginRequestDTO` | 200 `LoginResponseDTO{id,nombre,apellido,email,token}`; 401; 400 validación | idéntico. Token HS256 claims `sub`,`email`,`iat`,`exp`(24 h) + claim `tenant` (aditivo, ignorado por Java). ⚠ NUEVO: rate limit 5/min por IP → 429 `{"type":"/errors/rate-limited",...}` (documentado; solo afecta abuso) |
| GET `/administradores/{id}` | Bearer | id numérico | 200 DTO / 404 vacío | idéntico |
| PATCH `/administradores/{id}` | Bearer | campos parciales | 200 DTO / 404 vacío. ⚠ DEFECTO: `contrasenia` en texto plano | 200 DTO / 404 vacío. **⚠ DECISIÓN:** si el body trae `contrasenia` se hashea (Argon2id). Nunca texto plano (ADR-004) |
| DELETE `/administradores/{id}` | Bearer | — | 204 / 404 vacío | idéntico |
| GET `/administradores/{page}/{size}` | Bearer | path ints, page 0-based | 200 `PageAdministrador` con **hash** en `content[].contrasenia` | 200 misma forma. **⚠ DECISIÓN:** `content[].contrasenia` omitido (ADR-004). El frontend no consume esta ruta |
| GET `/administradores/exists/{id}` | Bearer | — | 200 bool | idéntico |

`LoginResponseDTO`: los 5 campos se emiten siempre (en Java son nullable pero el
login exitoso siempre los llena). Tipos TS: se endurece `required` en el nuevo spec
(sin cambio de runtime).

## /clientes

| Método Ruta | Auth | Input | Java | Go |
|---|---|---|---|---|
| GET `/clientes` | Bearer | — | 200 `ClienteResponseDTO[]` | idéntico |
| POST `/clientes` | Bearer | `{telefono*,nombre*,apellido*,email?}` | 201; 400 validación; 409 teléfono dup | idéntico (unicidad atómica) |
| POST `/clientes/all` | Bearer | `[ClienteRequestDTO]` | 201; 409 | idéntico (transaccional) |
| GET `/clientes/{telefono}` | Bearer | string PK | 200 DTO / 404 vacío | idéntico |
| PATCH `/clientes/{telefono}` | Bearer | parcial | 200 DTO / 404 vacío; copia no-nulos | idéntico (incluye `telefono` como cambio de PK documentado: no soportado → se actualizan campos no-clave; la ruta ya no coincide tras cambiarlo) |
| DELETE `/clientes/{telefono}` | Bearer | — | 204 / 404 vacío; FK-restrict → 500 | idéntico |
| GET `/clientes/{page}/{size}` | Bearer | ints | 200 `PageCliente` (entidades) | idéntico |
| GET `/clientes/exists/{id}` | Bearer | — | 200 bool | idéntico |

## /servicios

| Método Ruta | Auth | Input | Java | Go |
|---|---|---|---|---|
| GET `/servicios` | **pública** | — | 200 `ServicioResponseDTO[]` | idéntico |
| POST `/servicios` | Bearer | `{tipo*, precio?}` | 201 / 400 | idéntico |
| POST `/servicios/all` | Bearer | `[ServicioRequestDTO]` | 201 / 400 | idéntico |
| GET `/servicios/{id}` | **pública** | numérico | 200 DTO / 404 vacío | idéntico |
| PATCH `/servicios/{id}` | Bearer | parcial | 200 DTO / 404 vacío | idéntico |
| DELETE `/servicios/{id}` | Bearer | — | 204 / 404 vacío; FK-restrict → 500 | idéntico |
| GET `/servicios/{page}/{size}` | **pública** | ints | 200 `PageServicio` | idéntico |
| GET `/servicios/exists/{id}` | **pública** | — | 200 bool | idéntico |

## /turnos

| Método Ruta | Auth | Input | Java | Go |
|---|---|---|---|---|
| GET `/turnos` | Bearer | — | 200 `TurnoResponseDTO[]` (entidad→DTO aplanado, orden inserción) | idéntico, `ORDER BY id` (mismo orden observado, determinista) |
| POST `/turnos` | **pública** | `TurnoRequestDTO{fechaHora*(futura), idServicio*, telefonoCliente*, nombreCliente*, apellidoCliente*, emailCliente?}` | 201 DTO; 400 validación/regla; 404 servicio inexistente. Reglas: 13:00–20:00, :00/:30, sin duplicado futuro, find-or-create cliente | idéntico + **atómico** (UNIQUE fecha_hora por tenant). ⚠ NUEVO: tenant resuelto por Host. ⚠ NUEVO: rate limit 30/min por IP → 429 `{"type":"/errors/rate-limited",...}` (ver §"Endurecimiento de seguridad") |
| POST `/turnos/all` | Bearer | `[TurnoRequestDTO]` | ⚠ ROTO: 500 si el cliente es nuevo (FK; bypass del find-or-create). 201 solo con cliente existente | **⚠ DECISIÓN:** implementado correctamente con las mismas reglas que `POST /turnos` en transacción (ADR-005). Forma de respuesta idéntica |
| GET `/turnos/{id}` | Bearer | — | 200 DTO / 404 vacío | idéntico |
| PATCH `/turnos/{id}` | Bearer | entidad parcial `{fechaHora?, cliente?:{…}, servicio?:{id}}` | 200 DTO / 404 vacío. Sin validación de reglas. `servicio:{id}` inválido → 500. Respuesta con `tipoServicio:null,precioServicio:0.0` (objeto sin refrescar) | 200 DTO / 404 vacío. **⚠ DECISIÓN (M8):** `servicio:{id}` inexistente **o de otro tenant** → 404 `{"type":"/errors/resource-not-found",...,"detail":"No existe un servicio con el id X"}` (antes 500 por FK). La FK sigue dando 500 para un `cliente` inexistente (paridad Java). Resto idéntico, incluido el objeto sin refrescar |
| DELETE `/turnos/{id}` | Bearer | — | 204 / 404 vacío | idéntico |
| GET `/turnos/{page}/{size}` | Bearer | ints | 200 `PageTurno` (entidades anidadas `cliente`,`servicio`) | idéntico |
| GET `/turnos/exists/{id}` | Bearer | — | 200 bool | idéntico |
| GET `/turnos/findDateTimes` | **pública** | — | 200 `string[]` ISO **sin segundos** (`2026-10-22T15:00`), solo `>= ahora`, orden asc | idéntico. ⚠ NUEVO: solo turnos del tenant resuelto |

## /estadisticas

| Método Ruta | Auth | Java | Go |
|---|---|---|---|
| GET `/estadisticas/panel` | Bearer | 200 `PanelEstadisticaDTO` (turnosHoy/Ayer, ingresosMes/Anterior, cantClientes, ingresosDiarios hoy-primero ~31 buckets, servicios último mes count desc, horarios por hora desde 1/1). Sobre error interno: 200 con cuerpo `null` | 200 idéntico. **⚠ DECISIÓN:** error interno → 500 `/errors/internal-server-error` (el `null` del Java no es un contrato utilizable; el frontend solo trata el caso 200 con DTO) |

## Formatos de identificador y fecha

- IDs numéricos: int64 en JSON (sin strings). `telefono`: string.
- `fechaHora`: `"YYYY-MM-DDTHH:MM:SS"` (+ fracción `.fff` solo si la tiene).
- `findDateTimes`: `"YYYY-MM-DDTHH:MM"` (réplica de `LocalDateTime.toString()`:
  sin segundos cuando son cero).
- Precios: `number` (Java escribe `5000.0`; Go escribe `5000` — mismo valor JSON).

## Cambios aditivos (no rompen al frontend)

1. Claim `tenant` en el JWT emitido.
2. Rate limit de login (429).
3. `GET /healthz`, `GET /readyz` (nuevos, no colisionan con `/ping`).
4. `401` documentado en el OpenAPI (ya existía en runtime).
5. Respuestas de validación con `detail: "Invalid request content."` (réplica exacta).

## Regresión funcional vs tests Java (64)

Los 64 tests Java cubren: controllers (CRUD + duplicados + login), services
(reglas de turno), repositories (queries del panel) y Argon2. La suite Go replica
la cobertura semántica: ver `internal/httpapi/*_test.go` (contrato HTTP),
`internal/turnos/*_test.go` (reglas), `internal/auth/*_test.go` (Argon2/JWT) e
`internal/integration/` (repositorios, constraints, concurrencia). Diferencias
justificadas documentadas arriba (POST /turnos/all, hash omitido, panel 500→500).

## Desviaciones adicionales detectadas durante la implementación

Estas diferencias surgieron al ejecutar la suite de integración contra
PostgreSQL real con el esquema equivalente al de Hibernate. Ninguna afecta al
frontend React (verificado con `bun run build` + tipos regenerados).

1. **Duplicado de turno en el pasado → 400** (Java: lo permitía). El índice
   único `uq_turno_tenant_fecha` (ADR-006) rechaza cualquier duplicado, pasado
   o futuro, con `400 /errors/invalid-argument`. Java solo rechazaba duplicados
   *futuros* vía regla de negocio y permitía libremente duplicados pasados.
   Se considera corrección, no regresión: la tabla `turnos` es un calendario,
   un turno duplicado es un estado inválido.
2. **`PATCH /clientes/{telefono}` con `telefono` en el body**: Java actualizaba
   la PK (y dejaba la ruta desincronizada). Go **ignora** el `telefono` del body
   y conserva la PK original (decisión de integridad documentada; el frontend
   nunca lo envía).
3. **`PATCH /turnos/{id}` con `id` en el body**: Java lo ignoraba también
   (`@PathVariable` manda). Go idéntico: el `id` del body se ignora.
4. **Filas legacy con `nombre`/`apellido` NULL**: Java (Jackson) serializa
   `null` tal cual; Go con `pgx` las mapea a `""`. Solo aplica a filas creadas
   antes de que las columnas fueran NOT NULL; el endpoint de la página de
   administradores es el único afectable y el frontend no lo consume.
   *Mitigación*: el backfill de adopción NO altera datos; si se quiere salida
   idéntica habría que ajustar los DTOs con punteros. **Pendiente de decisión
   del operador** — sin dump real no se ha confirmado si existe tal dato.
5. **`Content-Type` de errores: `application/problem+json`** (RFC 9457), igual
   que Spring `ProblemDetail` (verificado en la auditoría §4; no añade
   `charset`, que además no aplica a JSON).

## Endurecimiento de seguridad (auditoría posterior a la migración)

Tras la migración se ejecutó una auditoría de seguridad sobre el backend Go y el
frontend React. Los hallazgos aplicados y su impacto en el contrato:

| # | Hallazgo | Cambio aplicado | Impacto en el contrato |
|---|---|---|---|
| A1/M6 | Data race en `tenant.Resolver` (map sin sincronizar) + consulta por Host en cada request con Host desconocido | `sync.RWMutex` + caché de "misses" | Ninguno (bug de crash) |
| M1 | Faltaban cabeceras de seguridad; sin `Cache-Control` en respuestas sensibles | `X-Content-Type-Options`, `X-Frame-Options`, `Referrer-Policy`, `Cross-Origin-Resource-Policy`; HSTS si TLS; `Cache-Control: no-store` en login y respuestas autenticadas | Ninguno (aditivo) |
| M2 | PATCH sin validar (Java tampoco validaba: `@Valid` solo en POST) | Validación de campos presentes en PATCH (`email`, `contrasenia ≥ 8`, `precio > 0`, `nombre`/`apellido`/`tipo` no vacíos) + topes de longitud en create/PATCH | 400 donde Java daba 200 con datos inválidos. El frontend valida con Zod antes, así que no lo dispara |
| M3 | `size` sin tope ni control de overflow en la paginación | `size > 100` → 500 (buckets de error consistentes); `page*size` fuera de int32 → 400 | 500/400 donde Java servía la página completa. La ruta no la consume el frontend React |
| M4 | `MAX_REQUEST_BODY` era config muerta (constante fija) | Conectada al tope real de lectura del body en todos los handlers | Ninguno (mismo default 1 MiB) |
| M5 | Sin límite en escrituras públicas | Rate limit `TURNO_RATE_LIMIT_PER_MINUTE` (30) por IP en `POST /turnos` → 429 | Aditivo; solo bajo abuso |
| M7 | Sin timeouts de servidor ni de pool | `Read/Write/IdleTimeout` en `http.Server`; `statement_timeout` y `AcquireTimeout` en el pool | Ninguno (protección ante consultas colgadas) |
| M8 | `PATCH /turnos` podía enlazar un servicio de otro tenant | Validación de pertenencia al tenant en `PATCH /turnos` | Ver fila de PATCH arriba (404) |
| B2 | Colisión de tabla/consulta en el resolver | Ver A1/M6 | — |
| B3 | `DATABASE_URL` podía filtrarse en el log de arranque | El error de parseo deja de incluir la cadena | Ninguno |
| B4 | Reflejo de la ruta en el `detail` del 404 | Mitigado por `X-Content-Type-Options: nosniff` + CSP `default-src 'none'`; se mantiene la paridad con Spring | Ninguno |
| B5 | El rate limit usaba `RemoteAddr` (tras proxy, todos los clientes comparten IP) | `TRUST_PROXY_HEADERS` (opt-in) usa el primer `X-Forwarded-For` | Ninguno por defecto |

Las desviaciones que cambian un código de estado (M2, M3, M5, M8) están marcadas
como "⚠ DECISIÓN" arriba y no afectan al frontend React: este valida con Zod,
no consume la paginación ni envía PATCH a clientes/servicios/administradores.
