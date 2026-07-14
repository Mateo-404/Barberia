package com.barber.barberBackend.controller;

import com.barber.barberBackend.auth.JwtService;
import com.barber.barberBackend.config.SecurityConfig;
import com.barber.barberBackend.dto.TurnoRequestDTO;
import com.barber.barberBackend.dto.TurnoResponseDTO;
import com.barber.barberBackend.exception.ResourceNotFoundException;
import com.barber.barberBackend.model.Cliente;
import com.barber.barberBackend.model.Servicio;
import com.barber.barberBackend.model.Turno;
import com.barber.barberBackend.service.TurnoMapper;
import com.barber.barberBackend.service.TurnoService;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.autoconfigure.web.servlet.WebMvcTest;
import org.springframework.boot.test.mock.mockito.MockBean;
import org.springframework.context.annotation.Import;
import org.springframework.http.MediaType;
import org.springframework.security.test.context.support.WithMockUser;
import org.springframework.test.context.ActiveProfiles;
import org.springframework.test.web.servlet.MockMvc;

import java.time.LocalDateTime;
import java.time.format.DateTimeFormatter;
import java.util.List;

import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.when;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.get;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.post;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.*;

@WebMvcTest(TurnoController.class)
@Import(SecurityConfig.class)
@ActiveProfiles("test")
class TurnoControllerTest {

    @Autowired
    private MockMvc mockMvc;

    @MockBean
    private TurnoService turnoService;

    @MockBean
    private TurnoMapper turnoMapper;

    @MockBean
    private JwtService jwtService;

    private final LocalDateTime futureDate = LocalDateTime.now().plusDays(1).withHour(15).withMinute(0).withSecond(0).withNano(0);

    @Test
    void create_withValidData_returns201() throws Exception {
        TurnoRequestDTO request = new TurnoRequestDTO(futureDate, 1L, "123456789", "Juan", "Pérez", "juan@email.com");

        Cliente cliente = new Cliente();
        cliente.setTelefono("123456789");
        cliente.setNombre("Juan");
        cliente.setApellido("Pérez");
        cliente.setEmail("juan@email.com");

        Servicio servicio = new Servicio(1L, "Corte", 500);

        Turno entity = new Turno(null, futureDate, cliente, servicio);
        Turno saved = new Turno(1L, futureDate, cliente, servicio);

        TurnoResponseDTO response = new TurnoResponseDTO(1L, futureDate,
            "123456789", "Juan", "Pérez", "juan@email.com",
            1L, "Corte", 500);

        when(turnoMapper.toEntity(any(TurnoRequestDTO.class))).thenReturn(entity);
        when(turnoService.save(entity)).thenReturn(saved);
        when(turnoMapper.toResponseDTO(saved)).thenReturn(response);

        String fechaStr = futureDate.format(DateTimeFormatter.ISO_LOCAL_DATE_TIME);

        mockMvc.perform(post("/turnos")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content("{\"fechaHora\":\"" + fechaStr + "\",\"idServicio\":1,\"telefonoCliente\":\"123456789\",\"nombreCliente\":\"Juan\",\"apellidoCliente\":\"Perez\",\"emailCliente\":\"juan@email.com\"}"))
                .andExpect(status().isCreated())
                .andExpect(jsonPath("$.id").value(1))
                .andExpect(jsonPath("$.telefonoCliente").value("123456789"))
                .andExpect(jsonPath("$.tipoServicio").value("Corte"));
    }

    @Test
    void create_withBlankTelefonoCliente_returns400() throws Exception {
        String fechaStr = futureDate.format(DateTimeFormatter.ISO_LOCAL_DATE_TIME);
        mockMvc.perform(post("/turnos")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content("{\"fechaHora\":\"" + fechaStr + "\",\"idServicio\":1,\"telefonoCliente\":\"\",\"nombreCliente\":\"Juan\",\"apellidoCliente\":\"Perez\",\"emailCliente\":\"juan@email.com\"}"))
                .andExpect(status().isBadRequest())
                .andExpect(jsonPath("$.title").value("Bad Request"))
                .andExpect(jsonPath("$.status").value(400));
    }

    @Test
    void create_withNullIdServicio_returns400() throws Exception {
        String fechaStr = futureDate.format(DateTimeFormatter.ISO_LOCAL_DATE_TIME);
        mockMvc.perform(post("/turnos")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content("{\"fechaHora\":\"" + fechaStr + "\",\"idServicio\":null,\"telefonoCliente\":\"123456789\",\"nombreCliente\":\"Juan\",\"apellidoCliente\":\"Perez\",\"emailCliente\":\"juan@email.com\"}"))
                .andExpect(status().isBadRequest())
                .andExpect(jsonPath("$.title").value("Bad Request"))
                .andExpect(jsonPath("$.status").value(400));
    }

    @Test
    void create_withPastFechaHora_returns400() throws Exception {
        LocalDateTime past = LocalDateTime.now().minusDays(1);
        String fechaStr = past.format(DateTimeFormatter.ISO_LOCAL_DATE_TIME);
        mockMvc.perform(post("/turnos")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content("{\"fechaHora\":\"" + fechaStr + "\",\"idServicio\":1,\"telefonoCliente\":\"123456789\",\"nombreCliente\":\"Juan\",\"apellidoCliente\":\"Perez\",\"emailCliente\":\"juan@email.com\"}"))
                .andExpect(status().isBadRequest())
                .andExpect(jsonPath("$.title").value("Bad Request"))
                .andExpect(jsonPath("$.status").value(400));
    }

    @Test
    void create_withServicioNotFound_returns404() throws Exception {
        when(turnoMapper.toEntity(any(TurnoRequestDTO.class)))
            .thenThrow(new ResourceNotFoundException("No existe un servicio con el id 999"));

        String fechaStr = futureDate.format(DateTimeFormatter.ISO_LOCAL_DATE_TIME);

        mockMvc.perform(post("/turnos")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content("{\"fechaHora\":\"" + fechaStr + "\",\"idServicio\":999,\"telefonoCliente\":\"123456789\",\"nombreCliente\":\"Juan\",\"apellidoCliente\":\"Perez\",\"emailCliente\":\"juan@email.com\"}"))
                .andExpect(status().isNotFound())
                .andExpect(jsonPath("$.title").value("Recurso no encontrado"))
                .andExpect(jsonPath("$.status").value(404))
                .andExpect(jsonPath("$.detail").value("No existe un servicio con el id 999"));
    }

    @Test
    void getFechasOcupadas_returnsList() throws Exception {
        LocalDateTime now = LocalDateTime.now();
        when(turnoService.findDateTimes()).thenReturn(List.of(now));

        mockMvc.perform(get("/turnos/findDateTimes"))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$[0]").value(now.toString()));
    }

    @Test
    void getFechasOcupadas_whenEmpty_returnsEmptyList() throws Exception {
        when(turnoService.findDateTimes()).thenReturn(List.of());

        mockMvc.perform(get("/turnos/findDateTimes"))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$").isEmpty());
    }

    @Test
    @WithMockUser
    void createMultiple_withValidData_returns201() throws Exception {
        String fechaStr = futureDate.format(DateTimeFormatter.ISO_LOCAL_DATE_TIME);

        TurnoRequestDTO request = new TurnoRequestDTO(futureDate, 1L, "123456789", "Juan", "Pérez", "juan@email.com");
        TurnoResponseDTO response = new TurnoResponseDTO(1L, futureDate,
            "123456789", "Juan", "Pérez", "juan@email.com",
            1L, "Corte", 500);

        when(turnoMapper.toResponseDTO(any())).thenReturn(response);

        mockMvc.perform(post("/turnos/all")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content("[{\"fechaHora\":\"" + fechaStr + "\",\"idServicio\":1,\"telefonoCliente\":\"123456789\",\"nombreCliente\":\"Juan\",\"apellidoCliente\":\"Perez\",\"emailCliente\":\"juan@email.com\"}]"))
                .andExpect(status().isCreated());
    }

    @Test
    void createMultiple_withoutAuth_returns401() throws Exception {
        String fechaStr = futureDate.format(DateTimeFormatter.ISO_LOCAL_DATE_TIME);

        mockMvc.perform(post("/turnos/all")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content("[{\"fechaHora\":\"" + fechaStr + "\",\"idServicio\":1,\"telefonoCliente\":\"123456789\",\"nombreCliente\":\"Juan\",\"apellidoCliente\":\"Perez\",\"emailCliente\":\"juan@email.com\"}]"))
                .andExpect(status().isUnauthorized());
    }
}
