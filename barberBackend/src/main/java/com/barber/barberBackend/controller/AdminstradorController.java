package com.barber.barberBackend.controller;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import com.barber.barberBackend.generics.GenericController;
import com.barber.barberBackend.model.Administrador;
import com.barber.barberBackend.service.AdministradorService;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;


@RestController
@RequestMapping("/administradores")
public class AdminstradorController extends GenericController<Administrador, Long, AdministradorService> {
    @Autowired
    private AdministradorService service;

    //login
    @PostMapping("/login")
    public ResponseEntity<?> login(@RequestBody Administrador request) {
        try {
            Administrador admin = service.login(request.getEmail(), request.getContrasenia());
            return ResponseEntity.ok(admin);
        } catch (Exception e) {   
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED).body("Credenciales inválidas");
        }
    }
    
}
