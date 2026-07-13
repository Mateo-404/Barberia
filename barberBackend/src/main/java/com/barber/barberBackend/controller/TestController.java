package com.barber.barberBackend.controller;

import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;

import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RestController;

@RestController
@Tag(name = "Health Check", description = "Endpoint de verificación del servicio")
public class TestController {
    @Operation(summary = "Verificar estado del backend", description = "Devuelve un mensaje de confirmación si el servidor está funcionando")
    @GetMapping("/ping")
    public String ping() {
        return "Backend funcionando!";
    }
}
