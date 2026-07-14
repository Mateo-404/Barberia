package com.barber.barberBackend.service;

import org.springframework.stereotype.Component;

import com.barber.barberBackend.dto.TurnoRequestDTO;
import com.barber.barberBackend.dto.TurnoResponseDTO;
import com.barber.barberBackend.exception.ResourceNotFoundException;
import com.barber.barberBackend.model.Cliente;
import com.barber.barberBackend.model.Servicio;
import com.barber.barberBackend.model.Turno;
import com.barber.barberBackend.repository.IServicioRepository;

/*
 * NOTA: Este mapper es diferente de ServicioMapper, ClienteMapper y
 * AdministradorMapper. A diferencia de aquellos (que solo transforman
 * campos entre DTO y entidad), TurnoMapper necesita:
 *
 * 1. Resolver referencias: como Turno tiene @ManyToOne a Servicio, el
 *    mapper inyecta IServicioRepository para buscar la entidad real
 *    por idServicio. Si no existe, lanza ResourceNotFoundException (404).
 *
 * 2. NO resolver Cliente: la lógica de "buscar o crear cliente" sigue
 *    en TurnoService.save(), que recibe un Turno con Cliente transient
 *    y decide si reusar el existente o persistir el nuevo. El mapper
 *    solo construye el Cliente transient con los campos del DTO.
 *
 * Si alguien agrega un mapper similar en el futuro que necesite repos,
 * este es el patrón a seguir. Si solo transforma campos (como los
 * otros tres mappers), NO debe inyectar repos.
 */
@Component
public class TurnoMapper {

    private final IServicioRepository servicioRepository;

    public TurnoMapper(IServicioRepository servicioRepository) {
        this.servicioRepository = servicioRepository;
    }

    public TurnoResponseDTO toResponseDTO(Turno entity) {
        Cliente cliente = entity.getCliente();
        Servicio servicio = entity.getServicio();
        return new TurnoResponseDTO(
            entity.getId(),
            entity.getFechaHora(),
            cliente != null ? cliente.getTelefono() : null,
            cliente != null ? cliente.getNombre() : null,
            cliente != null ? cliente.getApellido() : null,
            cliente != null ? cliente.getEmail() : null,
            servicio != null ? servicio.getId() : null,
            servicio != null ? servicio.getTipo() : null,
            servicio != null ? servicio.getPrecio() : 0
        );
    }

    public Turno toEntity(TurnoRequestDTO dto) {
        Servicio servicio = servicioRepository.findById(dto.idServicio())
            .orElseThrow(() -> new ResourceNotFoundException(
                "No existe un servicio con el id " + dto.idServicio()));

        Cliente cliente = new Cliente();
        cliente.setTelefono(dto.telefonoCliente());
        cliente.setNombre(dto.nombreCliente());
        cliente.setApellido(dto.apellidoCliente());
        cliente.setEmail(dto.emailCliente());

        return new Turno(null, dto.fechaHora(), cliente, servicio);
    }
}
