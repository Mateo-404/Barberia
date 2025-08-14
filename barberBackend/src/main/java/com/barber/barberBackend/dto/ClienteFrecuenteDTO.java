package com.barber.barberBackend.dto;

import com.barber.barberBackend.model.Cliente;

public record ClienteFrecuenteDTO (
    long id,
    String nombre,
    String apellido,
    long cantidadTurnos
) {}
