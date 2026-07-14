package com.barber.barberBackend.config;

import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

import io.swagger.v3.oas.annotations.OpenAPIDefinition;
import io.swagger.v3.oas.annotations.enums.SecuritySchemeType;
import io.swagger.v3.oas.annotations.info.Contact;
import io.swagger.v3.oas.annotations.info.Info;
import io.swagger.v3.oas.annotations.info.License;
import io.swagger.v3.oas.annotations.security.SecurityScheme;
import io.swagger.v3.oas.annotations.servers.Server;
import io.swagger.v3.oas.models.Paths;
import io.swagger.v3.oas.models.PathItem.HttpMethod;
import io.swagger.v3.oas.models.security.SecurityRequirement;

import org.springdoc.core.customizers.OpenApiCustomizer;

@Configuration
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
@SecurityScheme(
    name = "bearerAuth",
    type = SecuritySchemeType.HTTP,
    scheme = "bearer",
    bearerFormat = "JWT",
    description = "Token JWT obtenido de POST /administradores/login"
)
public class OpenApiConfig {

    @Bean
    public OpenApiCustomizer securityRequirementsCustomizer() {
        return openApi -> {
            SecurityRequirement secReq = new SecurityRequirement().addList("bearerAuth");
            Paths paths = openApi.getPaths();
            if (paths == null) return;

            paths.forEach((path, pathItem) -> {
                pathItem.readOperationsMap().forEach((method, operation) -> {
                    if (!isPublic(path, method)) {
                        operation.addSecurityItem(secReq);
                    }
                });
            });
        };
    }

    private static boolean isPublic(String path, HttpMethod method) {
        if (path.equals("/ping")) return true;
        if (path.equals("/turnos/findDateTimes")) return true;
        if (path.equals("/turnos") && method == HttpMethod.POST) return true;
        if (path.startsWith("/servicios") && method == HttpMethod.GET) return true;
        if (path.equals("/administradores/login") && method == HttpMethod.POST) return true;
        return false;
    }
}
