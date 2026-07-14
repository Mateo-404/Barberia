package com.barber.barberBackend.controller;

import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.media.Content;
import io.swagger.v3.oas.annotations.responses.ApiResponse;
import io.swagger.v3.oas.annotations.responses.ApiResponses;
import io.swagger.v3.oas.annotations.tags.Tag;

import java.time.LocalDateTime;
import java.util.List;
import java.util.stream.Collectors;

import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import com.barber.barberBackend.generics.GenericController;
import com.barber.barberBackend.model.Turno;
import com.barber.barberBackend.service.TurnoService;

@RestController
@RequestMapping("/turnos")
@Tag(name = "Turnos", description = "Gestión de turnos")
public class TurnoController extends GenericController<Turno, Turno, Long, TurnoService> {
    private final TurnoService service;

    public TurnoController(TurnoService service) {
        this.service = service;
    }

    @Override
    protected Turno toDTO(Turno entity) {
        // FIXME: temporal — DTO = T pass-through until Turno has its own DTO
        return entity;
    }

    @Operation(summary = "Crear un nuevo turno")
    @ApiResponses({
        @ApiResponse(responseCode = "201", description = "Turno creado"),
        @ApiResponse(responseCode = "400", description = "Datos inválidos", content = @Content)
    })
    @PostMapping
    public ResponseEntity<Turno> create(@RequestBody Turno entity) {
        Turno saved = service.save(entity);
        return ResponseEntity.status(HttpStatus.CREATED).body(saved);
    }

    @Operation(summary = "Crear múltiples turnos")
    @ApiResponses({
        @ApiResponse(responseCode = "201", description = "Turnos creados"),
        @ApiResponse(responseCode = "400", description = "Datos inválidos", content = @Content)
    })
    @PostMapping("/all")
    public ResponseEntity<List<Turno>> createAll(@RequestBody List<Turno> entities) {
        List<Turno> saved = service.saveAll(entities);
        return ResponseEntity.status(HttpStatus.CREATED).body(saved);
    }

    @Operation(summary = "Obtener fechas y horas ocupadas", description = "Devuelve una lista con las fechas y horas que ya tienen turno asignado")
    @GetMapping("/findDateTimes")
    public List<String> getFechasOcupadas() {
        List<LocalDateTime> dateTimes = service.findDateTimes();
        return dateTimes.stream()
            .map(LocalDateTime::toString)
            .collect(Collectors.toList());
    }
}
