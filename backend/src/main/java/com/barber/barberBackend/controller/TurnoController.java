package com.barber.barberBackend.controller;

import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.media.Content;
import io.swagger.v3.oas.annotations.media.Schema;
import io.swagger.v3.oas.annotations.responses.ApiResponse;
import io.swagger.v3.oas.annotations.responses.ApiResponses;
import io.swagger.v3.oas.annotations.tags.Tag;

import jakarta.validation.Valid;

import java.time.LocalDateTime;
import java.util.List;
import java.util.stream.Collectors;

import org.springframework.http.HttpStatus;
import org.springframework.http.ProblemDetail;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import com.barber.barberBackend.dto.TurnoRequestDTO;
import com.barber.barberBackend.dto.TurnoResponseDTO;
import com.barber.barberBackend.generics.GenericController;
import com.barber.barberBackend.model.Turno;
import com.barber.barberBackend.service.TurnoMapper;
import com.barber.barberBackend.service.TurnoService;

@RestController
@RequestMapping("/turnos")
@Tag(name = "Turnos", description = "Gestión de turnos")
public class TurnoController extends GenericController<Turno, TurnoResponseDTO, Long, TurnoService> {

    private final TurnoService service;
    private final TurnoMapper mapper;

    public TurnoController(TurnoService service, TurnoMapper mapper) {
        this.service = service;
        this.mapper = mapper;
    }

    @Override
    protected TurnoResponseDTO toDTO(Turno entity) {
        return mapper.toResponseDTO(entity);
    }

    @Operation(summary = "Crear un nuevo turno")
    @ApiResponses({
        @ApiResponse(responseCode = "201", description = "Turno creado"),
        @ApiResponse(responseCode = "400", description = "Datos inválidos", content = @Content(schema = @Schema(implementation = ProblemDetail.class))),
        @ApiResponse(responseCode = "404", description = "Servicio no encontrado", content = @Content(schema = @Schema(implementation = ProblemDetail.class)))
    })
    @PostMapping
    public ResponseEntity<TurnoResponseDTO> create(@RequestBody @Valid TurnoRequestDTO request) {
        Turno entity = mapper.toEntity(request);
        Turno saved = service.save(entity);
        return ResponseEntity.status(HttpStatus.CREATED).body(mapper.toResponseDTO(saved));
    }

    @Operation(summary = "Crear múltiples turnos")
    @ApiResponses({
        @ApiResponse(responseCode = "201", description = "Turnos creados"),
        @ApiResponse(responseCode = "400", description = "Datos inválidos", content = @Content(schema = @Schema(implementation = ProblemDetail.class))),
        @ApiResponse(responseCode = "404", description = "Servicio no encontrado", content = @Content(schema = @Schema(implementation = ProblemDetail.class)))
    })
    @PostMapping("/all")
    public ResponseEntity<List<TurnoResponseDTO>> createMultiple(@RequestBody @Valid List<TurnoRequestDTO> requests) {
        List<Turno> entities = requests.stream().map(mapper::toEntity).toList();
        List<Turno> saved = service.saveAll(entities);
        return ResponseEntity.status(HttpStatus.CREATED).body(
            saved.stream().map(mapper::toResponseDTO).toList()
        );
    }

    @Operation(summary = "Obtener fechas y horas ocupadas", description = "Devuelve una lista con las fechas y horas que ya tienen turno asignado")
    @ApiResponses({
        @ApiResponse(responseCode = "200", description = "Lista de fechas y horas ocupadas"),
        @ApiResponse(responseCode = "401", description = "No autorizado", content = @Content)
    })
    @GetMapping("/findDateTimes")
    public List<String> getFechasOcupadas() {
        List<LocalDateTime> dateTimes = service.findDateTimes();
        return dateTimes.stream()
            .map(LocalDateTime::toString)
            .collect(Collectors.toList());
    }
}
