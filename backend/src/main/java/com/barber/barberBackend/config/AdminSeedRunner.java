package com.barber.barberBackend.config;

import org.springframework.beans.factory.annotation.Value;
import org.springframework.boot.CommandLineRunner;
import org.springframework.stereotype.Component;

import com.barber.barberBackend.model.Administrador;
import com.barber.barberBackend.repository.IAdministradorRepository;
import com.barber.barberBackend.service.AdministradorService;

import lombok.extern.slf4j.Slf4j;

@Slf4j
@Component
public class AdminSeedRunner implements CommandLineRunner {

    private final IAdministradorRepository repository;
    private final AdministradorService service;

    @Value("${admin.seed.email:admin@admin.com}")
    private String seedEmail;

    @Value("${admin.seed.password:admin1234}")
    private String seedPassword;

    public AdminSeedRunner(IAdministradorRepository repository, AdministradorService service) {
        this.repository = repository;
        this.service = service;
    }

    @Override
    public void run(String... args) {
        if (repository.count() > 0) {
            log.info("Ya existen administradores en la DB — se omite seed");
            return;
        }

        if (seedPassword.length() < 8) {
            log.error("ADMIN_SEED_PASSWORD debe tener al menos 8 caracteres (recibido: {}) — abortando seed", seedPassword.length());
            throw new IllegalStateException(
                "La contraseña del administrador semilla debe tener al menos 8 caracteres. "
                + "Configurá ADMIN_SEED_PASSWORD en tus variables de entorno.");
        }

        Administrador admin = new Administrador();
        admin.setNombre("Admin");
        admin.setApellido("Por Defecto");
        admin.setEmail(seedEmail);
        admin.setContrasenia(seedPassword);
        service.save(admin);

        log.info("Administrador semilla creado: email={}", seedEmail);
    }
}
