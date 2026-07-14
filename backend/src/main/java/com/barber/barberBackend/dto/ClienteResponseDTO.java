package com.barber.barberBackend.dto;

public record ClienteResponseDTO(
    String telefono,
    String nombre,
    String apellido,
    String email
) {}
