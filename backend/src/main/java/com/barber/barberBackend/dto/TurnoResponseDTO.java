package com.barber.barberBackend.dto;

import java.time.LocalDateTime;

public record TurnoResponseDTO(
    Long id,
    LocalDateTime fechaHora,
    String telefonoCliente,
    String nombreCliente,
    String apellidoCliente,
    String emailCliente,
    Long idServicio,
    String tipoServicio,
    float precioServicio
) {}
