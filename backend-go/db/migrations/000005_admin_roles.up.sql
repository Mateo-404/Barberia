-- 000005_admin_roles.up.sql
-- Roles de administrador (SPEC §8: /api/admin/* exige owner/admin).
-- Las filas existentes (instalación single-tenant) quedan como 'owner' para
-- no romper el acceso del admin actual.

ALTER TABLE administradores ADD COLUMN IF NOT EXISTS role TEXT NOT NULL DEFAULT 'owner';

DO
$$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'chk_administradores_role') THEN
        ALTER TABLE administradores ADD CONSTRAINT chk_administradores_role
            CHECK (role IN ('owner', 'admin', 'staff'));
    END IF;
END
$$;

UPDATE administradores SET role = 'owner' WHERE role IS NULL OR role = '';
