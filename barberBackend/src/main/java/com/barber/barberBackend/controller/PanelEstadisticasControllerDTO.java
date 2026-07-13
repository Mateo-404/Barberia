package com.barber.barberBackend.controller;

import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import com.barber.barberBackend.dto.PanelEstadisticaDTO;
import com.barber.barberBackend.service.PanelEstadisticaDTOService;
import org.springframework.web.bind.annotation.GetMapping;

@RestController
@RequestMapping("/estadisticas")
@Tag(name = "Estadísticas", description = "Panel de estadísticas del negocio")
public class PanelEstadisticasControllerDTO {
    @Autowired
    private PanelEstadisticaDTOService panelEstadisticaDTOService;

    @Operation(summary = "Obtener panel de estadísticas", description = "Devuelve turnos hoy, ayer, ingresos del mes, servicios más pedidos y horarios con más actividad")
    @GetMapping("/panel")
    public PanelEstadisticaDTO estadisticas() {
        return panelEstadisticaDTOService.obtenerPanelEstadisticaDTO();
    }
}
