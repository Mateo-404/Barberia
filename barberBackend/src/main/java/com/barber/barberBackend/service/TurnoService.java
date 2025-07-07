package com.barber.barberBackend.service;

import org.springframework.stereotype.Service;

import com.barber.barberBackend.generics.GenericService;
import com.barber.barberBackend.model.Turno;
import com.barber.barberBackend.repository.ITurnoRepository;
@Service
public class TurnoService extends GenericService<Turno, Long, ITurnoRepository> implements ITurnoService {
    
}
