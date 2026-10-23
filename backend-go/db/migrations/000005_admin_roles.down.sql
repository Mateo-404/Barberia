-- 000005_admin_roles.down.sql
ALTER TABLE administradores DROP CONSTRAINT IF EXISTS chk_administradores_role;
ALTER TABLE administradores DROP COLUMN IF EXISTS role;
