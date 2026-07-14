package com.barber.barberBackend.controller;

import java.util.List;

import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.media.Content;
import io.swagger.v3.oas.annotations.responses.ApiResponse;
import io.swagger.v3.oas.annotations.responses.ApiResponses;
import io.swagger.v3.oas.annotations.tags.Tag;

import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import com.barber.barberBackend.generics.GenericController;
import com.barber.barberBackend.model.Cliente;
import com.barber.barberBackend.service.ClienteService;

@RestController
@RequestMapping("/clientes")
@Tag(name = "Clientes", description = "Gestión de clientes")
public class ClienteController extends GenericController<Cliente, Cliente, String, ClienteService> {
    private final ClienteService service;

    public ClienteController(ClienteService service) {
        this.service = service;
    }

    @Override
    protected Cliente toDTO(Cliente entity) {
        // FIXME: temporal — DTO = T pass-through until Cliente has its own DTO
        return entity;
    }

    @Operation(summary = "Crear un nuevo cliente")
    @ApiResponses({
        @ApiResponse(responseCode = "201", description = "Cliente creado"),
        @ApiResponse(responseCode = "400", description = "Datos inválidos", content = @Content)
    })
    @PostMapping
    public ResponseEntity<Cliente> create(@RequestBody Cliente entity) {
        Cliente saved = service.save(entity);
        return ResponseEntity.status(HttpStatus.CREATED).body(saved);
    }

    @Operation(summary = "Crear múltiples clientes")
    @ApiResponses({
        @ApiResponse(responseCode = "201", description = "Clientes creados"),
        @ApiResponse(responseCode = "400", description = "Datos inválidos", content = @Content)
    })
    @PostMapping("/all")
    public ResponseEntity<List<Cliente>> createAll(@RequestBody List<Cliente> entities) {
        List<Cliente> saved = service.saveAll(entities);
        return ResponseEntity.status(HttpStatus.CREATED).body(saved);
    }
}
