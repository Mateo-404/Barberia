package com.barber.barberBackend.controller;

import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;

import java.time.LocalDateTime;
import java.util.List;
import java.util.stream.Collectors;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import com.barber.barberBackend.generics.GenericController;
import com.barber.barberBackend.model.Turno;
import com.barber.barberBackend.service.TurnoService;
import org.springframework.web.bind.annotation.GetMapping;

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

    @Operation(summary = "Obtener fechas y horas ocupadas", description = "Devuelve una lista con las fechas y horas que ya tienen turno asignado")
    @GetMapping("/findDateTimes")
    public List<String> getFechasOcupadas() {
        List<LocalDateTime> dateTimes = service.findDateTimes();
        return dateTimes.stream()
            .map(LocalDateTime::toString)
            .collect(Collectors.toList());
    }
}
