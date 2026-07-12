package com.barber.barberBackend.repository;

import com.barber.barberBackend.model.Administrador;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.autoconfigure.orm.jpa.DataJpaTest;
import org.springframework.test.context.ActiveProfiles;

import static org.junit.jupiter.api.Assertions.*;
import org.springframework.boot.test.autoconfigure.jdbc.AutoConfigureTestDatabase;
import static org.springframework.boot.test.autoconfigure.jdbc.AutoConfigureTestDatabase.Replace;

@DataJpaTest
@AutoConfigureTestDatabase(replace = Replace.NONE)
@ActiveProfiles("test")
class IAdministradorRepositoryTest {

    @Autowired
    private IAdministradorRepository repository;

    @Test
    void findByEmailAndContrasenia_whenExists_returnsAdmin() {
        Administrador admin = new Administrador();
        admin.setNombre("Admin");
        admin.setApellido("Test");
        admin.setEmail("admin@test.com");
        admin.setContrasenia("pass123");
        repository.save(admin);

        Administrador found = repository.findByEmailAndContrasenia("admin@test.com", "pass123");

        assertNotNull(found);
        assertEquals("admin@test.com", found.getEmail());
    }

    @Test
    void findByEmailAndContrasenia_whenNotFound_returnsNull() {
        Administrador found = repository.findByEmailAndContrasenia("noexiste@test.com", "pass");

        assertNull(found);
    }
}
