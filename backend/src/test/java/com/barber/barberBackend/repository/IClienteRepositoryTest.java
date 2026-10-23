package com.barber.barberBackend.repository;

import com.barber.barberBackend.model.Cliente;
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
class IClienteRepositoryTest {

    @Autowired
    private IClienteRepository repository;

    @Test
    void findByTelefono_whenExists_returnsCliente() {
        Cliente cliente = new Cliente();
        cliente.setTelefono("123456789");
        cliente.setNombre("Juan");
        cliente.setApellido("Perez");
        cliente.setEmail("juan@test.com");
        repository.save(cliente);

        Cliente found = repository.findByTelefono("123456789");

        assertNotNull(found);
        assertEquals("Juan", found.getNombre());
    }

    @Test
    void findByTelefono_whenNotFound_returnsNull() {
        Cliente found = repository.findByTelefono("999999999");

        assertNull(found);
    }
}
