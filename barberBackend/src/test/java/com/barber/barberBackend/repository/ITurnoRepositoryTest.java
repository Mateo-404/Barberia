package com.barber.barberBackend.repository;

import com.barber.barberBackend.dto.ServicioEstadisticaDTO;
import com.barber.barberBackend.dto.HorarioEstadisticaDTO;
import com.barber.barberBackend.model.Cliente;
import com.barber.barberBackend.model.Servicio;
import com.barber.barberBackend.model.Turno;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.autoconfigure.orm.jpa.DataJpaTest;
import org.springframework.test.context.ActiveProfiles;

import java.time.LocalDateTime;
import java.util.List;

import static org.junit.jupiter.api.Assertions.*;
import org.springframework.boot.test.autoconfigure.jdbc.AutoConfigureTestDatabase;
import static org.springframework.boot.test.autoconfigure.jdbc.AutoConfigureTestDatabase.Replace;

@DataJpaTest
@AutoConfigureTestDatabase(replace = Replace.NONE)
@ActiveProfiles("test")
class ITurnoRepositoryTest {

    @Autowired
    private ITurnoRepository turnoRepository;

    @Autowired
    private IClienteRepository clienteRepository;

    @Autowired
    private IServicioRepository servicioRepository;

    private Cliente cliente;
    private Servicio servicio;
    private Servicio servicio2;

    private LocalDateTime now() {
        return LocalDateTime.now().withNano(0);
    }

    @BeforeEach
    void setUp() {
        cliente = new Cliente();
        cliente.setTelefono("111111111");
        cliente.setNombre("Carlos");
        cliente.setApellido("Lopez");
        cliente.setEmail("carlos@test.com");
        clienteRepository.save(cliente);

        servicio = new Servicio(null, "Corte", 1500f);
        servicio2 = new Servicio(null, "Afeitado", 1000f);
        servicio = servicioRepository.save(servicio);
        servicio2 = servicioRepository.save(servicio2);
    }

    @Test
    void countByFecha_returnsCorrectCount() {
        LocalDateTime now = now();
        Turno turno = new Turno(null, now, cliente, servicio);
        turnoRepository.save(turno);

        int count = turnoRepository.countByFecha(now);

        assertEquals(1, count);
    }

    @Test
    void countByFecha_whenNoTurnos_returnsZero() {
        int count = turnoRepository.countByFecha(now().plusDays(10));

        assertEquals(0, count);
    }

    @Test
    void findIngresosByFecha_returnsSum() {
        LocalDateTime now = now();
        Turno t1 = new Turno(null, now, cliente, servicio);
        Turno t2 = new Turno(null, now, cliente, servicio2);
        turnoRepository.save(t1);
        turnoRepository.save(t2);

        LocalDateTime inicio = now.withHour(0).withMinute(0).withSecond(0);
        LocalDateTime fin = inicio.plusDays(1);
        Double ingresos = turnoRepository.findIngresosByFecha(inicio, fin);

        assertEquals(2500.0, ingresos, 0.001);
    }

    @Test
    void findIngresosByFecha_whenNoData_returnsZero() {
        Double ingresos = turnoRepository.findIngresosByFecha(
                now().minusDays(10), now().minusDays(9));

        assertEquals(0.0, ingresos, 0.001);
    }

    @Test
    void findDateTimes_returnsFutureDates() {
        LocalDateTime future = now().plusDays(1);
        Turno turno = new Turno(null, future, cliente, servicio);
        turnoRepository.save(turno);

        List<LocalDateTime> dateTimes = turnoRepository.findDateTimes(now());

        assertFalse(dateTimes.isEmpty());
        assertTrue(dateTimes.contains(future));
    }

    @Test
    void countServiciosByFecha_returnsGroupedCounts() {
        LocalDateTime now = now();
        Turno t1 = new Turno(null, now, cliente, servicio);
        Turno t2 = new Turno(null, now, cliente, servicio);
        Turno t3 = new Turno(null, now, cliente, servicio2);
        turnoRepository.save(t1);
        turnoRepository.save(t2);
        turnoRepository.save(t3);

        LocalDateTime desde = now.minusDays(1);
        LocalDateTime hasta = now.plusDays(1);
        List<ServicioEstadisticaDTO> servicios = turnoRepository.countServiciosByFecha(desde, hasta);

        assertEquals(2, servicios.size());
        assertEquals(2, servicios.get(0).cantidadRealizado());
    }

    @Test
    void countTurnosByHorario_returnsHourlyCounts() {
        LocalDateTime now = now();
        Turno t1 = new Turno(null, now.withHour(14).withMinute(0), cliente, servicio);
        Turno t2 = new Turno(null, now.withHour(14).withMinute(30), cliente, servicio);
        turnoRepository.save(t1);
        turnoRepository.save(t2);

        LocalDateTime desde = now.withDayOfYear(1).withYear(now.getYear());
        List<HorarioEstadisticaDTO> horarios = turnoRepository.countTurnosByHorario(desde);

        assertFalse(horarios.isEmpty());
        HorarioEstadisticaDTO hora14 = horarios.stream()
                .filter(h -> h.hora() == 14)
                .findFirst().orElse(null);
        assertNotNull(hora14);
        assertEquals(2, hora14.cantidadRealizado());
    }
}
