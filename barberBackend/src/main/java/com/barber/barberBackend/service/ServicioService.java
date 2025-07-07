package com.barber.barberBackend.service;

import org.springframework.stereotype.Service;

import com.barber.barberBackend.generics.GenericService;
import com.barber.barberBackend.model.Servicio;
import com.barber.barberBackend.repository.IServicioRepository;

@Service
public class ServicioService extends GenericService<Servicio, Long, IServicioRepository> implements IServicioService {
}
