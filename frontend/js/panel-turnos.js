// dashboard-turnos.js
document.addEventListener('DOMContentLoaded', function() {
    // Datos simulados - reemplazar con datos reales de la API
    let turnosData = [
        {
            id: 1,
            fecha: '2024-08-09',
            hora: '10:00',
            cliente: 'Juan Pérez',
            telefono: '351-123-4567',
            email: 'juan@email.com',
            servicio: 'Corte + Barba',
            estado: 'confirmado',
            precio: 2500
        },
        {
            id: 2,
            fecha: '2024-08-09',
            hora: '11:30',
            cliente: 'Carlos López',
            telefono: '351-987-6543',
            email: 'carlos@email.com',
            servicio: 'Corte',
            estado: 'confirmado',
            precio: 1500
        },
        {
            id: 3,
            fecha: '2024-08-10',
            hora: '09:00',
            cliente: 'Miguel Rodríguez',
            telefono: '351-456-7890',
            email: 'miguel@email.com',
            servicio: 'Barba',
            estado: 'pendiente',
            precio: 1000
        },
        {
            id: 4,
            fecha: '2024-08-10',
            hora: '14:00',
            cliente: 'Diego Martínez',
            telefono: '351-789-0123',
            email: 'diego@email.com',
            servicio: 'Corte + Barba',
            estado: 'confirmado',
            precio: 2500
        },
        {
            id: 5,
            fecha: '2024-08-08',
            hora: '16:00',
            cliente: 'Andrés García',
            telefono: '351-234-5678',
            email: 'andres@email.com',
            servicio: 'Corte',
            estado: 'cancelado',
            precio: 1500
        }
    ];

    const serviciosDisponibles = [
        { id: 'corte', nombre: 'Corte', precio: 1500 },
        { id: 'barba', nombre: 'Barba', precio: 1000 },
        { id: 'combo', nombre: 'Corte + Barba', precio: 2500 }
    ];

    const horariosDisponibles = [
        '09:00', '09:30', '10:00', '10:30', '11:00', '11:30', 
        '12:00', '12:30', '14:00', '14:30', '15:00', '15:30', 
        '16:00', '16:30', '17:00', '17:30', '18:00', '18:30', '19:00'
    ];

    let fechaActualCalendario = new Date();
    let turnoSeleccionado = null;
    let vistaActual = 'lista';

    // Inicialización
    inicializarEventos();
    cargarServicios();
    cargarHorarios();
    mostrarTurnos();
    
    function inicializarEventos() {
        // Eventos de filtros
        document.getElementById('filtro-fecha').addEventListener('change', aplicarFiltros);
        document.getElementById('filtro-estado').addEventListener('change', aplicarFiltros);
        document.getElementById('filtro-servicio').addEventListener('change', aplicarFiltros);
        document.getElementById('buscar-turno').addEventListener('input', aplicarFiltros);
        
        // Eventos de navegación del calendario
        fechaActualCalendario = new Date();
        actualizarMesActual();
    }

    function cargarServicios() {
        const selectServicio = document.getElementById('editarServicio');
        selectServicio.innerHTML = '<option value="">Seleccionar servicio</option>';
        serviciosDisponibles.forEach(servicio => {
            selectServicio.innerHTML += `<option value="${servicio.id}">${servicio.nombre} - $${servicio.precio}</option>`;
        });
    }

    function cargarHorarios() {
        const selectHora = document.getElementById('editarHora');
        selectHora.innerHTML = '<option value="">Seleccionar hora</option>';
        horariosDisponibles.forEach(hora => {
            selectHora.innerHTML += `<option value="${hora}">${hora}</option>`;
        });
    }

    window.cambiarVista = function(vista) {
        vistaActual = vista;
        
        if (vista === 'lista') {
            document.getElementById('vista-lista').className = 'vista-activa';
            document.getElementById('vista-calendario').className = 'vista-oculta';
            document.getElementById('btn-lista').className = 'btn btn-primario active';
            document.getElementById('btn-calendario').className = 'btn btn-secundario';
            mostrarTurnos();
        } else {
            document.getElementById('vista-lista').className = 'vista-oculta';
            document.getElementById('vista-calendario').className = 'vista-activa';
            document.getElementById('btn-lista').className = 'btn btn-secundario';
            document.getElementById('btn-calendario').className = 'btn btn-primario active';
            generarCalendario();
        }
    };

    function mostrarTurnos() {
        const turnosFiltrados = aplicarFiltrosData();
        const tbody = document.getElementById('tabla-turnos');
        tbody.innerHTML = '';
        
        document.getElementById('total-turnos').textContent = `${turnosFiltrados.length} turnos`;
        
        turnosFiltrados.forEach(turno => {
            const row = tbody.insertRow();
            const fechaFormateada = new Date(turno.fecha).toLocaleDateString('es-ES');
            const estadoBadge = getEstadoBadge(turno.estado);
            
            row.innerHTML = `
                <td>${fechaFormateada}</td>
                <td><strong>${turno.hora}</strong></td>
                <td>
                    <div>${turno.cliente}</div>
                    <small class="text-muted">${turno.telefono}</small>
                </td>
                <td>${turno.servicio}</td>
                <td>${estadoBadge}</td>
                <td><strong>$${turno.precio.toLocaleString()}</strong></td>
                <td>
                    <div class="btn-group btn-group-sm">
                        <button class="btn btn-outline-warning btn-sm" onclick="editarTurno(${turno.id})" title="Editar">
                            <i class="bi bi-pencil"></i>
                        </button>
                        <button class="btn btn-outline-danger btn-sm" onclick="cancelarTurno(${turno.id})" title="Cancelar">
                            <i class="bi bi-trash"></i>
                        </button>
                    </div>
                </td>
            `;
        });
    }

    function getEstadoBadge(estado) {
        const badges = {
            'confirmado': '<span class="badge bg-success">Confirmado</span>',
            'pendiente': '<span class="badge bg-warning text-dark">Pendiente</span>',
            'cancelado': '<span class="badge bg-danger">Cancelado</span>'
        };
        return badges[estado] || '<span class="badge bg-secondary">Desconocido</span>';
    }

    function aplicarFiltros() {
        mostrarTurnos();
    }

    function aplicarFiltrosData() {
        let turnosFiltrados = [...turnosData];
        
        // Filtro por fecha
        const filtroFecha = document.getElementById('filtro-fecha').value;
        const hoy = new Date();
        const manana = new Date(hoy);
        manana.setDate(hoy.getDate() + 1);
        
        switch (filtroFecha) {
            case 'hoy':
                turnosFiltrados = turnosFiltrados.filter(t => 
                    new Date(t.fecha).toDateString() === hoy.toDateString()
                );
                break;
            case 'manana':
                turnosFiltrados = turnosFiltrados.filter(t => 
                    new Date(t.fecha).toDateString() === manana.toDateString()
                );
                break;
            case 'semana':
                const inicioSemana = new Date(hoy);
                inicioSemana.setDate(hoy.getDate() - hoy.getDay());
                const finSemana = new Date(inicioSemana);
                finSemana.setDate(inicioSemana.getDate() + 6);
                turnosFiltrados = turnosFiltrados.filter(t => {
                    const fechaTurno = new Date(t.fecha);
                    return fechaTurno >= inicioSemana && fechaTurno <= finSemana;
                });
                break;
            case 'mes':
                turnosFiltrados = turnosFiltrados.filter(t => {
                    const fechaTurno = new Date(t.fecha);
                    return fechaTurno.getMonth() === hoy.getMonth() && 
                           fechaTurno.getFullYear() === hoy.getFullYear();
                });
                break;
        }
        
        // Filtro por estado
        const filtroEstado = document.getElementById('filtro-estado').value;
        if (filtroEstado !== 'todos') {
            turnosFiltrados = turnosFiltrados.filter(t => t.estado === filtroEstado);
        }
        
        // Filtro por servicio
        const filtroServicio = document.getElementById('filtro-servicio').value;
        if (filtroServicio !== 'todos') {
            turnosFiltrados = turnosFiltrados.filter(t => 
                t.servicio.toLowerCase().includes(filtroServicio)
            );
        }
        
        // Filtro por búsqueda
        const busqueda = document.getElementById('buscar-turno').value.toLowerCase();
        if (busqueda) {
            turnosFiltrados = turnosFiltrados.filter(t => 
                t.cliente.toLowerCase().includes(busqueda) ||
                t.telefono.includes(busqueda)
            );
        }
        
        return turnosFiltrados;
    }

    window.limpiarFiltros = function() {
        document.getElementById('filtro-fecha').value = 'todos';
        document.getElementById('filtro-estado').value = 'todos';
        document.getElementById('filtro-servicio').value = 'todos';
        document.getElementById('buscar-turno').value = '';
        mostrarTurnos();
    };

    window.editarTurno = function(id) {
        turnoSeleccionado = turnosData.find(t => t.id === id);
        if (!turnoSeleccionado) return;
        
        // Llenar el modal con los datos del turno
        document.getElementById('editarTurnoId').value = turnoSeleccionado.id;
        document.getElementById('editarFecha').value = turnoSeleccionado.fecha;
        document.getElementById('editarHora').value = turnoSeleccionado.hora;
        document.getElementById('editarNombre').value = turnoSeleccionado.cliente;
        document.getElementById('editarTelefono').value = turnoSeleccionado.telefono;
        
        // Seleccionar el servicio
        const servicioId = serviciosDisponibles.find(s => 
            s.nombre === turnoSeleccionado.servicio
        )?.id || '';
        document.getElementById('editarServicio').value = servicioId;
        document.getElementById('editarEstado').value = turnoSeleccionado.estado;
        
        // Mostrar el modal
        new bootstrap.Modal(document.getElementById('modalEditarTurno')).show();
    };

    window.guardarEdicionTurno = function() {
        const id = parseInt(document.getElementById('editarTurnoId').value);
        const turnoIndex = turnosData.findIndex(t => t.id === id);
        
        if (turnoIndex === -1) return;
        
        // Obtener el servicio seleccionado
        const servicioId = document.getElementById('editarServicio').value;
        const servicio = serviciosDisponibles.find(s => s.id === servicioId);
        
        // Actualizar el turno
        turnosData[turnoIndex] = {
            ...turnosData[turnoIndex],
            fecha: document.getElementById('editarFecha').value,
            hora: document.getElementById('editarHora').value,
            cliente: document.getElementById('editarNombre').value,
            telefono: document.getElementById('editarTelefono').value,
            servicio: servicio.nombre,
            precio: servicio.precio,
            estado: document.getElementById('editarEstado').value
        };
        
        // Cerrar modal y actualizar vista
        bootstrap.Modal.getInstance(document.getElementById('modalEditarTurno')).hide();
        
        if (vistaActual === 'lista') {
            mostrarTurnos();
        } else {
            generarCalendario();
        }
        
        // Mostrar mensaje de éxito
        mostrarMensajeExito('Turno actualizado exitosamente');
    };

    window.cancelarTurno = function(id) {
        const turno = turnosData.find(t => t.id === id);
        if (!turno) return;
        
        document.getElementById('modalConfirmarTexto').textContent = 
            `¿Estás seguro de que deseas cancelar el turno de ${turno.cliente} del ${new Date(turno.fecha).toLocaleDateString('es-ES')} a las ${turno.hora}?`;
        
        document.getElementById('btnConfirmarAccion').onclick = function() {
            const turnoIndex = turnosData.findIndex(t => t.id === id);
            if (turnoIndex !== -1) {
                turnosData[turnoIndex].estado = 'cancelado';
                
                if (vistaActual === 'lista') {
                    mostrarTurnos();
                } else {
                    generarCalendario();
                }
                
                bootstrap.Modal.getInstance(document.getElementById('modalConfirmar')).hide();
                mostrarMensajeExito('Turno cancelado exitosamente');
            }
        };
        
        new bootstrap.Modal(document.getElementById('modalConfirmar')).show();
    };

    // Funciones del calendario
    window.navegarMes = function(direccion) {
        fechaActualCalendario.setMonth(fechaActualCalendario.getMonth() + direccion);
        actualizarMesActual();
        generarCalendario();
    };

    function actualizarMesActual() {
        const meses = [
            'Enero', 'Febrero', 'Marzo', 'Abril', 'Mayo', 'Junio',
            'Julio', 'Agosto', 'Septiembre', 'Octubre', 'Noviembre', 'Diciembre'
        ];
        document.getElementById('mes-actual').textContent = 
            `${meses[fechaActualCalendario.getMonth()]} ${fechaActualCalendario.getFullYear()}`;
    }

    function generarCalendario() {
        const container = document.getElementById('calendario-container');
        const year = fechaActualCalendario.getFullYear();
        const month = fechaActualCalendario.getMonth();
        
        const primerDia = new Date(year, month, 1);
        const ultimoDia = new Date(year, month + 1, 0);
        const diasEnMes = ultimoDia.getDate();
        const primerDiaSemana = primerDia.getDay();
        
        let html = '<div class="calendario-grid" style="display: grid; grid-template-columns: repeat(7, 1fr); gap: 1px; background: #333;">';
        
        // Encabezados de días
        const diasSemana = ['Dom', 'Lun', 'Mar', 'Mié', 'Jue', 'Vie', 'Sáb'];
        diasSemana.forEach(dia => {
            html += `<div class="calendario-header text-center py-2 bg-secondary text-white fw-bold">${dia}</div>`;
        });
        
        // Días vacíos al inicio
        for (let i = 0; i < primerDiaSemana; i++) {
            html += '<div class="calendario-dia bg-dark" style="min-height: 80px;"></div>';
        }
        
        // Días del mes
        for (let dia = 1; dia <= diasEnMes; dia++) {
            const fechaDia = `${year}-${String(month + 1).padStart(2, '0')}-${String(dia).padStart(2, '0')}`;
            const turnosDelDia = turnosData.filter(t => t.fecha === fechaDia);
            const esHoy = new Date().toDateString() === new Date(year, month, dia).toDateString();
            
            html += `
                <div class="calendario-dia bg-dark p-1 ${esHoy ? 'border border-warning' : ''}" 
                     style="min-height: 80px; cursor: pointer; transition: background-color 0.2s;" 
                     onclick="seleccionarDia('${fechaDia}')"
                     onmouseover="this.style.backgroundColor='#333'"
                     onmouseout="this.style.backgroundColor='#1c1c1c'">
                    <div class="text-white fw-bold mb-1">${dia}</div>
                    ${turnosDelDia.map(t => 
                        `<div class="small badge me-1 mb-1" style="background-color: ${getColorEstado(t.estado)}; font-size: 0.6rem;">
                            ${t.hora}
                        </div>`
                    ).join('')}
                </div>
            `;
        }
        
        html += '</div>';
        container.innerHTML = html;
    }

    function getColorEstado(estado) {
        const colores = {
            'confirmado': '#28a745',
            'pendiente': '#ffc107',
            'cancelado': '#dc3545'
        };
        return colores[estado] || '#6c757d';
    }

    window.seleccionarDia = function(fecha) {
        const turnosDelDia = turnosData.filter(t => t.fecha === fecha);
        const fechaFormateada = new Date(fecha).toLocaleDateString('es-ES', {
            weekday: 'long',
            year: 'numeric',
            month: 'long',
            day: 'numeric'
        });
        
        document.getElementById('fecha-seleccionada').textContent = fechaFormateada;
        
        const container = document.getElementById('turnos-dia');
        if (turnosDelDia.length === 0) {
            container.innerHTML = '<p class="text-muted text-center">No hay turnos para este día</p>';
        } else {
            container.innerHTML = turnosDelDia.map(turno => `
                <div class="card bg-secondary mb-2">
                    <div class="card-body p-2">
                        <div class="d-flex justify-content-between align-items-start">
                            <div>
                                <h6 class="card-title mb-1">${turno.hora} - ${turno.cliente}</h6>
                                <small class="text-muted">${turno.servicio}</small>
                            </div>
                            <div>
                                ${getEstadoBadge(turno.estado)}
                            </div>
                        </div>
                        <div class="mt-2">
                            <button class="btn btn-outline-warning btn-sm me-1" onclick="editarTurno(${turno.id})">
                                <i class="bi bi-pencil"></i>
                            </button>
                            <button class="btn btn-outline-danger btn-sm" onclick="cancelarTurno(${turno.id})">
                                <i class="bi bi-trash"></i>
                            </button>
                        </div>
                    </div>
                </div>
            `).join('');
        }
    };

    function mostrarMensajeExito(mensaje) {
        // Crear elemento de alerta temporal
        const alerta = document.createElement('div');
        alerta.className = 'alert alert-success position-fixed top-0 end-0 m-3';
        alerta.style.zIndex = '9999';
        alerta.textContent = mensaje;
        
        document.body.appendChild(alerta);
        
        setTimeout(() => {
            alerta.remove();
        }, 3000);
    }

    // Función para actualizar datos desde la API
    window.actualizarTurnos = function(nuevosTurnos) {
        turnosData = nuevosTurnos;
        if (vistaActual === 'lista') {
            mostrarTurnos();
        } else {
            generarCalendario();
        }
    };

    // Estilos CSS adicionales
    const style = document.createElement('style');
    style.textContent = `
        .vista-activa { display: block !important; }
        .vista-oculta { display: none !important; }
        .calendario-grid .calendario-dia:hover {
            background-color: #333 !important;
        }
    `;
    document.head.appendChild(style);
});