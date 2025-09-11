package com.barber.barberBackend.dto;

import java.util.List;

public record PanelEstadisticaDTO (
    int turnosHoy,
    int turnosAyer,
    double ingresosMes,
    double ingresosMesAnterior,
    long cantClientes,
    // 
    List<IngresoDiarioDTO> ingresosDiarios,
    List<ServicioEstadisticaDTO> servicios,
    List<HorarioEstadisticaDTO> horarios
    //List<ClienteFrecuenteDTO> clientesFrecuentes
) {}
