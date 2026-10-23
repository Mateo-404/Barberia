package com.barber.barberBackend.dto;

import io.swagger.v3.oas.annotations.media.Schema;

import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.Size;

@Schema(description = "Credenciales de inicio de sesión para administradores")
public record LoginRequestDTO(
    @Schema(description = "Email del administrador", example = "admin@barberia.com")
    @NotBlank(message = "El email es obligatorio")
    String email,
    @Schema(description = "Contraseña del administrador", example = "miPassword123")
    @NotBlank(message = "La contraseña es obligatoria")
    @Size(min = 8, message = "La contraseña debe tener al menos 8 caracteres")
    String contrasenia
) {}
