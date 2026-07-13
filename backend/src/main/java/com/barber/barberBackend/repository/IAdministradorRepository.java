package com.barber.barberBackend.repository;

import org.springframework.stereotype.Repository;

import com.barber.barberBackend.generics.GenericRepository;
import com.barber.barberBackend.model.Administrador;
@Repository
public interface IAdministradorRepository extends GenericRepository<Administrador, Long> {
    // Login
    Administrador findByEmail(String email);
}
