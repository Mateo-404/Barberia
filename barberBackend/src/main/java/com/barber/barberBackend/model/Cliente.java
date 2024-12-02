package com.barber.barberBackend.model;

import jakarta.persistence.Entity;
import jakarta.persistence.Id;
import lombok.Getter;
import lombok.Setter;
@Getter
@Setter
@Entity
public class Cliente extends Persona {
    @Id
    private int telefono;

    public Cliente(String nombre, String apellido, String contrasenia, String email, int telefono, int dni) {
        super(nombre, apellido, contrasenia, email);
        this.telefono = telefono;
    }

    
}
