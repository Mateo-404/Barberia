package com.barber.barberBackend.dto;

import java.time.LocalDateTime;

public record HorarioEstadisticaDTO (
    LocalDateTime hora,
    long cantidadRealizado
) {}
