package com.barber.barberBackend.service;

import org.springframework.stereotype.Component;

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
}
