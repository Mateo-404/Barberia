package com.barber.barberBackend.dto;

import lombok.AllArgsConstructor;
import lombok.Getter;
import lombok.NoArgsConstructor;
import lombok.Setter;


@NoArgsConstructor
@Setter
@Getter
public class ClienteFrecuenteDTO {
    private Long id;
    private String nombre;
    private String apellido;
    private Long cantidadTurnos;

    public ClienteFrecuenteDTO(Long id, String nombre, String apellido, Long cantidadTurnos) {
        this.id = id;
        this.nombre = nombre;
        this.apellido = apellido;
        this.cantidadTurnos = cantidadTurnos;
    }
}
