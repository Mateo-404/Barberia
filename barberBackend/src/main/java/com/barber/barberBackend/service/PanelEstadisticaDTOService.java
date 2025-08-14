package com.barber.barberBackend.service;

import java.time.LocalDateTime;
import java.util.ArrayList;
import java.util.List;

import org.springframework.stereotype.Service;

import com.barber.barberBackend.dto.ClienteFrecuenteDTO;
import com.barber.barberBackend.dto.HorarioEstadisticaDTO;
import com.barber.barberBackend.dto.IngresoDiarioDTO;
import com.barber.barberBackend.dto.PanelEstadisticaDTO;
import com.barber.barberBackend.dto.ServicioEstadisticaDTO;
import com.barber.barberBackend.repository.IClienteRepository;
import com.barber.barberBackend.repository.IServicioRepository;
import com.barber.barberBackend.repository.ITurnoRepository;


import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class PanelEstadisticaDTOService implements IPanelEstadisticaDTO {
    private final ITurnoRepository turnoRepository;
    private final IClienteRepository clienteRepository;

    @Override
    public PanelEstadisticaDTO obtenerPanelEstadisticaDTO() {
        try {
            final LocalDateTime fechaActual = LocalDateTime.now();
            int turnosHoy = turnoRepository.countByFecha(fechaActual);
            int turnosAyer = turnoRepository.countByFecha(fechaActual.minusDays(1));
            double ingresosMes = turnoRepository.findIngresosByFecha(
                fechaActual.withDayOfMonth(1),
                fechaActual.plusMonths(1).withDayOfMonth(1)
            );
            double ingresosMesAnterior = turnoRepository.findIngresosByFecha(
                fechaActual.minusMonths(1).withDayOfMonth(1),
                fechaActual.withDayOfMonth(1)
            );
            long cantClientes = clienteRepository.count();

            List<IngresoDiarioDTO> ingresosDiarios = new ArrayList<>();
            LocalDateTime fechaLimite = fechaActual.minusMonths(1);
            for (int i = 0; ; i++) {
                LocalDateTime fecha = fechaActual.minusDays(i).withHour(0).withMinute(0).withSecond(0).withNano(0);
                if (fecha.isBefore(fechaLimite)) break;

                LocalDateTime inicio = fecha;
                LocalDateTime fin = fecha.plusDays(1);

                double ingresoTotal = turnoRepository.findIngresosByFecha(inicio, fin);
                ingresosDiarios.add(new IngresoDiarioDTO(fecha, ingresoTotal));
            }
            
            List<ServicioEstadisticaDTO> servicios = new ArrayList<>();
            servicios = turnoRepository.countServiciosByFecha(
                fechaActual.minusMonths(1).withHour(0).withMinute(0).withSecond(0).withNano(0),
                fechaActual.plusDays(1).withHour(0).withMinute(0).withSecond(0).withNano(0)
            );

            List<HorarioEstadisticaDTO> horarios = new ArrayList<>();
            horarios = turnoRepository.countTurnosByHorario(
                //Año actual
                fechaActual.withYear(fechaActual.getYear()).withMonth(1).withDayOfMonth(1)
            );

            List<ClienteFrecuenteDTO> clientesFrecuentes = new ArrayList<>();
            clientesFrecuentes = turnoRepository.findClientesFrecuentes(
                //Año actual
                fechaActual.withYear(fechaActual.getYear()).withMonth(1).withDayOfMonth(1)
            );

            if (ingresosDiarios == null || servicios == null || horarios == null || clientesFrecuentes == null) {
                throw new IllegalStateException("ERROR: Algun valor de los atributos es nulo");
            }

            return new PanelEstadisticaDTO(
                turnosHoy,
                turnosAyer,
                ingresosMes,
                ingresosMesAnterior,
                cantClientes,
                ingresosDiarios,
                servicios,
                horarios,
                clientesFrecuentes
            );            
        } catch (Exception e) {
            e.printStackTrace();
            return null;
        }
    }
    
}
