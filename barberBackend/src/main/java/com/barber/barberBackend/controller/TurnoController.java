package com.barber.barberBackend.controller;

import java.time.LocalDateTime;
import java.util.List;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import com.barber.barberBackend.generics.GenericController;
import com.barber.barberBackend.model.Turno;
import com.barber.barberBackend.service.TurnoService;
import org.springframework.web.bind.annotation.GetMapping;


@RestController
@RequestMapping("/turnos")
public class TurnoController extends GenericController<Turno, Long, TurnoService> {
    @Autowired
    private TurnoService service;

    @GetMapping("/findDateTimes")
    public String getMethodName() {
        List<LocalDateTime> dateTimes = service.findDateTimes();
        return dateTimes.toString();
    }
    
}
