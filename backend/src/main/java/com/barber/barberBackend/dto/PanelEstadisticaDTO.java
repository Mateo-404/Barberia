package com.barber.barberBackend.dto;

import io.swagger.v3.oas.annotations.media.Schema;

import java.util.List;

@Schema(description = "Panel completo de estadísticas del negocio")
public record PanelEstadisticaDTO (
    @Schema(description = "Cantidad de turnos para el día de hoy") int turnosHoy,
    @Schema(description = "Cantidad de turnos para el día de ayer") int turnosAyer,
    @Schema(description = "Ingresos totales del mes actual") double ingresosMes,
    @Schema(description = "Ingresos totales del mes anterior") double ingresosMesAnterior,
    @Schema(description = "Cantidad total de clientes registrados") long cantClientes,
    @Schema(description = "Ingresos diarios de los últimos 30 días") List<IngresoDiarioDTO> ingresosDiarios,
    @Schema(description = "Servicios más realizados en el último mes") List<ServicioEstadisticaDTO> servicios,
    @Schema(description = "Horarios con más turnos en lo que va del año") List<HorarioEstadisticaDTO> horarios
) {}
