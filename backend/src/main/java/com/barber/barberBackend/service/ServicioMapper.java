package com.barber.barberBackend.service;

import org.springframework.stereotype.Component;

import com.barber.barberBackend.dto.ServicioRequestDTO;
import com.barber.barberBackend.dto.ServicioResponseDTO;
import com.barber.barberBackend.model.Servicio;

@Component
public class ServicioMapper {

    public ServicioResponseDTO toResponseDTO(Servicio entity) {
        return new ServicioResponseDTO(entity.getId(), entity.getTipo(), entity.getPrecio());
    }

    public Servicio toEntity(ServicioRequestDTO dto) {
        return new Servicio(null, dto.tipo(), dto.precio());
    }
}
