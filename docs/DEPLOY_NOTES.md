# DEPLOY_NOTES.md — despliegue del backend Go en producción

## Variables
- `MIGRATE_DATABASE_URL`: conexión como dueño/superusuario para ejecutar migraciones (golang-migrate).
- `DATABASE_URL`: conexión como `app_user` (sin SUPERUSER, sin BYPASSRLS) para la app.
- `APP_ENV`: `dev` | `prod` (default `prod`). Si != `dev`, al arrancar se aborta si `current_user` es superusuario o tiene BYPASSRLS.
- `JWT_SECRET`: obligatorio, ≥32 bytes (`openssl rand -base64 32`).
- `AUTO_MIGRATE`: dejar en `false` en producción (`app_user` no tiene permisos de DDL; las migraciones se corren aparte).
- `CORS_ALLOW_ORIGINS`: origen exacto del frontend publicado.
- `TRUST_PROXY_HEADERS=true` solo detrás de un proxy que sobrescriba `X-Forwarded-For`.
- Resto de variables y defaults: `backend-go/.env.example`.

## Crear rol app_user (manual)
```sql
CREATE ROLE app_user WITH LOGIN PASSWORD '<strong>';
-- NO: SUPERUSER, NO: BYPASSRLS
```

## Grants y ALTER DEFAULT PRIVILEGES
```sql
GRANT CONNECT ON DATABASE <db> TO app_user;
GRANT USAGE ON SCHEMA public TO app_user;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO app_user;
GRANT USAGE, SELECT, UPDATE ON ALL SEQUENCES IN SCHEMA public TO app_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO app_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT USAGE, SELECT, UPDATE ON SEQUENCES TO app_user;
```

## Orden de despliegue
0. Backup (`pg_dump`). Si la BD viene del backend Java (Hibernate), probar antes `000002` contra un dump real en staging (ver `docs/migracion-go/03-plan-migracion.md`).
1. Ejecutar migraciones con `MIGRATE_DATABASE_URL` (dueño): `cd backend-go && MIGRATE_DATABASE_URL=... go run ./cmd/migrate`.
2. Configurar `DATABASE_URL` apuntando a `app_user`.
3. Arrancar la app (imagen de `backend-go/Dockerfile`) sin `APP_ENV` o con `APP_ENV=prod` (chequeo de rol obligatorio).
4. Frontend: `cd frontend && VITE_API_BASE_URL=https://<api> bun run build` y publicar `dist/` (la CSP permite automáticamente ese origen).

## Tokens legados
Tokens sin claim `tenant` (claim nombre `tenant`, valor BIGINT) —p. ej. emitidos por el backend Java— son rechazados con 401 → el frontend limpia la sesión y pide re-login. No hay fallback.
