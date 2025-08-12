package com.barber.barberBackend.dto;

public record PanelEstadisticaDTO (
    int turnosHoy,
    int turnosAyer,
    double ingresosMes,
    double ingresosMesAnterior,
    int cantClientes,
    /*
     * 
     List<IngresoDiarioDTO> ingresosDiarios,
     List<ServicioEstadisticaDTO> servicios,
     List<HorarioEstadisticaDTO> horarios,
     List<ClienteFrecuenteDTO> clientesFrecuentes,
     */
    double ocupacionPromedio
) {  
}
