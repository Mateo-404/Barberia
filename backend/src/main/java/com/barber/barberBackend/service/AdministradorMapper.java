package com.barber.barberBackend.service;

import org.springframework.stereotype.Component;

import com.barber.barberBackend.dto.AdministradorRequestDTO;
import com.barber.barberBackend.dto.AdministradorResponseDTO;
import com.barber.barberBackend.model.Administrador;

@Component
public class AdministradorMapper {

    public AdministradorResponseDTO toResponseDTO(Administrador admin) {
        return new AdministradorResponseDTO(
            admin.getId(),
            admin.getNombre(),
            admin.getApellido(),
            admin.getEmail()
        );
    }

    public Administrador toEntity(AdministradorRequestDTO dto) {
        Administrador admin = new Administrador();
        admin.setNombre(dto.nombre());
        admin.setApellido(dto.apellido());
        admin.setEmail(dto.email());
        admin.setContrasenia(dto.contrasenia());
        return admin;
    }
}
