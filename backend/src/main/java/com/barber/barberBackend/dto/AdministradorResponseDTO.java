package com.barber.barberBackend.dto;

import io.swagger.v3.oas.annotations.media.Schema;

@Schema(description = "Administrador registrado en el sistema")
public record AdministradorResponseDTO(
    @Schema(description = "ID único del administrador", example = "1", requiredMode = Schema.RequiredMode.REQUIRED)
    Long id,
    @Schema(description = "Nombre del administrador", example = "Carlos", requiredMode = Schema.RequiredMode.REQUIRED)
    String nombre,
    @Schema(description = "Apellido del administrador", example = "López", requiredMode = Schema.RequiredMode.REQUIRED)
    String apellido,
    @Schema(description = "Email del administrador", example = "carlos.lopez@barberia.com", requiredMode = Schema.RequiredMode.REQUIRED)
    String email
) {}
