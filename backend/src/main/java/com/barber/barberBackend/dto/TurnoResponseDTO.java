package com.barber.barberBackend.dto;

import java.time.LocalDateTime;

import io.swagger.v3.oas.annotations.media.Schema;

@Schema(description = "Turno registrado con datos del cliente y servicio")
public record TurnoResponseDTO(
    @Schema(description = "ID único del turno", example = "1")
    Long id,
    @Schema(description = "Fecha y hora del turno", example = "2026-07-15T14:00:00")
    LocalDateTime fechaHora,
    @Schema(description = "Teléfono del cliente", example = "1155551234")
    String telefonoCliente,
    @Schema(description = "Nombre del cliente", example = "Juan")
    String nombreCliente,
    @Schema(description = "Apellido del cliente", example = "Pérez")
    String apellidoCliente,
    @Schema(description = "Email del cliente", example = "juan.perez@email.com")
    String emailCliente,
    @Schema(description = "ID del servicio reservado", example = "1")
    Long idServicio,
    @Schema(description = "Nombre del servicio", example = "Corte de pelo")
    String tipoServicio,
    @Schema(description = "Precio del servicio", example = "1500.0")
    float precioServicio
) {}
