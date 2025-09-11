package com.barber.barberBackend.generics;

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

    @GetMapping
    public ResponseEntity<List<T>> getAll() {
        return ResponseEntity.ok(service.findAll());
    }

    @GetMapping("/{id}")
    public ResponseEntity<T> getById(@PathVariable ID id) {
        return service.findById(id)
                .map(ResponseEntity::ok)
                .orElseGet(() -> ResponseEntity.notFound().build());
    }

    @GetMapping("{page}/{size}")
    public ResponseEntity<Page<T>> getPage(@PathVariable int page, @PathVariable int size) {
        return ResponseEntity.ok(service.findAll(page, size));
    }


    @GetMapping("/exists/{id}")
    public ResponseEntity<Boolean> existsById(@PathVariable ID id) {
        return ResponseEntity.ok(service.existsById(id));
    }

    @PostMapping
    public ResponseEntity<T> create(@RequestBody T entity) {
        service.save(entity);
        return ResponseEntity.status(HttpStatus.CREATED).build();
    }
    @PostMapping("/all")
    public ResponseEntity<List<T>> createAll(@RequestBody List<T> entities) {
        service.saveAll(entities);
        return ResponseEntity.status(HttpStatus.CREATED).build();
    }

    @PatchMapping("/{id}")
    public ResponseEntity<T> update(@PathVariable ID id, @RequestBody T entity) {
        return service.update(id, entity)
                .map(ResponseEntity::ok)
                .orElseGet(() -> ResponseEntity.notFound().build());
    }

    @DeleteMapping("/{id}")
    public ResponseEntity<Void> delete(@PathVariable ID id) {
        if (service.existsById(id)) {
            service.deleteById(id);
            return ResponseEntity.noContent().build();
        }
        return ResponseEntity.notFound().build();
    }
}
