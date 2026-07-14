package com.barber.barberBackend.dto;

import java.time.LocalDateTime;

import io.swagger.v3.oas.annotations.media.Schema;

import jakarta.validation.constraints.Email;
import jakarta.validation.constraints.Future;
import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;

@Schema(description = "Datos para crear un nuevo turno (incluye datos del cliente)")
public record TurnoRequestDTO(
    @Schema(description = "Fecha y hora del turno", example = "2026-07-15T14:00:00")
    @NotNull(message = "La fecha y hora es obligatoria")
    @Future(message = "La fecha debe ser futura")
    LocalDateTime fechaHora,
    @Schema(description = "ID del servicio solicitado", example = "1")
    @NotNull(message = "El servicio es obligatorio")
    Long idServicio,
    @Schema(description = "Teléfono del cliente (si no existe, se crea)", example = "1155551234")
    @NotBlank(message = "El teléfono del cliente es obligatorio")
    String telefonoCliente,
    @Schema(description = "Nombre del cliente", example = "Juan")
    @NotBlank(message = "El nombre del cliente es obligatorio")
    String nombreCliente,
    @Schema(description = "Apellido del cliente", example = "Pérez")
    @NotBlank(message = "El apellido del cliente es obligatorio")
    String apellidoCliente,
    @Schema(description = "Email del cliente (opcional)", example = "juan.perez@email.com")
    @Email(message = "El email del cliente no tiene un formato válido")
    String emailCliente
) {}
