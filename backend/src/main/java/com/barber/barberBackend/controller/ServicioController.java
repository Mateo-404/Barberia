package com.barber.barberBackend.controller;

import java.util.List;

import io.swagger.v3.oas.annotations.tags.Tag;

import jakarta.validation.Valid;

import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import com.barber.barberBackend.dto.ServicioRequestDTO;
import com.barber.barberBackend.dto.ServicioResponseDTO;
import com.barber.barberBackend.generics.GenericController;
import com.barber.barberBackend.model.Servicio;
import com.barber.barberBackend.service.ServicioMapper;
import com.barber.barberBackend.service.ServicioService;

@RestController
@RequestMapping("/servicios")
@Tag(name = "Servicios", description = "Gestión de servicios ofrecidos")
public class ServicioController extends GenericController<Servicio, ServicioResponseDTO, Long, ServicioService> {

    private final ServicioService service;
    private final ServicioMapper mapper;

    public ServicioController(ServicioService service, ServicioMapper mapper) {
        this.service = service;
        this.mapper = mapper;
    }

    @Override
    protected ServicioResponseDTO toDTO(Servicio entity) {
        return mapper.toResponseDTO(entity);
    }

    @PostMapping
    public ResponseEntity<ServicioResponseDTO> create(@RequestBody @Valid ServicioRequestDTO request) {
        Servicio entity = mapper.toEntity(request);
        Servicio saved = service.save(entity);
        return ResponseEntity.status(HttpStatus.CREATED).body(mapper.toResponseDTO(saved));
    }

    @PostMapping("/all")
    public ResponseEntity<List<ServicioResponseDTO>> createMultiple(@RequestBody @Valid List<ServicioRequestDTO> requests) {
        List<Servicio> entities = requests.stream().map(mapper::toEntity).toList();
        List<Servicio> saved = service.saveAll(entities);
        return ResponseEntity.status(HttpStatus.CREATED).body(
            saved.stream().map(mapper::toResponseDTO).toList()
        );
    }
}
