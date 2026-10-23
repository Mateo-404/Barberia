package com.barber.barberBackend.service;

import com.barber.barberBackend.dto.HorarioEstadisticaDTO;
import com.barber.barberBackend.dto.IngresoDiarioDTO;
import com.barber.barberBackend.dto.PanelEstadisticaDTO;
import com.barber.barberBackend.dto.ServicioEstadisticaDTO;
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
import static org.mockito.Mockito.when;

@ExtendWith(MockitoExtension.class)
class PanelEstadisticaDTOServiceTest {

    @Mock
    private ITurnoRepository turnoRepository;

    @Mock
    private IClienteRepository clienteRepository;

    @InjectMocks
    private PanelEstadisticaDTOService service;

    @Test
    void obtenerPanelEstadisticaDTO_returnsCompleteDTO() {
        LocalDateTime now = LocalDateTime.now();
        when(turnoRepository.countByFecha(any(LocalDateTime.class))).thenReturn(5);
        when(turnoRepository.findIngresosByFecha(any(LocalDateTime.class), any(LocalDateTime.class))).thenReturn(10000.0);
        when(clienteRepository.count()).thenReturn(50L);
        when(turnoRepository.countServiciosByFecha(any(LocalDateTime.class), any(LocalDateTime.class)))
                .thenReturn(List.of(new ServicioEstadisticaDTO(1L, "Corte", 10L)));
        when(turnoRepository.countTurnosByHorario(any(LocalDateTime.class)))
                .thenReturn(List.of(new HorarioEstadisticaDTO(15, 5)));

        PanelEstadisticaDTO result = service.obtenerPanelEstadisticaDTO();

        assertNotNull(result);
        assertEquals(5, result.turnosHoy());
        assertEquals(10000.0, result.ingresosMes(), 0.001);
        assertEquals(50, result.cantClientes());
        assertFalse(result.ingresosDiarios().isEmpty());
        assertFalse(result.servicios().isEmpty());
        assertFalse(result.horarios().isEmpty());
    }
}
