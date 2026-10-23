package com.barber.barberBackend.service;

import java.time.LocalDateTime;
import java.time.LocalTime;
import java.util.List;

import org.springframework.stereotype.Service;

import com.barber.barberBackend.exception.BusinessRuleViolationException;
import com.barber.barberBackend.generics.GenericService;
import com.barber.barberBackend.model.Cliente;
import com.barber.barberBackend.model.Turno;
import com.barber.barberBackend.repository.IClienteRepository;
import com.barber.barberBackend.repository.ITurnoRepository;
@Service
public class TurnoService extends GenericService<Turno, Long, ITurnoRepository> implements ITurnoService {

    private final ITurnoRepository turnoRepository;
    private final IClienteRepository clienteRepository;

    public TurnoService(ITurnoRepository repository, IClienteRepository clienteRepository) {
        super(repository);
        this.turnoRepository = repository;
        this.clienteRepository = clienteRepository;
    }

    @Override
    public Turno save(Turno entity) {
        Cliente cliente = entity.getCliente();

        if (cliente != null && cliente.getTelefono() == null) {
            throw new BusinessRuleViolationException("El cliente debe tener un Teléfono antes de guardar el turno.");
        }

        validateDateTime(entity.getFechaHora());

        boolean turnoExists = turnoRepository.findDateTimes(entity.getFechaHora())
                                        .stream()
                                        .anyMatch(fechaHora -> fechaHora.equals(entity.getFechaHora()));
        if (turnoExists) {
            throw new BusinessRuleViolationException("Ya existe un turno para la fecha y hora especificada. Fecha y hora: " + entity.getFechaHora());
        }

        Cliente existingCliente = clienteRepository.findByTelefono(cliente.getTelefono());

        if (existingCliente != null) {
            entity.setCliente(existingCliente);
        } else {
            cliente = clienteRepository.save(cliente);
            entity.setCliente(cliente);
        }

        return repository.save(entity);
    }

    @Override
    public void validateDateTime(LocalDateTime fechaHora) {
        LocalTime hora = fechaHora.toLocalTime();
        LocalTime horaApertura = LocalTime.of(13, 0);
        LocalTime horaCierre = LocalTime.of(20, 0);

        if (hora.isBefore(horaApertura) || hora.isAfter(horaCierre)) {
            throw new BusinessRuleViolationException("La hora debe estar entre " + horaApertura + " y " + horaCierre + ".");
        }
        if (hora.getMinute() != 0 && hora.getMinute() != 30) {
            throw new BusinessRuleViolationException("La hora debe ser en intervalos de 30 minutos.");
        }
    }

    @Override
    public List<LocalDateTime> findDateTimes() {
        LocalDateTime ahora = LocalDateTime.now();
        return turnoRepository.findDateTimes(ahora);
    }

    
    
}
