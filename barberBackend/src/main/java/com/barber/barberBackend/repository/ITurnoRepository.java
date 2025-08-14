package com.barber.barberBackend.repository;

import java.time.LocalDateTime;
import java.util.List;

import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import com.barber.barberBackend.dto.ClienteFrecuenteDTO;
import com.barber.barberBackend.dto.HorarioEstadisticaDTO;
import com.barber.barberBackend.dto.PanelEstadisticaDTO;
import com.barber.barberBackend.dto.ServicioEstadisticaDTO;
import com.barber.barberBackend.generics.GenericRepository;
import com.barber.barberBackend.model.Turno;

@Repository
public interface ITurnoRepository extends GenericRepository<Turno, Long> {
     @Query("SELECT COUNT(t) FROM Turno t WHERE DATE(t.fechaHora) = DATE(:fecha)")
     int countByFecha(@Param("fecha") LocalDateTime fecha);
     @Query("""
          SELECT SUM(s.precio)
          FROM Turno t
          JOIN t.servicio s
          WHERE t.fechaHora >= :inicio
               AND t.fechaHora < :fin
     """)
     Double findIngresosByFecha(
          @Param("inicio") LocalDateTime inicio,
          @Param("fin") LocalDateTime fin
     );
     // cant Servicios por fecha
     @Query("""
          SELECT COUNT(s)
          FROM Turno t
          JOIN t.servicio s
          WHERE t.fechaHora >= :desde
               AND t.fechaHora < :hasta
               AND s.id = :servicioId
     """)
     double countServiciosByFecha(
          @Param("desde") LocalDateTime desde,
          @Param("hasta") LocalDateTime hasta,
          @Param("servicioId") Long servicioId
     );

     @Query("""
          SELECT new com.barber.barberBackend.dto.ServicioEstadisticaDTO(
               s.id,
               s.nombre,
               COUNT(s)
          )
          FROM Turno t
          JOIN t.servicio s
          WHERE t.fechaHora >= :desde
               AND t.fechaHora < :hasta
          GROUP BY s.id, s.nombre
          ORDER BY COUNT(s) DESC
     """)
     List<ServicioEstadisticaDTO> countServiciosByFecha(
          @Param("desde") LocalDateTime desde,
          @Param("hasta") LocalDateTime hasta
     );
     
     @Query("""
          SELECT new com.barber.barberBackend.dto.HorarioEstadisticaDTO(
               t.fechaHora,
               COUNT(t)
          )
          FROM Turno t
          WHERE t.fechaHora >= :desde
          GROUP BY HOUR(t.fechaHora)
          ORDER BY HOUR(t.fechaHora)
     """)
     List<HorarioEstadisticaDTO> countTurnosByHorario(@Param("desde") LocalDateTime desde); 

     @Query("""
          SELECT new com.barber.barberBackend.dto.ClienteFrecuenteDTO(
               c.id,
               c.nombre,
               c.apellido,
               COUNT(t)
     )
          FROM Turno t
          JOIN t.cliente c
          WHERE t.fechaHora >= :desde
          GROUP BY c.id, c.nombre
          ORDER BY COUNT(t) DESC
     """)
     List<ClienteFrecuenteDTO> findClientesFrecuentes(@Param("desde") LocalDateTime desde);

     // Turnos Ocupados
     @Query("SELECT t.fechaHora FROM Turno t WHERE t.fechaHora >= :desde")
     List<LocalDateTime> findDateTimes(@Param("desde") LocalDateTime desde);
     
}
