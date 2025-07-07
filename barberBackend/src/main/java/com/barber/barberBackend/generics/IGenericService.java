package com.barber.barberBackend.generics;

import org.springframework.data.domain.Page;
import java.io.Serializable;
import java.util.List;
import java.util.Optional;

public interface IGenericService<T, ID extends Serializable> {
    // Create
    void save(T entity);
    List<T> saveAll(List<T> entities);
    // Read
    List<T> findAll();
    Page<T> findAll(int page, int size);
    Optional<T> findById(ID id);
    boolean existsById(ID id);
    // Update
    Optional<T> update(ID id, T entity);
    // Delete
    void deleteById(ID id);
}