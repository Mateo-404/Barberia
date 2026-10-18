# Tech Debt — Barbería Fase 3b+

## Resuelto

- [x] Response DTOs Cliente/Turno/Servicio/Administrador (Fase 3b)
- [x] Request DTOs + Bean Validation las 4 entidades (Fase 3b+9)
- [x] getById()/update() migrados a DTO en GenericController (Fase OpenAPI)
- [x] LoginRequestDTO — login ya no expone entidad Administrador cruda
- [x] create()/createAll() retornan entidad persistida con ID (Fase 5)
- [x] Tests de repositorio con base real en memoria: `@DataJpaTest` sobre H2 (perfil `test`) para Turno/Cliente/Administrador (`ITurnoRepositoryTest`, `IClienteRepositoryTest`, `IAdministradorRepositoryTest`). Verificado: 64 tests en verde.

## Pendiente

- Tests end-to-end (app completa levantada + HTTP + DB): la cobertura actual son slices de controller con MockMvc y tests de repositorio por separado.
- getPage() en GenericController retorna `Page<T>` (entidad JPA) en vez de `Page<DTO>` — no migrado porque afectaría la paginación. Si se requiere en frontend, evaluar migración.
- `@RequestBody T entity` en `update()` de GenericController acepta entidad JPA directamente en vez de un RequestDTO — exposición de datos en el spec OpenAPI. Requiere diseñar un UpdateRequestDTO genérico o que cada controller concrete defina su propio PATCH.
- `login()` de AdministradorController ahora usa `@RequestBody @Valid LoginRequestDTO`, pero este DTO no extiende ninguna interfaz genérica; cualquier otro controller que necesite login (ej: cliente) necesitaría su propio DTO.
- PanelEstadisticaDTOService no tiene manejo de errores específico; devuelve `null` si falla algo, pero PanelEstadisticasControllerDTO responde 200 con null. Considerar 500 explícito.
