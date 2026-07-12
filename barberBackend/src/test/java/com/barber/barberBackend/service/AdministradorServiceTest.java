package com.barber.barberBackend.service;

import com.barber.barberBackend.model.Administrador;
import com.barber.barberBackend.repository.IAdministradorRepository;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.ArgumentMatchers.anyString;
import static org.mockito.Mockito.when;

@ExtendWith(MockitoExtension.class)
class AdministradorServiceTest {

    @Mock
    private IAdministradorRepository repository;

    @InjectMocks
    private AdministradorService service;

    @Test
    void login_whenCredentialsMatch_returnsAdmin() {
        Administrador admin = new Administrador(1L, "pass123");
        admin.setEmail("admin@test.com");
        when(repository.findByEmailAndContrasenia("admin@test.com", "pass123")).thenReturn(admin);

        Administrador result = service.login("admin@test.com", "pass123");

        assertNotNull(result);
        assertEquals("admin@test.com", result.getEmail());
    }

    @Test
    void login_whenEmailIsNull_throwsException() {
        Exception e = assertThrows(RuntimeException.class,
                () -> service.login(null, "pass"));
        assertTrue(e.getMessage().contains("no pueden ser nulos"));
    }

    @Test
    void login_whenPasswordIsNull_throwsException() {
        Exception e = assertThrows(RuntimeException.class,
                () -> service.login("admin@test.com", null));
        assertTrue(e.getMessage().contains("no pueden ser nulos"));
    }

    @Test
    void login_whenNoMatch_throwsException() {
        when(repository.findByEmailAndContrasenia(anyString(), anyString())).thenReturn(null);

        Exception e = assertThrows(RuntimeException.class,
                () -> service.login("wrong@test.com", "wrong"));
        assertTrue(e.getMessage().contains("Credenciales inválidas"));
    }
}
