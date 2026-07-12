package com.barber.barberBackend.controller;

import com.barber.barberBackend.service.TurnoService;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.autoconfigure.web.servlet.WebMvcTest;
import org.springframework.boot.test.mock.mockito.MockBean;
import org.springframework.test.context.ActiveProfiles;
import org.springframework.test.web.servlet.MockMvc;

import java.time.LocalDateTime;
import java.util.List;

import static org.mockito.Mockito.when;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.get;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.*;

@WebMvcTest(TurnoController.class)
@ActiveProfiles("test")
class TurnoControllerTest {

    @Autowired
    private MockMvc mockMvc;

    @MockBean
    private TurnoService turnoService;

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
}
