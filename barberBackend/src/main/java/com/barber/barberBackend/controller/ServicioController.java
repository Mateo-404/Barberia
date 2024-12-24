package com.barber.barberBackend.controller;

import java.util.List;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.DeleteMapping;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

import com.barber.barberBackend.model.Servicio;
import com.barber.barberBackend.service.IServicioService;

import jakarta.websocket.server.PathParam;

@RestController
@RequestMapping("/servicio")
public class ServicioController{
    @Autowired
    private IServicioService servicioService;

    @DeleteMapping("/borrarPorId/{id}")
    public void borrarPorId(@PathVariable long id) {
        servicioService.borrarPorId(id);
    }

    @GetMapping("/buscarId/{id}")
    public Servicio buscarId(@PathVariable long id) {
        return servicioService.buscarId(id);
    }

    @GetMapping("/buscarNombre/{nombre}")
    public Servicio buscarNombre(@PathVariable String nombre) {
        return servicioService.buscarNombre(nombre);
    }

    @GetMapping("/buscarTodos")
    public List<Servicio> buscarTodos() {
        return servicioService.buscarTodos();
    }

    @PostMapping("/guardar")
    public Servicio guardar(@RequestBody Servicio servicio) {
        return servicioService.guardar(servicio);
    }

    
}
