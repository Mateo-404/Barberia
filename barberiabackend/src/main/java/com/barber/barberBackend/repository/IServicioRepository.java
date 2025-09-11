package com.barber.barberBackend.repository;

import org.springframework.stereotype.Repository;

import com.barber.barberBackend.generics.GenericRepository;
import com.barber.barberBackend.model.Servicio;

@Repository
public interface IServicioRepository extends GenericRepository<Servicio, Long> {
}
