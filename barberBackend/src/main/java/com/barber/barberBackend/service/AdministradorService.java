package com.barber.barberBackend.service;

import org.springframework.stereotype.Service;

import com.barber.barberBackend.generics.GenericService;
import com.barber.barberBackend.model.Administrador;
import com.barber.barberBackend.repository.IAdministradorRepository;
import org.springframework.security.crypto.argon2.Argon2PasswordEncoder;

@Service
public class AdministradorService extends GenericService<Administrador, Long, IAdministradorRepository> implements IAdministradorService {

    private final Argon2PasswordEncoder passwordEncoder;
    private final IAdministradorRepository adminRepository;

    public AdministradorService(Argon2PasswordEncoder passwordEncoder, IAdministradorRepository repository) {
        super(repository);
        this.passwordEncoder = passwordEncoder;
        this.adminRepository = repository;
    }

    @Override
    public Administrador save(Administrador admin) {
        admin.setContrasenia(passwordEncoder.encode(admin.getContrasenia()));
        return super.save(admin);
    }

    @Override
    public Administrador login(String email, String contrasenia) {
        if (email == null || contrasenia == null) {
            throw new IllegalArgumentException("El email y la contraseña no pueden ser nulos");
        }

        Administrador administrador = adminRepository.findByEmail(email);

        if (administrador == null || !passwordEncoder.matches(contrasenia, administrador.getContrasenia())) {
            throw new IllegalArgumentException("Credenciales inválidas");
        }

        return administrador;
    }

}
