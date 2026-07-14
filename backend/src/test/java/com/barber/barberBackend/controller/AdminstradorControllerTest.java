package com.barber.barberBackend.controller;

import com.barber.barberBackend.dto.AdministradorRequestDTO;
import com.barber.barberBackend.dto.AdministradorResponseDTO;
import com.barber.barberBackend.exception.InvalidCredentialsException;
import com.barber.barberBackend.model.Administrador;
import com.barber.barberBackend.repository.IAdministradorRepository;
import com.barber.barberBackend.service.AdministradorMapper;
import com.barber.barberBackend.service.AdministradorService;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.autoconfigure.web.servlet.WebMvcTest;
import org.springframework.boot.test.mock.mockito.MockBean;
import org.springframework.http.MediaType;
import org.springframework.test.context.ActiveProfiles;
import org.springframework.test.web.servlet.MockMvc;

import static org.mockito.ArgumentMatchers.any;
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

    @MockBean
    private IAdministradorRepository administradorRepository;

    @Test
    void create_withValidData_returns201() throws Exception {
        AdministradorRequestDTO request = new AdministradorRequestDTO("Carlos", "López", "carlos@email.com", "pass1234");
        Administrador entity = new Administrador();
        entity.setNombre("Carlos");
        entity.setApellido("López");
        entity.setEmail("carlos@email.com");
        entity.setContrasenia("pass1234");
        Administrador saved = new Administrador();
        saved.setId(1L);
        saved.setNombre("Carlos");
        saved.setApellido("López");
        saved.setEmail("carlos@email.com");
        saved.setContrasenia("hashedpass1234");
        AdministradorResponseDTO response = new AdministradorResponseDTO(1L, "Carlos", "López", "carlos@email.com");

        when(administradorRepository.existsByEmail("carlos@email.com")).thenReturn(false);
        when(administradorMapper.toEntity(request)).thenReturn(entity);
        when(administradorService.save(entity)).thenReturn(saved);
        when(administradorMapper.toResponseDTO(saved)).thenReturn(response);

        mockMvc.perform(post("/administradores")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content("{\"nombre\":\"Carlos\",\"apellido\":\"López\",\"email\":\"carlos@email.com\",\"contrasenia\":\"pass1234\"}"))
                .andExpect(status().isCreated())
                .andExpect(jsonPath("$.id").value(1))
                .andExpect(jsonPath("$.nombre").value("Carlos"))
                .andExpect(jsonPath("$.apellido").value("López"))
                .andExpect(jsonPath("$.email").value("carlos@email.com"))
                .andExpect(jsonPath("$.contrasenia").doesNotExist());
    }

    @Test
    void create_withBlankNombre_returns400() throws Exception {
        mockMvc.perform(post("/administradores")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content("{\"nombre\":\"\",\"apellido\":\"López\",\"email\":\"carlos@email.com\",\"contrasenia\":\"pass1234\"}"))
                .andExpect(status().isBadRequest())
                .andExpect(jsonPath("$.title").value("Bad Request"))
                .andExpect(jsonPath("$.status").value(400));
    }

    @Test
    void create_withBlankApellido_returns400() throws Exception {
        mockMvc.perform(post("/administradores")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content("{\"nombre\":\"Carlos\",\"apellido\":\"\",\"email\":\"carlos@email.com\",\"contrasenia\":\"pass1234\"}"))
                .andExpect(status().isBadRequest())
                .andExpect(jsonPath("$.title").value("Bad Request"))
                .andExpect(jsonPath("$.status").value(400));
    }

    @Test
    void create_withBlankEmail_returns400() throws Exception {
        mockMvc.perform(post("/administradores")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content("{\"nombre\":\"Carlos\",\"apellido\":\"López\",\"email\":\"\",\"contrasenia\":\"pass1234\"}"))
                .andExpect(status().isBadRequest())
                .andExpect(jsonPath("$.title").value("Bad Request"))
                .andExpect(jsonPath("$.status").value(400));
    }

    @Test
    void create_withInvalidEmail_returns400() throws Exception {
        mockMvc.perform(post("/administradores")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content("{\"nombre\":\"Carlos\",\"apellido\":\"López\",\"email\":\"email-invalido\",\"contrasenia\":\"pass1234\"}"))
                .andExpect(status().isBadRequest())
                .andExpect(jsonPath("$.title").value("Bad Request"))
                .andExpect(jsonPath("$.status").value(400));
    }

    @Test
    void create_withBlankContrasenia_returns400() throws Exception {
        mockMvc.perform(post("/administradores")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content("{\"nombre\":\"Carlos\",\"apellido\":\"López\",\"email\":\"carlos@email.com\",\"contrasenia\":\"\"}"))
                .andExpect(status().isBadRequest())
                .andExpect(jsonPath("$.title").value("Bad Request"))
                .andExpect(jsonPath("$.status").value(400));
    }

    @Test
    void create_withShortContrasenia_returns400() throws Exception {
        mockMvc.perform(post("/administradores")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content("{\"nombre\":\"Carlos\",\"apellido\":\"López\",\"email\":\"carlos@email.com\",\"contrasenia\":\"1234567\"}"))
                .andExpect(status().isBadRequest())
                .andExpect(jsonPath("$.title").value("Bad Request"))
                .andExpect(jsonPath("$.status").value(400));
    }

    @Test
    void create_withDuplicateEmail_returns409() throws Exception {
        when(administradorRepository.existsByEmail("carlos@email.com")).thenReturn(true);

        mockMvc.perform(post("/administradores")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content("{\"nombre\":\"Carlos\",\"apellido\":\"López\",\"email\":\"carlos@email.com\",\"contrasenia\":\"pass1234\"}"))
                .andExpect(status().isConflict())
                .andExpect(jsonPath("$.title").value("El recurso ya existe"))
                .andExpect(jsonPath("$.status").value(409))
                .andExpect(jsonPath("$.detail").value("Ya existe un administrador con el email carlos@email.com"));
    }

    @Test
    void login_withValidCredentials_returnsOk() throws Exception {
        Administrador admin = new Administrador(1L, "pass1234");
        admin.setEmail("admin@test.com");
        when(administradorService.login("admin@test.com", "pass1234")).thenReturn(admin);
        when(administradorMapper.toResponseDTO(admin)).thenReturn(
            new AdministradorResponseDTO(1L, admin.getNombre(), admin.getApellido(), admin.getEmail()));

        mockMvc.perform(post("/administradores/login")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content("{\"email\":\"admin@test.com\",\"contrasenia\":\"pass1234\"}"))
                .andExpect(status().isOk());
    }

    @Test
    void login_withInvalidCredentials_returnsUnauthorized() throws Exception {
        when(administradorService.login(anyString(), anyString()))
                .thenThrow(new InvalidCredentialsException("Credenciales inválidas"));

        mockMvc.perform(post("/administradores/login")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content("{\"email\":\"wrong@test.com\",\"contrasenia\":\"wrong\"}"))
                .andExpect(status().isUnauthorized());
    }
}
