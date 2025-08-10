package com.barber.barberBackend.service;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import com.barber.barberBackend.generics.GenericService;
import com.barber.barberBackend.model.Administrador;
import com.barber.barberBackend.repository.IAdministradorRepository;
@Service
public class AdministradorService extends GenericService<Administrador, Long, IAdministradorRepository> implements IAdministradorService {

    @Autowired
    private IAdministradorRepository repository;

    @Override
    public Administrador login(String email, String contrasenia) {
        try {
            if (email == null || contrasenia == null) {
                throw new IllegalArgumentException("El email y la contraseña no pueden ser nulos");
            }

            Administrador administrador = repository.findByEmailAndContrasenia(email, contrasenia);

            if (administrador == null) {
                throw new IllegalArgumentException("Credenciales inválidas");
            }

            return administrador;
        } catch (Exception e) {
            throw new RuntimeException("Error al realizar el login: " + e.getMessage(), e);
        }
    }
    
}
