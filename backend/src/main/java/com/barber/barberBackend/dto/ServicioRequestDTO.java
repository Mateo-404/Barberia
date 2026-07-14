package com.barber.barberBackend.dto;

import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.Positive;

public record ServicioRequestDTO(
    @NotBlank(message = "El tipo de servicio es obligatorio")
    String tipo,
    @Positive(message = "El precio debe ser positivo")
    float precio
) {}
