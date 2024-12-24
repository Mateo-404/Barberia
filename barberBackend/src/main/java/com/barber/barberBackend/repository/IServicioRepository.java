package com.barber.barberBackend.repository;

import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import com.barber.barberBackend.model.Servicio;

@Repository
public interface IServicioRepository extends JpaRepository<Servicio, Long> {
}
