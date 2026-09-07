package com.barber.barberBackend.config;

import java.time.LocalDate;
import java.time.LocalDateTime;
import java.time.LocalTime;
import java.time.temporal.ChronoUnit;
import java.util.ArrayList;
import java.util.List;

import org.springframework.boot.CommandLineRunner;
import org.springframework.context.annotation.Profile;
import org.springframework.stereotype.Component;

import com.barber.barberBackend.model.Cliente;
import com.barber.barberBackend.model.Servicio;
import com.barber.barberBackend.model.Turno;
import com.barber.barberBackend.repository.IClienteRepository;
import com.barber.barberBackend.repository.IServicioRepository;
import com.barber.barberBackend.repository.ITurnoRepository;

import lombok.extern.slf4j.Slf4j;

@Slf4j
@Component
@Profile("h2")
public class DemoDataSeedRunner implements CommandLineRunner {

    private static final LocalTime[] SLOTS = {
        LocalTime.of(13, 0), LocalTime.of(14, 0), LocalTime.of(15, 0),
        LocalTime.of(16, 0), LocalTime.of(17, 0), LocalTime.of(18, 0),
        LocalTime.of(19, 0), LocalTime.of(19, 30)
    };

    private final IClienteRepository clienteRepository;
    private final IServicioRepository servicioRepository;
    private final ITurnoRepository turnoRepository;

    public DemoDataSeedRunner(IClienteRepository clienteRepository,
                              IServicioRepository servicioRepository,
                              ITurnoRepository turnoRepository) {
        this.clienteRepository = clienteRepository;
        this.servicioRepository = servicioRepository;
        this.turnoRepository = turnoRepository;
    }

    @Override
    public void run(String... args) {
        if (turnoRepository.count() > 0) {
            log.info("Ya existen turnos — se omite demo data");
            return;
        }

        List<Servicio> servicios = servicioRepository.findAll();
        if (servicios.isEmpty()) {
            String[][] defaults = {
                {"Corte clásico", "5000"},
                {"Corte + barba", "7000"},
                {"Afeitado", "3500"}
            };
            for (String[] s : defaults) {
                Servicio sv = new Servicio();
                sv.setTipo(s[0]);
                sv.setPrecio(Float.parseFloat(s[1]));
                servicios.add(servicioRepository.save(sv));
            }
            log.info("Servicios por defecto creados: {}", servicios.size());
        }

        List<Cliente> clientes = new ArrayList<>();
        String[][] datos = {
            {"1155551001", "Martín", "González"},
            {"1155551002", "Nadia", "Fernández"},
            {"1155551003", "Julián", "Rodríguez"},
            {"1155551004", "Carla", "Molina"},
            {"1155551005", "Diego", "Acosta"},
            {"1155551006", "Paula", "Sosa"},
            {"1155551007", "Tomás", "Rojas"},
            {"1155551008", "Florencia", "Domínguez"},
            {"1155551009", "Santiago", "Bravo"},
            {"1155551010", "Luciana", "Pereyra"}
        };
        for (String[] d : datos) {
            Cliente c = new Cliente();
            c.setTelefono(d[0]);
            c.setNombre(d[1]);
            c.setApellido(d[2]);
            clientes.add(c);
        }
        clienteRepository.saveAll(clientes);

        int total = 0;
        LocalDate hoy = LocalDate.now();

        // Histórico: 1-3 turnos por día los últimos 35 días (turnos pasados → gráficos)
        for (int diasAtras = 35; diasAtras >= 1; diasAtras--) {
            int porDia = 1 + (int) (Math.random() * 3);
            if (diasAtras > 1) {
                total += seedDia(hoy.minusDays(diasAtras), porDia, clientes, servicios);
            }
        }

        // Hoy y ayer con turnos adrede (salen en los contadores del panel)
        total += seedDia(hoy.minusDays(1), 3, clientes, servicios);
        total += seedDia(hoy, 4, clientes, servicios);

        // Próxima semana: algunos turnos futuros visibles en la agenda
        for (int i = 1; i <= 6; i++) {
            if (Math.random() < 0.6) {
                total += seedDia(hoy.plusDays(i), 1 + (int) (Math.random() * 2), clientes, servicios);
            }
        }

        log.info("Demo data sembrada: {} turnos, {} clientes", total, clientes.size());
    }

    private int seedDia(LocalDate fecha, int cantidad, List<Cliente> clientes, List<Servicio> servicios) {
        int n = 0;
        List<LocalTime> slots = new ArrayList<>(List.of(SLOTS));
        java.util.Collections.shuffle(slots);
        for (int i = 0; i < cantidad && i < slots.size(); i++) {
            Turno t = new Turno();
            t.setFechaHora(LocalDateTime.of(fecha, slots.get(i)));
            t.setCliente(clientes.get((int) (Math.random() * clientes.size())));
            t.setServicio(servicios.get((int) (Math.random() * servicios.size())));
            turnoRepository.save(t);
            n++;
        }
        return n;
    }
}