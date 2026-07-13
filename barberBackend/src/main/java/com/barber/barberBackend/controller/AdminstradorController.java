package com.barber.barberBackend.controller;

import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.media.Content;
import io.swagger.v3.oas.annotations.responses.ApiResponse;
import io.swagger.v3.oas.annotations.responses.ApiResponses;
import io.swagger.v3.oas.annotations.tags.Tag;

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
@Tag(name = "Administradores", description = "Gestión de administradores")
public class AdminstradorController extends GenericController<Administrador, Long, AdministradorService> {
    private final AdministradorService service;

    public AdminstradorController(AdministradorService service) {
        this.service = service;
    }

    @Operation(summary = "Iniciar sesión", description = "Autentica un administrador por email y contraseña")
    @ApiResponses({
        @ApiResponse(responseCode = "200", description = "Inicio de sesión exitoso"),
        @ApiResponse(responseCode = "401", description = "Credenciales inválidas", content = @Content)
    })
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
