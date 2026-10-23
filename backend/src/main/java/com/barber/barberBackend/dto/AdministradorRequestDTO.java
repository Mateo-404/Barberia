package com.barber.barberBackend.dto;

import io.swagger.v3.oas.annotations.media.Schema;

import jakarta.validation.constraints.Email;
import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.Size;

@Schema(description = "Datos para crear un nuevo administrador")
public record AdministradorRequestDTO(
    @Schema(description = "Nombre del administrador", example = "Carlos")
    @NotBlank(message = "El nombre es obligatorio")
    String nombre,
    @Schema(description = "Apellido del administrador", example = "López")
    @NotBlank(message = "El apellido es obligatorio")
    String apellido,
    @Schema(description = "Email del administrador (usado para iniciar sesión)", example = "carlos.lopez@barberia.com")
    @NotBlank(message = "El email es obligatorio")
    @Email(message = "El email no tiene un formato válido")
    String email,
    @Schema(description = "Contraseña en texto plano (se hashea antes de persistir)", example = "miPassword123")
    @NotBlank(message = "La contraseña es obligatoria")
    @Size(min = 8, message = "La contraseña debe tener al menos 8 caracteres")
    String contrasenia
) {}
