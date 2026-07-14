package com.barber.barberBackend.service;

import org.springframework.stereotype.Component;

import com.barber.barberBackend.dto.ClienteRequestDTO;
import com.barber.barberBackend.dto.ClienteResponseDTO;
import com.barber.barberBackend.model.Cliente;

@Component
public class ClienteMapper {

    public ClienteResponseDTO toResponseDTO(Cliente entity) {
        return new ClienteResponseDTO(
            entity.getTelefono(),
            entity.getNombre(),
            entity.getApellido(),
            entity.getEmail()
        );
    }

    public Cliente toEntity(ClienteRequestDTO dto) {
        Cliente cliente = new Cliente();
        cliente.setTelefono(dto.telefono());
        cliente.setNombre(dto.nombre());
        cliente.setApellido(dto.apellido());
        cliente.setEmail(dto.email());
        return cliente;
    }
}
