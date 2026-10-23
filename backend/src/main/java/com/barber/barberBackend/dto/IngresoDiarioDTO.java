package com.barber.barberBackend.dto;

import io.swagger.v3.oas.annotations.media.Schema;

import java.time.LocalDateTime;

@Schema(description = "Ingreso total registrado en un día específico")
public record IngresoDiarioDTO (
    @Schema(description = "Fecha del ingreso") LocalDateTime fecha,
    @Schema(description = "Monto total ingresado en el día") double ingresoTotal
) {}
