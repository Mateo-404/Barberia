package com.barber.barberBackend.dto;

import java.time.LocalDateTime;

public record HorarioEstadisticaDTO (
    int hora,
    long cantidadRealizado
) {}
