package com.barber.barberBackend.controller;

import com.barber.barberBackend.dto.ServicioRequestDTO;
import com.barber.barberBackend.dto.ServicioResponseDTO;
import com.barber.barberBackend.model.Servicio;
import com.barber.barberBackend.service.ServicioMapper;
import com.barber.barberBackend.service.ServicioService;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.autoconfigure.web.servlet.WebMvcTest;
import org.springframework.boot.test.mock.mockito.MockBean;
import org.springframework.http.MediaType;
import org.springframework.test.context.ActiveProfiles;
import org.springframework.test.web.servlet.MockMvc;

import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.when;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.post;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.*;

@WebMvcTest(ServicioController.class)
@ActiveProfiles("test")
class ServicioControllerTest {

    @Autowired
    private MockMvc mockMvc;

    @MockBean
    private ServicioService servicioService;

    @MockBean
    private ServicioMapper servicioMapper;

    @Test
    void create_withValidData_returns201() throws Exception {
        ServicioRequestDTO request = new ServicioRequestDTO("Corte", 500);
        Servicio entity = new Servicio(null, "Corte", 500);
        Servicio saved = new Servicio(1L, "Corte", 500);
        ServicioResponseDTO response = new ServicioResponseDTO(1L, "Corte", 500);

        when(servicioMapper.toEntity(request)).thenReturn(entity);
        when(servicioService.save(entity)).thenReturn(saved);
        when(servicioMapper.toResponseDTO(saved)).thenReturn(response);

        mockMvc.perform(post("/servicios")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content("{\"tipo\":\"Corte\",\"precio\":500}"))
                .andExpect(status().isCreated())
                .andExpect(jsonPath("$.id").value(1))
                .andExpect(jsonPath("$.tipo").value("Corte"))
                .andExpect(jsonPath("$.precio").value(500.0));
    }

    @Test
    void create_withBlankTipo_returns400() throws Exception {
        mockMvc.perform(post("/servicios")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content("{\"tipo\":\"\",\"precio\":500}"))
                .andExpect(status().isBadRequest())
                .andExpect(jsonPath("$.title").value("Bad Request"))
                .andExpect(jsonPath("$.status").value(400));
    }

    @Test
    void create_withNegativePrecio_returns400() throws Exception {
        mockMvc.perform(post("/servicios")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content("{\"tipo\":\"Corte\",\"precio\":-1}"))
                .andExpect(status().isBadRequest())
                .andExpect(jsonPath("$.title").value("Bad Request"))
                .andExpect(jsonPath("$.status").value(400));
    }
}
