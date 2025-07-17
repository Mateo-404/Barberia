package com.barber.barberBackend.service;

import org.springframework.stereotype.Service;

import com.barber.barberBackend.generics.GenericService;
import com.barber.barberBackend.model.Cliente;
import com.barber.barberBackend.repository.IClienteRepository;

@Service
public class ClienteService extends GenericService<Cliente, String, IClienteRepository> implements IClienteService {

}
