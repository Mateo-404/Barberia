package com.barber.barberBackend.generics;

import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.media.Content;
import io.swagger.v3.oas.annotations.responses.ApiResponse;
import io.swagger.v3.oas.annotations.responses.ApiResponses;

import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;
import java.io.Serializable;
import java.util.List;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.data.domain.Page;

public abstract class GenericController<T, ID extends Serializable, S extends IGenericService<T, ID>> {
    @Autowired
    private S service;

    @Operation(summary = "Obtener todos los registros")
    @ApiResponse(responseCode = "200", description = "Lista de registros")
    @GetMapping
    public ResponseEntity<List<T>> getAll() {
        return ResponseEntity.ok(service.findAll());
    }

    @Operation(summary = "Obtener un registro por ID")
    @ApiResponses({
        @ApiResponse(responseCode = "200", description = "Registro encontrado"),
        @ApiResponse(responseCode = "404", description = "Registro no encontrado", content = @Content)
    })
    @GetMapping("/{id}")
    public ResponseEntity<T> getById(@PathVariable ID id) {
        return service.findById(id)
                .map(ResponseEntity::ok)
                .orElseGet(() -> ResponseEntity.notFound().build());
    }

    @Operation(summary = "Obtener registros paginados")
    @ApiResponse(responseCode = "200", description = "Página de registros")
    @GetMapping("{page}/{size}")
    public ResponseEntity<Page<T>> getPage(@PathVariable int page, @PathVariable int size) {
        return ResponseEntity.ok(service.findAll(page, size));
    }

    @Operation(summary = "Verificar si un registro existe por ID")
    @ApiResponse(responseCode = "200", description = "true si existe, false si no")
    @GetMapping("/exists/{id}")
    public ResponseEntity<Boolean> existsById(@PathVariable ID id) {
        return ResponseEntity.ok(service.existsById(id));
    }

    @Operation(summary = "Crear un nuevo registro")
    @ApiResponses({
        @ApiResponse(responseCode = "201", description = "Registro creado"),
        @ApiResponse(responseCode = "400", description = "Datos inválidos", content = @Content)
    })
    @PostMapping
    public ResponseEntity<T> create(@RequestBody T entity) {
        service.save(entity);
        return ResponseEntity.status(HttpStatus.CREATED).build();
    }

    @Operation(summary = "Crear múltiples registros")
    @ApiResponses({
        @ApiResponse(responseCode = "201", description = "Registros creados"),
        @ApiResponse(responseCode = "400", description = "Datos inválidos", content = @Content)
    })
    @PostMapping("/all")
    public ResponseEntity<List<T>> createAll(@RequestBody List<T> entities) {
        service.saveAll(entities);
        return ResponseEntity.status(HttpStatus.CREATED).build();
    }

    @Operation(summary = "Actualizar parcialmente un registro por ID")
    @ApiResponses({
        @ApiResponse(responseCode = "200", description = "Registro actualizado"),
        @ApiResponse(responseCode = "404", description = "Registro no encontrado", content = @Content)
    })
    @PatchMapping("/{id}")
    public ResponseEntity<T> update(@PathVariable ID id, @RequestBody T entity) {
        return service.update(id, entity)
                .map(ResponseEntity::ok)
                .orElseGet(() -> ResponseEntity.notFound().build());
    }

    @Operation(summary = "Eliminar un registro por ID")
    @ApiResponses({
        @ApiResponse(responseCode = "204", description = "Registro eliminado"),
        @ApiResponse(responseCode = "404", description = "Registro no encontrado", content = @Content)
    })
    @DeleteMapping("/{id}")
    public ResponseEntity<Void> delete(@PathVariable ID id) {
        if (service.existsById(id)) {
            service.deleteById(id);
            return ResponseEntity.noContent().build();
        }
        return ResponseEntity.notFound().build();
    }
}
