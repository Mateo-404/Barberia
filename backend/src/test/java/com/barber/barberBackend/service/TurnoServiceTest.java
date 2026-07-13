package com.barber.barberBackend.service;

import com.barber.barberBackend.model.Cliente;
import com.barber.barberBackend.model.Servicio;
import com.barber.barberBackend.model.Turno;
import com.barber.barberBackend.repository.IClienteRepository;
import com.barber.barberBackend.repository.ITurnoRepository;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import java.time.LocalDateTime;
import java.util.List;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.ArgumentMatchers.anyString;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
class TurnoServiceTest {

    @Mock
    private ITurnoRepository repository;

    @Mock
    private IClienteRepository clienteRepository;

    @InjectMocks
    private TurnoService service;

    private final Cliente cliente = clienteConTelefono("123456789");
    private final Servicio servicio = new Servicio(1L, "Corte", 1500f);

    private static Cliente clienteConTelefono(String telefono) {
        Cliente c = new Cliente(telefono);
        c.setNombre("Juan");
        c.setApellido("Perez");
        c.setEmail("juan@test.com");
        return c;
    }

    @Test
    void save_withNewClient_createsClienteAndTurno() {
        Turno turno = new Turno(null, LocalDateTime.now().plusDays(1).withHour(15).withMinute(0), cliente, servicio);

        when(clienteRepository.findByTelefono(cliente.getTelefono())).thenReturn(null);
        when(clienteRepository.save(any(Cliente.class))).thenReturn(cliente);
        when(repository.findDateTimes(any(LocalDateTime.class))).thenReturn(List.of());

        service.save(turno);

        verify(clienteRepository).save(any(Cliente.class));
        verify(repository).save(turno);
    }

    @Test
    void save_withExistingClient_reusesClient() {
        Turno turno = new Turno(null, LocalDateTime.now().plusDays(1).withHour(15).withMinute(0), cliente, servicio);

        when(clienteRepository.findByTelefono(cliente.getTelefono())).thenReturn(cliente);
        when(repository.findDateTimes(any(LocalDateTime.class))).thenReturn(List.of());

        service.save(turno);

        verify(clienteRepository, never()).save(any(Cliente.class));
        verify(repository).save(turno);
    }

    @Test
    void save_whenClientHasNoTelefono_throwsException() {
        Cliente clienteSinTel = new Cliente(null);
        Turno turno = new Turno(null, LocalDateTime.now().plusDays(1).withHour(15).withMinute(0), clienteSinTel, servicio);

        Exception e = assertThrows(RuntimeException.class, () -> service.save(turno));
        assertTrue(e.getMessage().contains("Teléfono"));
    }

    @Test
    void save_whenDuplicateDateTime_throwsException() {
        LocalDateTime fechaHora = LocalDateTime.now().plusDays(1).withHour(15).withMinute(0);
        Turno turno = new Turno(null, fechaHora, cliente, servicio);

        when(repository.findDateTimes(any(LocalDateTime.class))).thenReturn(List.of(fechaHora));

        Exception e = assertThrows(RuntimeException.class, () -> service.save(turno));
        assertTrue(e.getMessage().contains("Ya existe un turno"));
    }

    @Test
    void save_whenTimeOutsideBusinessHours_throwsException() {
        Turno turno = new Turno(null, LocalDateTime.now().plusDays(1).withHour(10).withMinute(0), cliente, servicio);

        Exception e = assertThrows(RuntimeException.class, () -> service.save(turno));
        assertTrue(e.getMessage().contains("debe estar entre"));
    }

    @Test
    void save_whenInvalidMinuteInterval_throwsException() {
        Turno turno = new Turno(null, LocalDateTime.now().plusDays(1).withHour(15).withMinute(17), cliente, servicio);

        Exception e = assertThrows(RuntimeException.class, () -> service.save(turno));
        assertTrue(e.getMessage().contains("intervalos de 30 minutos"));
    }

    @Test
    void validateDateTime_withValidTime_doesNotThrow() {
        LocalDateTime valid = LocalDateTime.now().plusDays(1).withHour(15).withMinute(30);

        assertDoesNotThrow(() -> service.validateDateTime(valid));
    }

    @Test
    void validateDateTime_withHourBeforeOpening_throwsException() {
        LocalDateTime invalid = LocalDateTime.now().plusDays(1).withHour(12).withMinute(0);

        Exception e = assertThrows(RuntimeException.class, () -> service.validateDateTime(invalid));
        assertTrue(e.getMessage().contains("debe estar entre"));
    }

    @Test
    void validateDateTime_withHourAfterClosing_throwsException() {
        LocalDateTime invalid = LocalDateTime.now().plusDays(1).withHour(21).withMinute(0);

        Exception e = assertThrows(RuntimeException.class, () -> service.validateDateTime(invalid));
        assertTrue(e.getMessage().contains("debe estar entre"));
    }

    @Test
    void validateDateTime_withInvalidMinutes_throwsException() {
        LocalDateTime invalid = LocalDateTime.now().plusDays(1).withHour(15).withMinute(45);

        Exception e = assertThrows(RuntimeException.class, () -> service.validateDateTime(invalid));
        assertTrue(e.getMessage().contains("intervalos de 30 minutos"));
    }

    @Test
    void findDateTimes_returnsList() {
        LocalDateTime now = LocalDateTime.now();
        LocalDateTime future = now.plusDays(1).withHour(15).withMinute(0);
        when(repository.findDateTimes(any(LocalDateTime.class))).thenReturn(List.of(future));

        List<LocalDateTime> result = service.findDateTimes();

        assertFalse(result.isEmpty());
        assertEquals(1, result.size());
        assertEquals(future, result.get(0));
    }
}
