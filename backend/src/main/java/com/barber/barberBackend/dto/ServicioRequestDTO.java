package com.barber.barberBackend.dto;

import io.swagger.v3.oas.annotations.media.Schema;

import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.Positive;

@Schema(description = "Datos para crear o actualizar un servicio")
public record ServicioRequestDTO(
    @Schema(description = "Nombre o tipo de servicio ofrecido", example = "Corte de pelo")
    @NotBlank(message = "El tipo de servicio es obligatorio")
    String tipo,
    @Schema(description = "Precio del servicio en pesos", example = "1500.0")
    @Positive(message = "El precio debe ser positivo")
    float precio
) {}
