package com.barber.barberBackend.controller;

import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.media.Content;
import io.swagger.v3.oas.annotations.responses.ApiResponse;
import io.swagger.v3.oas.annotations.responses.ApiResponses;
import io.swagger.v3.oas.annotations.tags.Tag;

import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import com.barber.barberBackend.dto.AdministradorResponseDTO;
import com.barber.barberBackend.generics.GenericController;
import com.barber.barberBackend.model.Administrador;
import com.barber.barberBackend.service.AdministradorMapper;
import com.barber.barberBackend.service.AdministradorService;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;

@RestController
@RequestMapping("/administradores")
@Tag(name = "Administradores", description = "Gestión de administradores")
public class AdminstradorController extends GenericController<Administrador, AdministradorResponseDTO, Long, AdministradorService> {
    private final AdministradorService service;
    private final AdministradorMapper mapper;

    public AdminstradorController(AdministradorService service, AdministradorMapper mapper) {
        this.service = service;
        this.mapper = mapper;
    }

    @Override
    protected AdministradorResponseDTO toDTO(Administrador entity) {
        return mapper.toResponseDTO(entity);
    }

    @Operation(summary = "Iniciar sesión", description = "Autentica un administrador por email y contraseña")
    @ApiResponses({
        @ApiResponse(responseCode = "200", description = "Inicio de sesión exitoso"),
        @ApiResponse(responseCode = "401", description = "Credenciales inválidas", content = @Content)
    })
    @PostMapping("/login")
    public ResponseEntity<AdministradorResponseDTO> login(@RequestBody Administrador request) {
        Administrador admin = service.login(request.getEmail(), request.getContrasenia());
        return ResponseEntity.ok(mapper.toResponseDTO(admin));
    }
}
