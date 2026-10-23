package com.barber.barberBackend.dto;

import io.swagger.v3.oas.annotations.media.Schema;

@Schema(description = "Respuesta de inicio de sesión con JWT")
public record LoginResponseDTO(
    @Schema(description = "ID único del administrador", example = "1")
    Long id,
    @Schema(description = "Nombre del administrador", example = "Carlos")
    String nombre,
    @Schema(description = "Apellido del administrador", example = "López")
    String apellido,
    @Schema(description = "Email del administrador", example = "carlos@barberia.com")
    String email,
    @Schema(description = "Token JWT para autenticación", example = "eyJhbGciOiJIUzI1NiJ9...")
    String token
) {}
