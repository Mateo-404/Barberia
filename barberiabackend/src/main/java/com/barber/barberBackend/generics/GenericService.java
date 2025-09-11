package com.barber.barberBackend.generics;

import org.springframework.beans.BeanUtils;
import org.springframework.beans.BeanWrapper;
import org.springframework.beans.BeanWrapperImpl;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.transaction.annotation.Transactional;
import org.springframework.data.domain.PageRequest;

import java.beans.PropertyDescriptor;
import java.io.Serializable;
import java.util.HashSet;
import java.util.List;
import java.util.Optional;
import java.util.Set;

@Transactional
public abstract class GenericService<T, ID extends Serializable, R extends GenericRepository<T, ID>> implements IGenericService<T, ID> {
    @Autowired
    private R repository;

    public void save(T entity) {
        repository.save(entity);
    }
    public List<T> saveAll(List<T> entities) {
        return repository.saveAll(entities);
    }

    @Transactional(readOnly = true)
    public List<T> findAll() {
        return repository.findAll();
    }

    @Transactional(readOnly = true)
    public Page<T> findAll(int page, int size) {
        Pageable pageable = PageRequest.of(page, size);
        return repository.findAll(pageable);
    }

    @Transactional(readOnly = true)
    public Optional<T> findById(ID id) {
        return repository.findById(id);
    }

    public Optional<T> update(ID id, T partialEntity) {
        return repository.findById(id)
            .map(existingEntity -> {
                // Copia solo las propiedades no nulas del objeto recibido
                BeanUtils.copyProperties(partialEntity, existingEntity, getNullPropertyNames(partialEntity));
                return repository.save(existingEntity);
            });
    }
    
    // Método auxiliar para obtener las propiedades nulas del objeto
    private String[] getNullPropertyNames(T source) {
        final BeanWrapper src = new BeanWrapperImpl(source);
        PropertyDescriptor[] pds = src.getPropertyDescriptors();
        
        Set<String> nullProperties = new HashSet<>();
        for (PropertyDescriptor pd : pds) {
            Object srcValue = src.getPropertyValue(pd.getName());
            if (srcValue == null) {
                nullProperties.add(pd.getName());
            }
        }
        return nullProperties.toArray(new String[0]);
    }

    public void deleteById(ID id) {
        repository.deleteById(id);
    }

    @Transactional(readOnly = true)
    public boolean existsById(ID id) {
        return repository.existsById(id);
    }
}