package com.barber.barberBackend.service;

import com.barber.barberBackend.model.Administrador;

public interface IAdministradorService {
    Administrador login(String email, String contrasenia);
}
