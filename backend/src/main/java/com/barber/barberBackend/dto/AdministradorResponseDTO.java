package com.barber.barberBackend.dto;

public record AdministradorResponseDTO(
    Long id,
    String nombre,
    String apellido,
    String email
) {}
