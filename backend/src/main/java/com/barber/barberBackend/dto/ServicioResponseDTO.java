package com.barber.barberBackend.dto;

import io.swagger.v3.oas.annotations.media.Schema;

@Schema(description = "Servicio disponible en la barbería")
public record ServicioResponseDTO(
    @Schema(description = "ID único del servicio", example = "1")
    Long id,
    @Schema(description = "Nombre o tipo de servicio", example = "Corte de pelo")
    String tipo,
    @Schema(description = "Precio del servicio en pesos", example = "1500.0")
    float precio
) {}
