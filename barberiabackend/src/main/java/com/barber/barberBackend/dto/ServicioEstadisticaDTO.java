package com.barber.barberBackend.dto;

public record ServicioEstadisticaDTO (
    Long id,
    String nombre,
    long cantidadRealizado
) {}
