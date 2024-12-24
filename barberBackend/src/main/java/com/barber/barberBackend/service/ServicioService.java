package com.barber.barberBackend.service;

import java.util.List;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.HttpStatus;
import org.springframework.stereotype.Service;
import org.springframework.web.server.ResponseStatusException;

import com.barber.barberBackend.model.Servicio;
import com.barber.barberBackend.repository.IServicioRepository;

@Service
@jakarta.transaction.Transactional
public class ServicioService implements IServicioService {
    @Autowired
    private IServicioRepository repository;

    @Override
    public void borrarPorId(long id) {
        if (!repository.existsById(id)) {
            throw new ResponseStatusException(HttpStatus.NOT_FOUND, "No se encontro el servicio");
        }
        repository.deleteById(id);
    }

    @Override
    public Servicio buscarId(long id) {
        if (!repository.existsById(id)) {
            throw new ResponseStatusException(HttpStatus.NOT_FOUND, "No se encontro el servicio");
        } else {
            return repository.findById(id).get();
        }
    }

    @Override
    public Servicio buscarNombre(String nombre) {
        for (Servicio servicio : repository.findAll()) {
            if (servicio.getTipo().equals(nombre)) {
                return servicio;
            }
        }
        throw new ResponseStatusException(HttpStatus.NOT_FOUND, "No se encontro el servicio");
    }

    @Override
    public List<Servicio> buscarTodos() {
        return repository.findAll();
    }

    @Override
    public Servicio guardar(Servicio servicio) {
        return repository.save(servicio);
    }

    
}
