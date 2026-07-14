package com.barber.barberBackend.controller;

import com.barber.barberBackend.dto.ClienteRequestDTO;
import com.barber.barberBackend.dto.ClienteResponseDTO;
import com.barber.barberBackend.model.Cliente;
import com.barber.barberBackend.service.ClienteMapper;
import com.barber.barberBackend.service.ClienteService;
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

@WebMvcTest(ClienteController.class)
@ActiveProfiles("test")
class ClienteControllerTest {

    @Autowired
    private MockMvc mockMvc;

    @MockBean
    private ClienteService clienteService;

    @MockBean
    private ClienteMapper clienteMapper;

    @Test
    void create_withValidData_returns201() throws Exception {
        ClienteRequestDTO request = new ClienteRequestDTO("123456789", "Juan", "Pérez", "juan@email.com");
        Cliente entity = new Cliente();
        entity.setTelefono("123456789");
        entity.setNombre("Juan");
        entity.setApellido("Pérez");
        entity.setEmail("juan@email.com");
        Cliente saved = new Cliente();
        saved.setTelefono("123456789");
        saved.setNombre("Juan");
        saved.setApellido("Pérez");
        saved.setEmail("juan@email.com");
        ClienteResponseDTO response = new ClienteResponseDTO("123456789", "Juan", "Pérez", "juan@email.com");

        when(clienteService.existsById("123456789")).thenReturn(false);
        when(clienteMapper.toEntity(request)).thenReturn(entity);
        when(clienteService.save(entity)).thenReturn(saved);
        when(clienteMapper.toResponseDTO(saved)).thenReturn(response);

        mockMvc.perform(post("/clientes")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content("{\"telefono\":\"123456789\",\"nombre\":\"Juan\",\"apellido\":\"Pérez\",\"email\":\"juan@email.com\"}"))
                .andExpect(status().isCreated())
                .andExpect(jsonPath("$.telefono").value("123456789"))
                .andExpect(jsonPath("$.nombre").value("Juan"))
                .andExpect(jsonPath("$.apellido").value("Pérez"))
                .andExpect(jsonPath("$.email").value("juan@email.com"));
    }

    @Test
    void create_withBlankTelefono_returns400() throws Exception {
        mockMvc.perform(post("/clientes")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content("{\"telefono\":\"\",\"nombre\":\"Juan\",\"apellido\":\"Pérez\",\"email\":\"juan@email.com\"}"))
                .andExpect(status().isBadRequest())
                .andExpect(jsonPath("$.title").value("Bad Request"))
                .andExpect(jsonPath("$.status").value(400));
    }

    @Test
    void create_withBlankNombre_returns400() throws Exception {
        mockMvc.perform(post("/clientes")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content("{\"telefono\":\"123456789\",\"nombre\":\"\",\"apellido\":\"Pérez\",\"email\":\"juan@email.com\"}"))
                .andExpect(status().isBadRequest())
                .andExpect(jsonPath("$.title").value("Bad Request"))
                .andExpect(jsonPath("$.status").value(400));
    }

    @Test
    void create_withBlankApellido_returns400() throws Exception {
        mockMvc.perform(post("/clientes")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content("{\"telefono\":\"123456789\",\"nombre\":\"Juan\",\"apellido\":\"\",\"email\":\"juan@email.com\"}"))
                .andExpect(status().isBadRequest())
                .andExpect(jsonPath("$.title").value("Bad Request"))
                .andExpect(jsonPath("$.status").value(400));
    }

    @Test
    void create_withInvalidEmail_returns400() throws Exception {
        mockMvc.perform(post("/clientes")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content("{\"telefono\":\"123456789\",\"nombre\":\"Juan\",\"apellido\":\"Pérez\",\"email\":\"email-invalido\"}"))
                .andExpect(status().isBadRequest())
                .andExpect(jsonPath("$.title").value("Bad Request"))
                .andExpect(jsonPath("$.status").value(400));
    }

    @Test
    void create_withDuplicateTelefono_returns409() throws Exception {
        when(clienteService.existsById("123456789")).thenReturn(true);

        mockMvc.perform(post("/clientes")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content("{\"telefono\":\"123456789\",\"nombre\":\"Juan\",\"apellido\":\"Pérez\",\"email\":\"juan@email.com\"}"))
                .andExpect(status().isConflict())
                .andExpect(jsonPath("$.title").value("El recurso ya existe"))
                .andExpect(jsonPath("$.status").value(409))
                .andExpect(jsonPath("$.detail").value("Ya existe un cliente con el teléfono 123456789"));
    }
}
