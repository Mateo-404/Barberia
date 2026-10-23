# ADR-004 — Correcciones de seguridad deliberadas (hash expuesto / password en claro)

- **Estado**: aceptado (2026-10-21)
- **Defectos** (verificados en el Java real):
  1. `GET /administradores/{page}/{size}` devuelve `content[].contrasenia` con el
     hash Argon2 (confirmado en ejecución).
  2. `PATCH /administradores/{id}` con `contrasenia` persiste el texto plano
     (`BeanUtils.copyProperties` no pasa por el `save()` que hashea).
- **Decisión**: el backend Go
  1. emite la página de administradores con la misma forma pero **sin** el campo
     `contrasenia` (el objeto conserva `id`, `nombre`, `apellido`, `email`);
  2. hashea `contrasenia` cuando aparece en PATCH (mismos parámetros, formato
     `$argon2id$…`), nunca persiste claro.
- **Justificación**: son fugas de credenciales; corregirlas no rompe al frontend
  (no consume esas rutas/ campos) y no silencia la rotura: queda documentado en la
  matriz de compatibilidad como desvío intencional.
- **Nota**: para rollback a Java, el hash de la BD ya es válido en ambos (mismo
  formato); un admin que cambió su contraseña vía PATCH Go sigue pudiendo loguear
  en Java.
