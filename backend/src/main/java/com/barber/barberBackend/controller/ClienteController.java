package com.barber.barberBackend.controller;

import java.util.List;

import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.media.Content;
import io.swagger.v3.oas.annotations.media.Schema;
import io.swagger.v3.oas.annotations.responses.ApiResponse;
import io.swagger.v3.oas.annotations.responses.ApiResponses;
import io.swagger.v3.oas.annotations.tags.Tag;

import jakarta.validation.Valid;

import org.springframework.http.HttpStatus;
import org.springframework.http.ProblemDetail;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import com.barber.barberBackend.dto.ClienteRequestDTO;
import com.barber.barberBackend.dto.ClienteResponseDTO;
import com.barber.barberBackend.exception.ResourceAlreadyExistsException;
import com.barber.barberBackend.generics.GenericController;
import com.barber.barberBackend.model.Cliente;
import com.barber.barberBackend.service.ClienteMapper;
import com.barber.barberBackend.service.ClienteService;

@RestController
@RequestMapping("/clientes")
@Tag(name = "Clientes", description = "Gestión de clientes")
public class ClienteController extends GenericController<Cliente, ClienteResponseDTO, String, ClienteService> {

    private final ClienteService service;
    private final ClienteMapper mapper;

    public ClienteController(ClienteService service, ClienteMapper mapper) {
        this.service = service;
        this.mapper = mapper;
    }

    @Override
    protected ClienteResponseDTO toDTO(Cliente entity) {
        return mapper.toResponseDTO(entity);
    }

    @Operation(summary = "Crear un nuevo cliente")
    @ApiResponses({
        @ApiResponse(responseCode = "201", description = "Cliente creado"),
        @ApiResponse(responseCode = "400", description = "Datos inválidos", content = @Content(schema = @Schema(implementation = ProblemDetail.class))),
        @ApiResponse(responseCode = "409", description = "Ya existe un cliente con ese teléfono", content = @Content(schema = @Schema(implementation = ProblemDetail.class)))
    })
    @PostMapping
    public ResponseEntity<ClienteResponseDTO> create(@RequestBody @Valid ClienteRequestDTO request) {
        if (service.existsById(request.telefono())) {
            throw new ResourceAlreadyExistsException(
                "Ya existe un cliente con el teléfono " + request.telefono());
        }
        Cliente entity = mapper.toEntity(request);
        Cliente saved = service.save(entity);
        return ResponseEntity.status(HttpStatus.CREATED).body(mapper.toResponseDTO(saved));
    }

    @Operation(summary = "Crear múltiples clientes")
    @ApiResponses({
        @ApiResponse(responseCode = "201", description = "Clientes creados"),
        @ApiResponse(responseCode = "400", description = "Datos inválidos", content = @Content(schema = @Schema(implementation = ProblemDetail.class))),
        @ApiResponse(responseCode = "409", description = "Ya existe un cliente con ese teléfono", content = @Content(schema = @Schema(implementation = ProblemDetail.class)))
    })
    @PostMapping("/all")
    public ResponseEntity<List<ClienteResponseDTO>> createMultiple(@RequestBody @Valid List<ClienteRequestDTO> requests) {
        for (ClienteRequestDTO request : requests) {
            if (service.existsById(request.telefono())) {
                throw new ResourceAlreadyExistsException(
                    "Ya existe un cliente con el teléfono " + request.telefono());
            }
        }
        List<Cliente> entities = requests.stream().map(mapper::toEntity).toList();
        List<Cliente> saved = service.saveAll(entities);
        return ResponseEntity.status(HttpStatus.CREATED).body(
            saved.stream().map(mapper::toResponseDTO).toList()
        );
    }
}
