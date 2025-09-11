package com.barber.barberBackend.service;

import java.time.LocalDateTime;
import java.time.LocalTime;
import java.util.List;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import com.barber.barberBackend.generics.GenericService;
import com.barber.barberBackend.model.Cliente;
import com.barber.barberBackend.model.Turno;
import com.barber.barberBackend.repository.IClienteRepository;
import com.barber.barberBackend.repository.ITurnoRepository;
@Service
public class TurnoService extends GenericService<Turno, Long, ITurnoRepository> implements ITurnoService {

    @Autowired
    ITurnoRepository repository;

    @Autowired
    IClienteRepository clienteRepository;

    @Override
    public void save(Turno entity) {
        try {
            Cliente cliente = entity.getCliente();
            
            // <-- Validaciones -->
            if (cliente != null && cliente.getTelefono() == null) {
                throw new IllegalArgumentException("El cliente debe tener un Teléfono antes de guardar el turno.");
            }
            // Validamos que la fecha y hora del turno sea válida
            validateDateTime(entity.getFechaHora());
            // Validamos que el turno no esté dado a otro cliente
            boolean turnoExists = repository.findDateTimes(entity.getFechaHora())
                                            .stream()
                                            .anyMatch(fechaHora -> fechaHora.equals(entity.getFechaHora()));
            if (turnoExists) {throw new IllegalArgumentException("Ya existe un turno para la fecha y hora especificada. Fecha y hora: " + entity.getFechaHora());}
            /*

                // Verificamos si el cliente ya existe
                boolean existe = clienteRepository.exists(org.springframework.data.domain.Example.of(cliente));            
                if (!existe) {
                    // Si el cliente no existe, lo guardamos
                    cliente = clienteRepository.save(cliente);
                }

                // Guardamos el turno
                entity.setCliente(cliente);
            */
            // Buscamos el cliente por teléfono
            Cliente existingCliente = clienteRepository.findByTelefono(cliente.getTelefono());

            if (existingCliente != null) {
                // Si el cliente existe, lo asignamos al turno
                entity.setCliente(existingCliente);
            } else {
                // Si el cliente no existe, lo guardamos
                cliente = clienteRepository.save(cliente);
                entity.setCliente(cliente);
            }
            
            repository.save(entity);

        } catch (Exception e) {
            throw new RuntimeException("Error al guardar el turno: " + e.getMessage(), e);
        }
    }

    @Override
    public void validateDateTime(LocalDateTime fechaHora) {
        try {
            LocalTime hora = fechaHora.toLocalTime();
            LocalTime horaApertura = LocalTime.of(13, 0);
            LocalTime horaCierre = LocalTime.of(20, 0);

            if (hora.isBefore(horaApertura) || hora.isAfter(horaCierre)) {
                throw new IllegalArgumentException("La hora debe estar entre " + horaApertura + " y " + horaCierre + ".");
            }
            if (hora.getMinute() != 0 && hora.getMinute() != 30) {
                throw new IllegalArgumentException("La hora debe ser en intervalos de 30 minutos.");
            }
        } catch (Exception e) {
            throw new RuntimeException("Error validadando la hora: " + e.getMessage(), e);
        }
    }

    @Override
    public List<LocalDateTime> findDateTimes() {
        try {
            LocalDateTime ahora = LocalDateTime.now();
            List<LocalDateTime> turnos = repository.findDateTimes(ahora);
            
            return turnos;
        } catch (Exception e) {
            throw new RuntimeException("Error retrieving date times: " + e.getMessage(), e);
        }
    }

    
    
}
