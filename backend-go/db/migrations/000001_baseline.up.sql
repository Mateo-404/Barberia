-- 000001_baseline.up.sql
-- Esquema base equivalente al generado por Hibernate (ddl-auto=update) del
-- backend Java, para instalaciones nuevas. Todos los CREATE son idempotentes
-- (IF NOT EXISTS) para que esta migración también pueda ejecutarse contra una
-- base creada por el Java sin dañarla: en ese caso no hace nada y la
-- multi-tenantización la completa 000002.
--
-- Nombres de tablas/columnas calculados con la convención de Spring Boot
-- (CamelCaseToUnderscores): clientes(telefono,nombre,apellido,email),
-- servicios(id,tipo,precio), administradores(id,nombre,apellido,email,contrasenia),
-- turnos(id,fecha_hora,cliente_id,servicio_id).

CREATE TABLE IF NOT EXISTS clientes (
    telefono TEXT PRIMARY KEY,
    nombre   TEXT,
    apellido TEXT,
    email    TEXT
);

CREATE TABLE IF NOT EXISTS servicios (
    id     BIGSERIAL PRIMARY KEY,
    tipo   TEXT,
    precio REAL
);

CREATE TABLE IF NOT EXISTS administradores (
    id          BIGSERIAL PRIMARY KEY,
    nombre      TEXT,
    apellido    TEXT,
    email       TEXT,
    contrasenia TEXT
);

-- Mismo nombre de índice que creaba Hibernate en administradores.email.
CREATE INDEX IF NOT EXISTS idx_admin_email ON administradores (email);

CREATE TABLE IF NOT EXISTS turnos (
    id         BIGSERIAL PRIMARY KEY,
    fecha_hora TIMESTAMP NOT NULL,
    cliente_id TEXT REFERENCES clientes (telefono),
    servicio_id BIGINT REFERENCES servicios (id)
);

-- Mismo nombre de índice que creaba Hibernate en turnos.fecha_hora.
CREATE INDEX IF NOT EXISTS idx_turno_fechahora ON turnos (fecha_hora);
