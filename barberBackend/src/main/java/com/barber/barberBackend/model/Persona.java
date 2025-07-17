package com.barber.barberBackend.model;

import jakarta.persistence.MappedSuperclass;
import lombok.Getter;
import lombok.Setter;

@MappedSuperclass
@Getter
@Setter
abstract public class Persona {
    protected String nombre;
    protected String apellido;
    protected String email;       
}
