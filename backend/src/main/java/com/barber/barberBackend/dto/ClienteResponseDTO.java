package com.barber.barberBackend.dto;

import io.swagger.v3.oas.annotations.media.Schema;

@Schema(description = "Cliente registrado en la barbería")
public record ClienteResponseDTO(
    @Schema(description = "Número de teléfono del cliente", example = "1155551234", requiredMode = Schema.RequiredMode.REQUIRED)
    String telefono,
    @Schema(description = "Nombre del cliente", example = "Juan", requiredMode = Schema.RequiredMode.REQUIRED)
    String nombre,
    @Schema(description = "Apellido del cliente", example = "Pérez", requiredMode = Schema.RequiredMode.REQUIRED)
    String apellido,
    @Schema(description = "Email del cliente", example = "juan.perez@email.com")
    String email
) {}
