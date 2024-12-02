package com.barber.barberBackend.model;



abstract public class Persona {
    protected String nombre;
    protected String apellido;
    protected String contrasenia;
    protected String email;
    
    public Persona() {
    }

    public Persona(String nombre, String apellido, String contrasenia, String email) {
        this.nombre = nombre;
        this.apellido = apellido;
        this.contrasenia = contrasenia;
        this.email = email;
    }
    
    
}
