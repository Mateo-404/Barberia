package com.barber.barberBackend.dto;

import io.swagger.v3.oas.annotations.media.Schema;

@Schema(description = "Cantidad de turnos realizados en una hora específica")
public record HorarioEstadisticaDTO (
    @Schema(description = "Hora del día (0-23)", example = "14") int hora,
    @Schema(description = "Cantidad de turnos realizados en esa hora") long cantidadRealizado
) {}
