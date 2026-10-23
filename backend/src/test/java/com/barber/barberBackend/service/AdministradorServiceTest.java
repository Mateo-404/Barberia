package com.barber.barberBackend.service;

import com.barber.barberBackend.model.Administrador;
import com.barber.barberBackend.repository.IAdministradorRepository;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import org.springframework.security.crypto.argon2.Argon2PasswordEncoder;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.ArgumentMatchers.anyString;
import static org.mockito.Mockito.when;

@ExtendWith(MockitoExtension.class)
class AdministradorServiceTest {

    @Mock
    private IAdministradorRepository repository;

    @Mock
    private Argon2PasswordEncoder passwordEncoder;

    @InjectMocks
    private AdministradorService service;

    @Test
    void login_whenCredentialsMatch_returnsAdmin() {
        Administrador admin = new Administrador(1L, "argon2hash");
        admin.setEmail("admin@test.com");
        when(repository.findByEmail("admin@test.com")).thenReturn(admin);
        when(passwordEncoder.matches("pass123", "argon2hash")).thenReturn(true);

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
        when(repository.findByEmail(anyString())).thenReturn(null);

        Exception e = assertThrows(RuntimeException.class,
                () -> service.login("wrong@test.com", "wrong"));
        assertTrue(e.getMessage().contains("Credenciales inválidas"));
    }
}
