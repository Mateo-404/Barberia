package com.barber.barberBackend.dto;

import java.time.LocalDateTime;

import jakarta.validation.constraints.Email;
import jakarta.validation.constraints.Future;
import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;

public record TurnoRequestDTO(
    @NotNull(message = "La fecha y hora es obligatoria")
    @Future(message = "La fecha debe ser futura")
    LocalDateTime fechaHora,

    @NotNull(message = "El servicio es obligatorio")
    Long idServicio,

    @NotBlank(message = "El teléfono del cliente es obligatorio")
    String telefonoCliente,

    @NotBlank(message = "El nombre del cliente es obligatorio")
    String nombreCliente,

    @NotBlank(message = "El apellido del cliente es obligatorio")
    String apellidoCliente,

    @Email(message = "El email del cliente no tiene un formato válido")
    String emailCliente
) {}
