package com.barber.barberBackend.config;

import io.swagger.v3.oas.annotations.OpenAPIDefinition;
import io.swagger.v3.oas.annotations.info.Contact;
import io.swagger.v3.oas.annotations.info.Info;
import io.swagger.v3.oas.annotations.info.License;
import io.swagger.v3.oas.annotations.servers.Server;

@OpenAPIDefinition(
    info = @Info(
        title = "Barbería API",
        version = "1.0.0",
        description = "API REST para la gestión de turnos, clientes y servicios de una barbería.",
        contact = @Contact(
            name = "Desarrollador",
            email = "dev@barberia.com"
        ),
        license = @License(
            name = "MIT"
        )
    ),
    servers = {
        @Server(url = "http://localhost:8080", description = "Entorno de desarrollo local")
    }
)
public class OpenApiConfig {
}
