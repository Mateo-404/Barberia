package com.barber.barberBackend.controller;

import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.media.Content;
import io.swagger.v3.oas.annotations.responses.ApiResponse;
import io.swagger.v3.oas.annotations.responses.ApiResponses;
import io.swagger.v3.oas.annotations.tags.Tag;

import java.util.List;

import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.media.Content;
import io.swagger.v3.oas.annotations.responses.ApiResponse;
import io.swagger.v3.oas.annotations.responses.ApiResponses;

import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import com.barber.barberBackend.dto.AdministradorResponseDTO;
import com.barber.barberBackend.generics.GenericController;
import com.barber.barberBackend.model.Administrador;
import com.barber.barberBackend.service.AdministradorMapper;
import com.barber.barberBackend.service.AdministradorService;

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

    @Operation(summary = "Crear un nuevo administrador")
    @ApiResponses({
        @ApiResponse(responseCode = "201", description = "Administrador creado"),
        @ApiResponse(responseCode = "400", description = "Datos inválidos", content = @Content)
    })
    @PostMapping
    public ResponseEntity<AdministradorResponseDTO> create(@RequestBody Administrador entity) {
        Administrador saved = service.save(entity);
        return ResponseEntity.status(HttpStatus.CREATED).body(mapper.toResponseDTO(saved));
    }

    @Operation(summary = "Crear múltiples administradores")
    @ApiResponses({
        @ApiResponse(responseCode = "201", description = "Administradores creados"),
        @ApiResponse(responseCode = "400", description = "Datos inválidos", content = @Content)
    })
    @PostMapping("/all")
    public ResponseEntity<List<AdministradorResponseDTO>> createAll(@RequestBody List<Administrador> entities) {
        List<Administrador> saved = service.saveAll(entities);
        return ResponseEntity.status(HttpStatus.CREATED).body(
            saved.stream().map(mapper::toResponseDTO).toList()
        );
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
