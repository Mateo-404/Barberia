package com.barber.barberBackend.controller;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import com.barber.barberBackend.dto.PanelEstadisticaDTO;
import com.barber.barberBackend.service.PanelEstadisticaDTOService;
import org.springframework.web.bind.annotation.GetMapping;


@RestController
@RequestMapping("/estadisticas")
public class PanelEstadisticasControllerDTO {
    @Autowired
    private PanelEstadisticaDTOService panelEstadisticaDTOService;

    @GetMapping("/panel")
    public PanelEstadisticaDTO estadisticas() {
        return panelEstadisticaDTOService.obtenerPanelEstadisticaDTO();
    }
    
    
}
