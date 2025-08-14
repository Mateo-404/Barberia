package com.barber.barberBackend.dto;

import java.time.LocalDateTime;

public record IngresoDiarioDTO (
    LocalDateTime fecha,
    double ingresoTotal
) {}
