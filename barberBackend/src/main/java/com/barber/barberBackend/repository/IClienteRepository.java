package com.barber.barberBackend.repository;

import org.springframework.stereotype.Repository;

import com.barber.barberBackend.generics.GenericRepository;
import com.barber.barberBackend.model.Cliente;

@Repository
public interface IClienteRepository extends GenericRepository<Cliente, Integer> {
    
}
