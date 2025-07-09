package com.barber.barberBackend.repository;

import org.springframework.stereotype.Repository;

import com.barber.barberBackend.generics.GenericRepository;
import com.barber.barberBackend.model.Turno;

@Repository
public interface ITurnoRepository extends GenericRepository<Turno, Long> {
     
}
