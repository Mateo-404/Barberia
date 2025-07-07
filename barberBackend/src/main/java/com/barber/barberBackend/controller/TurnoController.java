package com.barber.barberBackend.controller;

import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import com.barber.barberBackend.generics.GenericController;
import com.barber.barberBackend.model.Turno;
import com.barber.barberBackend.service.TurnoService;

@RestController
@RequestMapping("/turnos")
public class TurnoController extends GenericController<Turno, Long, TurnoService> {
    
}
