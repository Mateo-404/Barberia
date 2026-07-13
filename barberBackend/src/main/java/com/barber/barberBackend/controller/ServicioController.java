package com.barber.barberBackend.controller;

import io.swagger.v3.oas.annotations.tags.Tag;

import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import com.barber.barberBackend.generics.GenericController;
import com.barber.barberBackend.model.Servicio;
import com.barber.barberBackend.service.ServicioService;

@RestController
@RequestMapping("/servicios")
@Tag(name = "Servicios", description = "Gestión de servicios ofrecidos")
public class ServicioController extends GenericController<Servicio, Long, ServicioService> {
}
