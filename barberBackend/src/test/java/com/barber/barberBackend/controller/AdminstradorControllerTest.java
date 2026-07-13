package com.barber.barberBackend.controller;

import com.barber.barberBackend.dto.AdministradorResponseDTO;
import com.barber.barberBackend.model.Administrador;
import com.barber.barberBackend.service.AdministradorMapper;
import com.barber.barberBackend.service.AdministradorService;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.autoconfigure.web.servlet.WebMvcTest;
import org.springframework.boot.test.mock.mockito.MockBean;
import org.springframework.http.MediaType;
import org.springframework.test.context.ActiveProfiles;
import org.springframework.test.web.servlet.MockMvc;

import static org.mockito.ArgumentMatchers.anyString;
import static org.mockito.Mockito.when;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.post;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.*;

@WebMvcTest(AdminstradorController.class)
@ActiveProfiles("test")
class AdminstradorControllerTest {

    @Autowired
    private MockMvc mockMvc;

    @MockBean
    private AdministradorService administradorService;

    @MockBean
    private AdministradorMapper administradorMapper;

    @Test
    void login_withValidCredentials_returnsOk() throws Exception {
        Administrador admin = new Administrador(1L, "pass123");
        admin.setEmail("admin@test.com");
        when(administradorService.login("admin@test.com", "pass123")).thenReturn(admin);
        when(administradorMapper.toResponseDTO(admin)).thenReturn(
            new AdministradorResponseDTO(1L, admin.getNombre(), admin.getApellido(), admin.getEmail()));

        mockMvc.perform(post("/administradores/login")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content("{\"email\":\"admin@test.com\",\"contrasenia\":\"pass123\"}"))
                .andExpect(status().isOk());
    }

    @Test
    void login_withInvalidCredentials_returnsUnauthorized() throws Exception {
        when(administradorService.login(anyString(), anyString()))
                .thenThrow(new IllegalArgumentException("Credenciales inválidas"));

        mockMvc.perform(post("/administradores/login")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content("{\"email\":\"wrong@test.com\",\"contrasenia\":\"wrong\"}"))
                .andExpect(status().isUnauthorized());
    }
}
