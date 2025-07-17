package com.barber.barberBackend.repository;

import java.time.LocalDateTime;
import java.util.List;

import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import com.barber.barberBackend.generics.GenericRepository;
import com.barber.barberBackend.model.Turno;

@Repository
public interface ITurnoRepository extends GenericRepository<Turno, Long> {
     @Query("SELECT t.fechaHora FROM Turno t WHERE t.fechaHora >= :desde")
     List<LocalDateTime> findDateTimes(@Param("desde") LocalDateTime desde);
     
}
