package com.barber.barberBackend.dto;

import io.swagger.v3.oas.annotations.media.Schema;

import jakarta.validation.constraints.Email;
import jakarta.validation.constraints.NotBlank;

@Schema(description = "Datos para crear o actualizar un cliente")
public record ClienteRequestDTO(
    @Schema(description = "Número de teléfono del cliente (usado como ID)", example = "1155551234")
    @NotBlank(message = "El teléfono es obligatorio")
    String telefono,
    @Schema(description = "Nombre del cliente", example = "Juan")
    @NotBlank(message = "El nombre es obligatorio")
    String nombre,
    @Schema(description = "Apellido del cliente", example = "Pérez")
    @NotBlank(message = "El apellido es obligatorio")
    String apellido,
    @Schema(description = "Email del cliente", example = "juan.perez@email.com")
    @Email(message = "El email no tiene un formato válido")
    String email
) {}
