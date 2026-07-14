package com.barber.barberBackend.dto;

import jakarta.validation.constraints.Email;
import jakarta.validation.constraints.NotBlank;

public record ClienteRequestDTO(
    @NotBlank(message = "El teléfono es obligatorio")
    String telefono,
    @NotBlank(message = "El nombre es obligatorio")
    String nombre,
    @NotBlank(message = "El apellido es obligatorio")
    String apellido,
    @Email(message = "El email no tiene un formato válido")
    String email
) {}
