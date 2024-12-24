package com.barber.barberBackend.service;

import java.util.List;

import com.barber.barberBackend.model.Servicio;

public interface IServicioService {
	
	public List<Servicio> buscarTodos();
	
	public Servicio buscarId(long id);

    public Servicio buscarNombre(String nombre);
	
	public Servicio guardar(Servicio servicio);
	
	public void borrarPorId(long id);
	
}

