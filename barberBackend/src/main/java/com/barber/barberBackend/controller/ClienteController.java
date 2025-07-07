package com.barber.barberBackend.controller;

import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import com.barber.barberBackend.generics.GenericController;
import com.barber.barberBackend.model.Cliente;
import com.barber.barberBackend.service.ClienteService;

@RestController
@RequestMapping("/clientes")
public class ClienteController extends GenericController<Cliente, Integer, ClienteService> {
    
}
