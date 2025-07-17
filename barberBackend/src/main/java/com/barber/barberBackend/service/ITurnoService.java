package com.barber.barberBackend.service;

import java.time.LocalDateTime;
import java.util.List;

import com.barber.barberBackend.generics.IGenericService;
import com.barber.barberBackend.model.Turno;

public interface ITurnoService extends IGenericService<Turno, Long> {
    List<LocalDateTime> findDateTimes();
    void validateDateTime(LocalDateTime fechaHora);
}
