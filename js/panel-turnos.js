import { ENDPOINTS, HORARIOS_LABORALES } from "./config.js";

// Llamadas a la API
class ApiService {
    constructor() {
        this.defaultHeaders = {
            'Content-Type': 'application/json',
            'Accept': 'application/json'
        };
    }

    async makeRequest(url, options = {}) {
        try {
            const config = {
                headers: this.defaultHeaders,
                ...options
            };

            const response = await fetch(url, config);
            
            // Manejar diferentes tipos de respuesta según el código de estado
            switch (response.status) {
                case 200: // OK - Respuesta con datos
                    return await response.json();
                    
                case 201: // CREATED - Recurso creado (sin datos en el body según tu backend)
                    return { success: true, status: 201, message: 'Resource created' };
                    
                case 204: // NO CONTENT - Operación exitosa sin datos
                    return { success: true, status: 204, message: 'Operation completed successfully' };
                    
                case 404: // NOT FOUND
                    throw new Error(`Resource not found (404)`);
                    
                case 400: // BAD REQUEST
                    let errorMessage = 'Bad request';
                    try {
                        const errorData = await response.json();
                        errorMessage = errorData.message || errorData.error || errorMessage;
                    } catch (e) {
                        // Si no puede parsear el error como JSON, usar mensaje genérico
                    }
                    throw new Error(`Bad request (400): ${errorMessage}`);
                    
                case 500: // INTERNAL SERVER ERROR
                    throw new Error(`Internal server error (500)`);
                    
                default:
                    if (!response.ok) {
                        throw new Error(`HTTP error! status: ${response.status}`);
                    }
                    
                    // Para otros códigos exitosos, intentar parsear JSON
                    const contentType = response.headers.get('content-type');
                    if (contentType && contentType.includes('application/json')) {
                        return await response.json();
                    }
                    
                    return await response.text();
            }
            
        } catch (error) {
            console.error('API Error:', error);
            throw error;
        }
    }

    // GET - Obtener todos los turnos
    async getTurnos() {
        return this.makeRequest(ENDPOINTS.turnos);
    }

    // GET - Obtener turno por ID
    async getTurnoById(id) {
        return this.makeRequest(`${ENDPOINTS.turnos}/${id}`);
    }
    
    // GET con paginación
    async getTurnosPaginated(page = 0, size = 10) {
        return this.makeRequest(`${ENDPOINTS.turnos}?page=${page}&size=${size}`);
    }

    // GET - Obtener fechas ocupadas
    async getOccupiedDates() {
        return this.makeRequest(`${ENDPOINTS.turnos}/findDateTimes`);
    }

    // GET - Obtener Servicios
    async getServicios() {
        return this.makeRequest(ENDPOINTS.servicios);
    }

    // POST - Crear nuevo turno
    async createTurno(turnoData) {
        return this.makeRequest(ENDPOINTS.turnos, {
            method: 'POST',
            body: JSON.stringify(turnoData)
        });
    }

    // PUT - Actualizar turno completo
    async updateTurno(id, turnoData) {
        return this.makeRequest(`${ENDPOINTS.turnos}/${id}`, {
            method: 'PUT',
            body: JSON.stringify(turnoData)
        });
    }

    // PATCH - Actualización parcial
    async patchTurno(id, partialData) {
        return this.makeRequest(`${ENDPOINTS.turnos}/${id}`, {
            method: 'PATCH',
            body: JSON.stringify(partialData)
        });
    }

    // DELETE - Eliminar turno
    async deleteTurno(id) {
        return this.makeRequest(`${ENDPOINTS.turnos}/${id}`, {
            method: 'DELETE'
        });
    }

}

// turno.model.js - Modelo y utilidades para turnos
class TurnoModel {
    constructor(data) {
        this.id = data.id;
        this.fechaHora = data.fechaHora;
        this.cliente = data.cliente || {};
        this.servicio = data.servicio || {};
    }

    // Convertir datos de la API al formato usado en la interfaz
    static fromApiResponse(apiData) {
        return {
            id: apiData.id,
            fecha: this.extractDate(apiData.fechaHora),
            hora: this.extractTime(apiData.fechaHora),
            cliente: this.getClienteName(apiData.cliente),
            telefono: apiData.cliente?.telefono || '',
            email: apiData.cliente?.email || '',
            servicio: apiData.servicio?.tipo || '',
            estado: 'confirmado', // Por defecto, ya que la API no maneja estados
            precio: apiData.servicio?.precio || 0
        };
    }

    // Convertir datos de la interfaz al formato de la API
    static toApiFormat(uiData) {
        return {
            id: uiData.id,
            fechaHora: `${uiData.fecha}T${uiData.hora}:00`,
            cliente: {
                nombre: uiData.nombre || null,
                apellido: uiData.apellido || null,
                email: null,
                telefono: uiData.telefono || null
            },
            servicio: {
                id: parseInt(uiData.servicioId),
                tipo: uiData.servicio,
                precio: parseFloat(uiData.precio || 0)
            }
        };
    }

    static extractDate(fechaHora) {
        return fechaHora ? fechaHora.split('T')[0] : '';
    }

    static extractTime(fechaHora) {
        if (!fechaHora) return '';
        const time = fechaHora.split('T')[1];
        return time ? time.substring(0, 5) : '';
    }

    static getClienteName(cliente) {
        if (!cliente) return 'Cliente sin nombre';
        const nombre = cliente.nombre || '';
        const apellido = cliente.apellido || '';
        return `${nombre} ${apellido}`.trim() || 'Cliente sin nombre';
    }
}


// Manejo mejorado de notificaciones con toast
    class NotificationService {
      static show(message, type = 'success') {
        const toast = document.getElementById('notificationToast');
        const toastMessage = document.getElementById('toastMessage');
        
        if (toast && toastMessage) {
          toastMessage.textContent = message;
          toast.className = `toast ${type === 'success' ? 'bg-success' : 'bg-danger'} text-white`;
          
          const toastBootstrap = new bootstrap.Toast(toast);
          toastBootstrap.show();
        }
      }

      static showError(message) {
        this.show(message, 'error');
      }
    }

    // Reemplazar la clase NotificationService original
    if (typeof window !== 'undefined') {
      window.NotificationService = NotificationService;
    }

// main.js - Lógica principal de la aplicación
class TurnosManager {
    constructor() {
        this.apiService = new ApiService();
        this.turnosData = [];
        
        this.serviciosDisponibles = [
            { id: 1, nombre: 'Corte de pelo', precio: 1500, tipo: 'Corte de pelo' },
            { id: 2, nombre: 'Tintura', precio: 3000, tipo: 'Tintura' },
            { id: 3, nombre: 'Perfilado de barba', precio: 2000, tipo: 'Perfilado de barba' }
        ];
        
        this.horariosDisponibles = HORARIOS_LABORALES;

        this.fechaActualCalendario = new Date();
        this.turnoSeleccionado = null;
        this.vistaActual = 'lista';

        this.init();
    }

    async init() {
        try {
            await this.cargarTurnosDesdeAPI();
            this.initializeEventListeners();
            this.cargarServicios();
            this.cargarHorarios();
            this.mostrarTurnos();
            this.actualizarMesActual();
        } catch (error) {
            console.error('Error al inicializar la aplicación:', error);
            NotificationService.showError('Error al cargar los datos iniciales');
        }
    }

    async cargarTurnosDesdeAPI() {
        try {
            const apiTurnos = await this.apiService.getTurnos();
            this.turnosData = apiTurnos.map(turno => TurnoModel.fromApiResponse(turno));
        } catch (error) {
            console.error('Error al cargar turnos:', error);
            // Usar datos de respaldo o mostrar error
            this.turnosData = [];
            throw error;
        }
    }

    initializeEventListeners() {
        const dateInput = document.getElementById('editarFecha');
        if (dateInput) {
            dateInput.addEventListener('change', (e) => {
                this.actualizarHorariosDisponibles(e.target.value);
            });
        }
        // Filtros
        const elementos = {
            'filtro-fecha': () => this.aplicarFiltros(),
            'filtro-estado': () => this.aplicarFiltros(),
            'filtro-servicio': () => this.aplicarFiltros(),
            'buscar-turno': () => this.aplicarFiltros()
        };

        Object.entries(elementos).forEach(([id, handler]) => {
            const elemento = document.getElementById(id);
            if (elemento) {
                const eventType = id === 'buscar-turno' ? 'input' : 'change';
                elemento.addEventListener(eventType, handler);
            }
        });

        // Exponer funciones globales necesarias
        this.exposeGlobalFunctions();
    }

    exposeGlobalFunctions() {
        window.cambiarVista = (vista) => this.cambiarVista(vista);
        window.limpiarFiltros = () => this.limpiarFiltros();
        window.editarTurno = (id) => this.editarTurno(id);
        window.guardarEdicionTurno = () => this.guardarEdicionTurno();
        window.cancelarTurno = (id) => this.cancelarTurno(id);
        window.navegarMes = (direccion) => this.navegarMes(direccion);
        window.seleccionarDia = (fecha) => this.seleccionarDia(fecha);
    }

    cargarServicios() {
        const selectServicio = document.getElementById('editarServicio');
        if (!selectServicio) return;

        selectServicio.innerHTML = '<option value="">Seleccionar servicio</option>';
        this.serviciosDisponibles.forEach(servicio => {
            selectServicio.innerHTML += `<option value="${servicio.id}">${servicio.nombre} - $${servicio.precio}</option>`;
        });
    }
    //! Revisar
    /*
    cargarHorarios() {
        const selectHora = document.getElementById('editarHora');
        if (!selectHora) return;

        selectHora.innerHTML = '<option value="">Seleccionar hora</option>';
        this.horariosDisponibles.forEach(hora => {
            selectHora.innerHTML += `<option value="${hora}">${hora}</option>`;
        });
    }
    */

    cambiarVista(vista) {
        this.vistaActual = vista;
        
        const elementos = {
            lista: {
                vista: document.getElementById('vista-lista'),
                btn: document.getElementById('btn-lista'),
                claseVista: 'vista-activa',
                claseBtn: 'btn btn-primario active'
            },
            calendario: {
                vista: document.getElementById('vista-calendario'),
                btn: document.getElementById('btn-calendario'),
                claseVista: 'vista-activa',
                claseBtn: 'btn btn-primario active'
            }
        };

        // Resetear todas las vistas
        Object.values(elementos).forEach(el => {
            if (el.vista) el.vista.className = 'vista-oculta';
            if (el.btn) el.btn.className = 'btn btn-secundario';
        });

        // Activar vista seleccionada
        if (elementos[vista]) {
            if (elementos[vista].vista) elementos[vista].vista.className = elementos[vista].claseVista;
            if (elementos[vista].btn) elementos[vista].btn.className = elementos[vista].claseBtn;
        }

        // Cargar contenido según la vista
        if (vista === 'lista') {
            this.mostrarTurnos();
        } else if (vista === 'calendario') {
            this.generarCalendario();
        }
    }

    mostrarTurnos() {
        const turnosFiltrados = this.aplicarFiltrosData();
        const tbody = document.getElementById('tabla-turnos');
        const totalTurnos = document.getElementById('total-turnos');
        const estadisticasTotalTurnos = document.getElementById('stats-total');
        const statsConfirmados = document.getElementById('stats-confirmados');
        const statsPendientes = document.getElementById('stats-pendientes');
        const statsIngresos = document.getElementById('stats-ingresos');


        if (!tbody) return;

        tbody.innerHTML = '';
        
        // Calcular estadísticas
        if (totalTurnos) {
            totalTurnos.textContent = `${turnosFiltrados.length} turnos`;
            estadisticasTotalTurnos.textContent = `${turnosFiltrados.length}`;
        }

        const confirmados = turnosFiltrados.filter(t => t.estado === 'confirmado').length;
        const pendientes = turnosFiltrados.filter(t => t.estado === 'pendiente').length;
        
        //? Chequear - CORREGIDO
        const ingresosHoy = turnosFiltrados.reduce((total, t) => {
        const hoy = new Date();
        hoy.setHours(0, 0, 0, 0);
        
        // Función helper para crear fecha local desde string
        const crearFechaLocal = (fechaString) => {
            const fecha = new Date(fechaString);
            // Si la fecha viene en formato ISO, asegurar que se interprete como local
            if (fechaString.includes('T')) {
            return new Date(fecha.getTime() + fecha.getTimezoneOffset() * 60000);
            }
            // Si solo es fecha (YYYY-MM-DD), crear fecha local
            const partes = fechaString.split('-');
            if (partes.length === 3) {
            return new Date(parseInt(partes[0]), parseInt(partes[1]) - 1, parseInt(partes[2]));
            }
            return fecha;
        };

        const fechaTurno = crearFechaLocal(t.fecha);
        fechaTurno.setHours(0, 0, 0, 0);
        
        if (fechaTurno.getTime() === hoy.getTime()) {
            return total + t.precio;
        }
        return total;
        }, 0);
        
        if (statsConfirmados) {
            statsConfirmados.textContent = confirmados;
        }
        if (statsPendientes) {
            statsPendientes.textContent = pendientes;
        }
        if (statsIngresos) {
            statsIngresos.textContent = `$${ingresosHoy.toLocaleString()}`;
        }
        
        turnosFiltrados.forEach(turno => {
            const row = tbody.insertRow();
            //const fechaFormateada = new Date(turno.fecha).toLocaleDateString('es-ES');
            const fechaFormateada = turno.fecha.split("T")[0].split("-").reverse().join("/");
            const estadoBadge = this.getEstadoBadge(turno.estado);
            
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

    getEstadoBadge(estado) {
        const badges = {
            'confirmado': '<span class="badge bg-success">Confirmado</span>',
            'pendiente': '<span class="badge bg-warning text-dark">Pendiente</span>',
            'cancelado': '<span class="badge bg-danger">Cancelado</span>'
        };
        return badges[estado] || '<span class="badge bg-secondary">Desconocido</span>';
    }

    aplicarFiltros() {
        this.mostrarTurnos();
    }

    aplicarFiltrosData() {
        let turnosFiltrados = [...this.turnosData];
        
        // Filtro por fecha
        const filtroFecha = document.getElementById('filtro-fecha')?.value;
        // Filtro por rango de fechas
        const fechaDesde = document.getElementById('filtro-fecha-desde')?.value;
        const fechaHasta = document.getElementById('filtro-fecha-hasta')?.value;
        
        if (fechaDesde || fechaHasta) {
            turnosFiltrados = turnosFiltrados.filter(turno => {
                const fechaTurno = new Date(turno.fecha);
                fechaTurno.setHours(0, 0, 0, 0);
                
                if (fechaDesde && fechaHasta) {
                    const desde = new Date(fechaDesde);
                    const hasta = new Date(fechaHasta);
                    hasta.setHours(23, 59, 59, 999);
                    return fechaTurno >= desde && fechaTurno <= hasta;
                } else if (fechaDesde) {
                    const desde = new Date(fechaDesde);
                    return fechaTurno >= desde;
                } else if (fechaHasta) {
                    const hasta = new Date(fechaHasta);
                    hasta.setHours(23, 59, 59, 999);
                    return fechaTurno <= hasta;
                }
                return true;
            });
        }

        if (filtroFecha) {
        const hoy = new Date();
        hoy.setHours(0, 0, 0, 0);
        const manana = new Date(hoy);
        manana.setDate(hoy.getDate() + 1);

        // Función helper para crear fecha local desde string
        const crearFechaLocal = (fechaString) => {
            const fecha = new Date(fechaString);
            // Si la fecha viene en formato ISO, asegurar que se interprete como local
            if (fechaString.includes('T')) {
            // Si tiene información de tiempo, usar como está
            return new Date(fecha.getTime() + fecha.getTimezoneOffset() * 60000);
            }
            // Si solo es fecha (YYYY-MM-DD), crear fecha local
            const partes = fechaString.split('-');
            if (partes.length === 3) {
            return new Date(parseInt(partes[0]), parseInt(partes[1]) - 1, parseInt(partes[2]));
            }
            return fecha;
        };

        switch (filtroFecha) {
            case 'hoy':
            turnosFiltrados = turnosFiltrados.filter(t => {
                const fechaTurno = crearFechaLocal(t.fecha);
                fechaTurno.setHours(0, 0, 0, 0);
                return fechaTurno.getTime() === hoy.getTime();
            });
            break;
            
            case 'manana':
            turnosFiltrados = turnosFiltrados.filter(t => {
                const fechaTurno = crearFechaLocal(t.fecha);
                fechaTurno.setHours(0, 0, 0, 0);
                return fechaTurno.getTime() === manana.getTime();
            });
            break;
            
            case 'semana':
            const inicioSemana = new Date(hoy);
            inicioSemana.setDate(hoy.getDate() - hoy.getDay());
            const finSemana = new Date(inicioSemana);
            finSemana.setDate(inicioSemana.getDate() + 6);
            finSemana.setHours(23, 59, 59, 999);
            
            turnosFiltrados = turnosFiltrados.filter(t => {
                const fechaTurno = crearFechaLocal(t.fecha);
                return fechaTurno >= inicioSemana && fechaTurno <= finSemana;
            });
            break;
            
            case 'mes':
            turnosFiltrados = turnosFiltrados.filter(t => {
                const fechaTurno = crearFechaLocal(t.fecha);
                return fechaTurno.getMonth() === hoy.getMonth() &&
                    fechaTurno.getFullYear() === hoy.getFullYear();
            });
            break;
        }
        }
        
        // Filtro por estado
        const filtroEstado = document.getElementById('filtro-estado')?.value;
        if (filtroEstado && filtroEstado !== 'todos') {
            turnosFiltrados = turnosFiltrados.filter(t => t.estado === filtroEstado);
        }
        
        // Filtro por servicio
        const filtroServicio = document.getElementById('filtro-servicio')?.value;
        if (filtroServicio && filtroServicio !== 'todos') {
            turnosFiltrados = turnosFiltrados.filter(t => 
                t.servicio.toLowerCase().includes(filtroServicio.toLowerCase())
            );
        }
        
        // Filtro por búsqueda
        const busqueda = document.getElementById('buscar-turno')?.value?.toLowerCase();
        if (busqueda) {
            turnosFiltrados = turnosFiltrados.filter(t => 
                t.cliente.toLowerCase().includes(busqueda) ||
                t.telefono.includes(busqueda)
            );
        }
        
        return turnosFiltrados;
    }

    limpiarFiltros() {
        // Elementos básicos
        const elementosBasicos = {
            'filtro-fecha': 'todos',
            'filtro-estado': 'todos', 
            'filtro-servicio': 'todos',
            'buscar-turno': '',
            'filtro-fecha-desde': '',
            'filtro-fecha-hasta': ''
        };

        // Limpiar cada elemento con su valor por defecto
        Object.entries(elementosBasicos).forEach(([id, valorPorDefecto]) => {
            const elemento = document.getElementById(id);
            if (elemento) {
                elemento.value = valorPorDefecto;
                
                // Disparar evento change para elementos select
                if (elemento.tagName === 'SELECT') {
                    elemento.dispatchEvent(new Event('change'));
                }
            }
        });

        // Recargar los turnos
        this.cargarTurnosDesdeAPI()
            .then(() => {
                this.mostrarTurnos();
                NotificationService.show('Filtros limpiados correctamente');
            })
            .catch(error => {
                console.error('Error al recargar turnos:', error);
                NotificationService.showError('Error al recargar los turnos');
            });
    }

    async editarTurno(id) {
        this.turnoSeleccionado = this.turnosData.find(t => t.id === id);
        if (!this.turnoSeleccionado) return;
        
        // Llenar el modal con los datos del turno
        const campos = {
            'editarTurnoId': this.turnoSeleccionado.id,
            'editarFecha': this.turnoSeleccionado.fecha,
            'editarHora': this.turnoSeleccionado.hora,
            'editarNombre': this.turnoSeleccionado.cliente.split(' ')[0] || '',
            'editarApellido': this.turnoSeleccionado.cliente.split(' ')[1] || '',
            'editarTelefono': this.turnoSeleccionado.telefono,
            'editarEmail': this.turnoSeleccionado.email,
            'editarEstado': this.turnoSeleccionado.estado
        };

        Object.entries(campos).forEach(([id, value]) => {
            const elemento = document.getElementById(id);
            if (elemento) elemento.value = value;
        });

        // Seleccionar el servicio
        const servicioId = this.serviciosDisponibles.find(s => 
            s.tipo === this.turnoSeleccionado.servicio
        )?.id || '';
        const selectServicio = document.getElementById('editarServicio');
        if (selectServicio) selectServicio.value = servicioId;

        // Actualizar horarios disponibles para la fecha del turno
        await this.actualizarHorariosDisponibles(this.turnoSeleccionado.fecha);
        
        // Mostrar el modal
        const modal = document.getElementById('modalEditarTurno');
        if (modal) {
            new bootstrap.Modal(modal).show();
        }
    }

    cargarHorarios() {
        const selectHora = document.getElementById('editarHora');
        if (!selectHora) return;
        const selectFecha = document.getElementById('editarFecha');

        // Solo cargar el placeholder inicial
        selectHora.innerHTML = '<option value="">Seleccione una fecha primero</option>';
        this.actualizarHorariosDisponibles(selectFecha.value);
    }

    async checkAvailability(fecha, hora) {
        try {
            const occupiedDates = await this.apiService.getOccupiedDates();
            const dateTimeToCheck = `${fecha}T${hora}`;
            
            return !occupiedDates.includes(dateTimeToCheck);
        } catch (error) {
            console.error('Error al verificar disponibilidad:', error);
            throw error;
        }
    }

    async validateDateTime(fecha, hora) {
        if (!fecha || !hora) return false;

        const isAvailable = await this.checkAvailability(fecha, hora);
        
        if (!isAvailable) {
            NotificationService.showError('El horario seleccionado no está disponible');
            return false;
        }

        return true;
    }

    
   async actualizarHorariosDisponibles(fecha) {
        if (!fecha) return;
        
        try {
            const occupiedDates = await this.apiService.getOccupiedDates();
            console.log('Fechas ocupadas:', occupiedDates);
            const selectHora = document.getElementById('editarHora');
            selectHora.innerHTML = '<option value="">Seleccionar hora</option>';
            
            // Filtrar horas ocupadas SOLO de la fecha seleccionada
            const ocupados = occupiedDates
                .filter(d => d.startsWith(fecha))
                .map(d => {
                    const fechaCompleta = new Date(d);
                    return fechaCompleta.toTimeString().slice(0, 5);
                });
            
            console.log('Fecha seleccionada:', fecha);
            console.log('Horarios ocupados para esta fecha:', ocupados);
            
            // Mostrar solo horarios disponibles
            let horariosDisponibles = [];
            HORARIOS_LABORALES.forEach(hora => {
                if (!ocupados.includes(hora)) {
                    horariosDisponibles.push(hora);
                    const selected = this.turnoSeleccionado?.hora === hora;
                    selectHora.add(new Option(hora, hora, selected, selected));
                }
            });
            
            console.log('Horarios disponibles agregados al select:', horariosDisponibles);
            console.log('Total opciones en el select:', selectHora.options.length);
            
            // Si no queda ninguna opción disponible
            if (selectHora.options.length === 1) {
                const opt = new Option("No hay horarios disponibles", "");
                opt.disabled = true;
                selectHora.add(opt);
            }
            
        } catch (err) {
            console.error('Error al actualizar horarios:', err);
            NotificationService.showError('Error al cargar horarios disponibles');
        }
    }



    async guardarEdicionTurno() {
        try {
        const id = parseInt(document.getElementById('editarTurnoId')?.value);
        if (!id) return;

        const fecha = document.getElementById('editarFecha')?.value;
        const hora = document.getElementById('editarHora')?.value;

        // Validar que haya fecha y hora seleccionadas
        if (!fecha || !hora) {
            NotificationService.showError('Debe seleccionar fecha y hora');
            return;
        }

        // Formatear correctamente la fecha y hora
        const fechaHora = `${fecha}T${hora}:00`;

        const servicioId = document.getElementById('editarServicio')?.value;
        const servicio = this.serviciosDisponibles.find(s => s.id == servicioId);
        
        if (!servicio) {
            NotificationService.showError('Debe seleccionar un servicio válido');
            return;
        }

        const apiData = {
            id: id,
            fechaHora: fechaHora,
            cliente: {
                nombre: document.getElementById('editarNombre')?.value || null,
                apellido: document.getElementById('editarApellido')?.value || null,
                email: null,
                telefono: document.getElementById('editarTelefono')?.value || null
            },
            servicio: {
                id: parseInt(servicioId),
                tipo: servicio.tipo,
                precio: servicio.precio
            }
        };

            // Actualizar en la API usando PATCH
            await this.apiService.patchTurno(id, apiData);
            
            // Actualizar en los datos locales
            const turnoIndex = this.turnosData.findIndex(t => t.id === id);
            if (turnoIndex !== -1) {
                this.turnosData[turnoIndex] = {
                    id: id,
                    fecha: document.getElementById('editarFecha')?.value,
                    hora: document.getElementById('editarHora')?.value,
                    cliente: `${apiData.cliente.nombre || ''} ${apiData.cliente.apellido || ''}`.trim(),
                    telefono: apiData.cliente.telefono || '',
                    servicio: servicio.tipo,
                    precio: servicio.precio,
                    estado: document.getElementById('editarEstado')?.value
                };
            }
                        
            // Cerrar modal y actualizar vista
            const modal = bootstrap.Modal.getInstance(document.getElementById('modalEditarTurno'));
            if (modal) modal.hide();
            
            // Recargar los datos desde la API para asegurar consistencia
            await this.cargarTurnosDesdeAPI();
            
            if (this.vistaActual === 'lista') {
                this.mostrarTurnos();
            } else {
                this.generarCalendario();
            }
            
            NotificationService.show('Turno actualizado exitosamente');

        } catch (error) {
            console.error('Error al actualizar turno:', error);
            NotificationService.showError('Error al actualizar el turno');
        }
    }

    async cancelarTurno(id) {
        const turno = this.turnosData.find(t => t.id === id);
        if (!turno) return;
        
        const modalTexto = document.getElementById('modalConfirmarTexto');
        if (modalTexto) {
            modalTexto.textContent = 
                `¿Estás seguro de que deseas eliminar el turno de ${turno.cliente} del ${new Date(turno.fecha).toLocaleDateString('es-ES')} a las ${turno.hora}?`;
        }
        
        const btnConfirmar = document.getElementById('btnConfirmarAccion');
        if (btnConfirmar) {
            btnConfirmar.onclick = async () => {
                try {
                    await this.apiService.deleteTurno(id);
                    
                    // Remover de los datos locales
                    this.turnosData = this.turnosData.filter(t => t.id !== id);
                    
                    if (this.vistaActual === 'lista') {
                        this.mostrarTurnos();
                    } else {
                        this.generarCalendario();
                    }
                    
                    const modal = bootstrap.Modal.getInstance(document.getElementById('modalConfirmar'));
                    if (modal) modal.hide();
                    
                    NotificationService.show('Turno eliminado exitosamente');
                } catch (error) {
                    console.error('Error al eliminar turno:', error);
                    NotificationService.showError('Error al eliminar el turno');
                }
            };
        }
        
        const modal = document.getElementById('modalConfirmar');
        if (modal) {
            new bootstrap.Modal(modal).show();
        }
    }

    // Funciones del calendario
    navegarMes(direccion) {
        this.fechaActualCalendario.setMonth(this.fechaActualCalendario.getMonth() + direccion);
        this.actualizarMesActual();
        this.generarCalendario();
    }

    actualizarMesActual() {
        const meses = [
            'Enero', 'Febrero', 'Marzo', 'Abril', 'Mayo', 'Junio',
            'Julio', 'Agosto', 'Septiembre', 'Octubre', 'Noviembre', 'Diciembre'
        ];
        const mesActual = document.getElementById('mes-actual');
        if (mesActual) {
            mesActual.textContent = 
                `${meses[this.fechaActualCalendario.getMonth()]} ${this.fechaActualCalendario.getFullYear()}`;
        }
    }

    generarCalendario() {
        const container = document.getElementById('calendario-container');
        if (!container) return;

        const year = this.fechaActualCalendario.getFullYear();
        const month = this.fechaActualCalendario.getMonth();
        
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
            const turnosDelDia = this.turnosData.filter(t => t.fecha === fechaDia);
            const esHoy = new Date().toDateString() === new Date(year, month, dia).toDateString();
            
            html += `
                <div class="calendario-dia bg-dark p-1 ${esHoy ? 'border border-warning' : ''}" 
                     style="min-height: 80px; cursor: pointer; transition: background-color 0.2s;" 
                     onclick="seleccionarDia('${fechaDia}')"
                     onmouseover="this.style.backgroundColor='#333'"
                     onmouseout="this.style.backgroundColor='#1c1c1c'">
                    <div class="text-white fw-bold mb-1">${dia}</div>
                    ${turnosDelDia.map(t => 
                        `<div class="small badge me-1 mb-1" style="background-color: ${this.getColorEstado(t.estado)}; font-size: 0.6rem;">
                            ${t.hora}
                        </div>`
                    ).join('')}
                </div>
            `;
        }
        
        html += '</div>';
        container.innerHTML = html;
    }

    getColorEstado(estado) {
        const colores = {
            'confirmado': '#28a745',
            'pendiente': '#ffc107',
            'cancelado': '#dc3545'
        };
        return colores[estado] || '#6c757d';
    }

    seleccionarDia(fecha) {
        const turnosDelDia = this.turnosData.filter(t => t.fecha === fecha);
        const fechaFormateada = new Date(fecha).toLocaleDateString('es-ES', {
            weekday: 'long',
            year: 'numeric',
            month: 'long',
            day: 'numeric'
        });
        
        const fechaSeleccionada = document.getElementById('fecha-seleccionada');
        if (fechaSeleccionada) {
            fechaSeleccionada.textContent = fechaFormateada;
        }
        
        const container = document.getElementById('turnos-dia');
        if (!container) return;

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
                                ${this.getEstadoBadge(turno.estado)}
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
    }

    // Método público para refrescar datos
    async refreshData() {
        try {
            await this.cargarTurnosDesdeAPI();
            if (this.vistaActual === 'lista') {
                this.mostrarTurnos();
            } else {
                this.generarCalendario();
            }
            NotificationService.show('Datos actualizados correctamente');
        } catch (error) {
            console.error('Error al refrescar datos:', error);
            NotificationService.showError('Error al actualizar los datos');
        }
    }
}

// Inicialización cuando el DOM esté listo
document.addEventListener('DOMContentLoaded', () => {
    // Verificar que Bootstrap esté disponible
    if (typeof bootstrap === 'undefined') {
        console.error('Bootstrap no está cargado');
        return;
    }

    const fechaInput = document.getElementById('editarFecha');
    fechaInput.min = new Date().toISOString().split('T')[0];

    // Inicializar la aplicación
    window.turnosManager = new TurnosManager();

    // Agregar estilos CSS adicionales
    const style = document.createElement('style');
    style.textContent = `
        .vista-activa { display: block !important; }
        .vista-oculta { display: none !important; }
        .calendario-grid .calendario-dia:hover {
            background-color: #333 !important;
        }
        .loading {
            opacity: 0.6;
            pointer-events: none;
        }
    `;
    document.head.appendChild(style);
    // Exponer función global para refrescar datos
    window.refreshTurnos = () => {
        if (window.turnosManager) {
            window.turnosManager.refreshData();
        }
    };
});

// <-- Script Refactor -->
// Funciones adicionales para la interfaz
    function toggleAdvancedFilters() {
      const advancedFilters = document.getElementById('advanced-filters');
      if (advancedFilters) {
        advancedFilters.classList.toggle('d-none');
      }
    }

    function applyAdvancedFilters() {
      if (!window.turnosManager) return;
      
      const criteria = {
        precioMinimo: document.getElementById('filtro-precio-min')?.value,
        precioMaximo: document.getElementById('filtro-precio-max')?.value,
        cliente: document.getElementById('filtro-cliente')?.value,
        // Agregar otros criterios según sea necesario
      };

      const results = window.turnosManager.advancedSearch(criteria);
      // Mostrar resultados filtrados
      window.turnosManager.turnosData = results;
      window.turnosManager.mostrarTurnos();
    }

    function updatePrecio() {
      const servicioSelect = document.getElementById('editarServicio');
      const precioInput = document.getElementById('editarPrecio');
      const previewElement = document.getElementById('preview-turno');
      
      if (!servicioSelect || !precioInput || !window.turnosManager) return;

      const servicioId = servicioSelect.value;
      const servicio = window.turnosManager.serviciosDisponibles.find(s => s.id == servicioId);
      
      if (servicio) {
        precioInput.value = servicio.precio;
        updatePreview();
      } else {
        precioInput.value = '';
      }
    }

    function updatePreview() {
      const previewElement = document.getElementById('preview-turno');
      if (!previewElement) return;

      const fecha = document.getElementById('editarFecha')?.value;
      const hora = document.getElementById('editarHora')?.value;
      const nombre = document.getElementById('editarNombre')?.value;
      const apellido = document.getElementById('editarApellido')?.value;
      const servicio = document.getElementById('editarServicio')?.selectedOptions[0]?.text;
      const precio = document.getElementById('editarPrecio')?.value;

        if (fecha && hora && nombre && servicio) {
        // Corregir el problema de la fecha
        const [year, month, day] = fecha.split("-");
        const fechaFormateada = new Date(
            year,
            month - 1,
            parseInt(day)
        ).toLocaleDateString("es-ES", {
            day: "2-digit",
            month: "2-digit",
            year: "numeric",
        });

        previewElement.innerHTML = `
                <strong>${nombre} ${apellido}</strong><br>
                ${fechaFormateada} a las ${hora}<br>
                <em>${servicio}</em><br>
            `;
        } else {
        previewElement.textContent =
            "Complete los datos para ver la vista previa";
        }
    }

    function goToToday() {
      if (!window.turnosManager) return;
      
      window.turnosManager.fechaActualCalendario = new Date();
      window.turnosManager.actualizarMesActual();
      window.turnosManager.generarCalendario();
      
      // Seleccionar el día de hoy automáticamente
      const today = new Date();
      const todayString = today.toISOString().split('T')[0];
      window.seleccionarDia(todayString);
    }

    function sortTable(column) {
      if (!window.turnosManager) return;
      
      // Implementar ordenamiento
      const isAsc = window.turnosManager.sortDirection !== 'asc';
      window.turnosManager.sortDirection = isAsc ? 'asc' : 'desc';
      window.turnosManager.sortColumn = column;
      
      window.turnosManager.turnosData.sort((a, b) => {
        let aValue = a[column];
        let bValue = b[column];
        
        if (column === 'precio') {
          aValue = parseFloat(aValue) || 0;
          bValue = parseFloat(bValue) || 0;
        } else if (column === 'fecha') {
          aValue = new Date(aValue);
          bValue = new Date(bValue);
        }
        
        if (aValue < bValue) return isAsc ? -1 : 1;
        if (aValue > bValue) return isAsc ? 1 : -1;
        return 0;
      });
      
      window.turnosManager.mostrarTurnos();
    }

    function limpiarBusquedaAvanzada() {
      const form = document.getElementById('formBusquedaAvanzada');
      if (form) {
        form.reset();
      }
    }

    function ejecutarBusquedaAvanzada() {
      if (!window.turnosManager) return;
      
      const criteria = {
        fechaDesde: document.getElementById('searchFechaDesde')?.value,
        fechaHasta: document.getElementById('searchFechaHasta')?.value,
        cliente: document.getElementById('searchCliente')?.value,
        servicio: document.getElementById('searchServicio')?.value,
        precioMinimo: document.getElementById('searchPrecioMin')?.value,
        precioMaximo: document.getElementById('searchPrecioMax')?.value,
        estado: document.getElementById('searchEstado')?.value
      };

      const results = window.turnosManager.advancedSearch(criteria);
      
      // Crear una vista temporal con los resultados
      const originalData = [...window.turnosManager.turnosData];
      window.turnosManager.turnosData = results;
      window.turnosManager.mostrarTurnos();
      
      // Cerrar modal
      const modal = bootstrap.Modal.getInstance(document.getElementById('modalBusquedaAvanzada'));
      if (modal) modal.hide();
      
      // Mostrar botón para restaurar vista
      showRestoreButton(originalData);
    }

    function showRestoreButton(originalData) {
      const existingButton = document.getElementById('restore-view-btn');
      if (existingButton) existingButton.remove();
      
      const button = document.createElement('button');
      button.id = 'restore-view-btn';
      button.className = 'btn btn-outline-warning btn-sm';
      button.innerHTML = '<i class="bi bi-arrow-clockwise me-1"></i>Mostrar todos los turnos';
      button.onclick = () => {
        if (window.turnosManager) {
          window.turnosManager.turnosData = originalData;
          window.turnosManager.mostrarTurnos();
          button.remove();
        }
      };
      
      const totalTurnos = document.getElementById('total-turnos');
      if (totalTurnos && totalTurnos.parentNode) {
        totalTurnos.parentNode.insertBefore(button, totalTurnos.nextSibling);
      }
    }

    // Event listeners para actualización en tiempo real del preview
    document.addEventListener('DOMContentLoaded', function() {
      // Agregar listeners para la vista previa
      const fieldsToWatch = ['editarFecha', 'editarHora', 'editarNombre', 'editarServicio'];
      fieldsToWatch.forEach(fieldId => {
        const field = document.getElementById(fieldId);
        if (field) {
          field.addEventListener('change', updatePreview);
          field.addEventListener('input', updatePreview);
        }
      });
    });


    // Reemplazar la clase NotificationService original
    if (typeof window !== 'undefined') {
      window.NotificationService = NotificationService;
    }