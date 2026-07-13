package com.barber.barberBackend.dto;

import io.swagger.v3.oas.annotations.media.Schema;

import lombok.Getter;
import lombok.NoArgsConstructor;
import lombok.Setter;

@NoArgsConstructor
@Setter
@Getter
@Schema(description = "Cliente con mayor frecuencia de turnos")
public class ClienteFrecuenteDTO {
    @Schema(description = "ID del cliente")
    private Long id;

    @Schema(description = "Nombre del cliente", example = "Juan")
    private String nombre;

    @Schema(description = "Apellido del cliente", example = "Pérez")
    private String apellido;

    @Schema(description = "Cantidad de turnos realizados")
    private Long cantidadTurnos;

    public ClienteFrecuenteDTO(Long id, String nombre, String apellido, Long cantidadTurnos) {
        this.id = id;
        this.nombre = nombre;
        this.apellido = apellido;
        this.cantidadTurnos = cantidadTurnos;
    }
}
