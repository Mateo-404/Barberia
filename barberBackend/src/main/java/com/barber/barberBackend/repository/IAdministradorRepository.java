package com.barber.barberBackend.repository;

import org.springframework.data.jpa.repository.Query;
import org.springframework.stereotype.Repository;

import com.barber.barberBackend.generics.GenericRepository;
import com.barber.barberBackend.model.Administrador;
@Repository
public interface IAdministradorRepository extends GenericRepository<Administrador, Long> {
    // Login
    @Query("SELECT a FROM Administrador a WHERE a.email = ?1 AND a.contrasenia = ?2")
    Administrador findByEmailAndContrasenia(String email, String contrasenia);
}
