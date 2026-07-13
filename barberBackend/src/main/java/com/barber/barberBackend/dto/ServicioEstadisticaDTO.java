package com.barber.barberBackend.dto;

import io.swagger.v3.oas.annotations.media.Schema;

@Schema(description = "Estadística de un servicio: cantidad de veces realizado en un período")
public record ServicioEstadisticaDTO (
    @Schema(description = "ID del servicio") Long id,
    @Schema(description = "Nombre o tipo del servicio") String nombre,
    @Schema(description = "Cantidad de veces que se realizó el servicio") long cantidadRealizado
) {}
