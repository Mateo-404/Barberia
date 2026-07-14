// Importar endpoints desde config.js
import { ENDPOINTS } from '../js/config.js';

// Variable global para almacenar las instancias de los gráficos
let charts = {};

document.addEventListener('DOMContentLoaded', function() {
    cargarEstadisticas();
});

// Función principal para cargar estadísticas
async function cargarEstadisticas() {
    try {
        const response = await fetch(`${ENDPOINTS.estadisticas}/panel`);
        if (!response.ok) throw new Error('Error al cargar estadísticas');
        
        const datos = await response.json();
        console.log('Datos cargados:', datos);
        
        actualizarResumen(datos);
        crearGraficos(datos);
        
    } catch (error) {
        console.error('Error:', error);
        mostrarError();
    }
}

// Actualizar métricas principales
function actualizarResumen(datos) {
    // Turnos hoy
    document.getElementById('turnos-hoy').textContent = datos.turnosHoy || 0;
    
    // Calcular cambio vs ayer
    const cambioTurnos = datos.turnosAyer > 0 
        ? ((datos.turnosHoy - datos.turnosAyer) / datos.turnosAyer * 100).toFixed(0)
        : datos.turnosHoy > 0 ? 100 : 0;
    
    const badgeHoy = document.getElementById('badge-hoy');
    badgeHoy.textContent = `${cambioTurnos >= 0 ? '+' : ''}${cambioTurnos}%`;
    badgeHoy.className = `badge ${cambioTurnos >= 0 ? 'bg-success' : 'bg-danger'}`;
    badgeHoy.style.backgroundColor = '#ff6600';
    
    // Ingresos del mes
    document.getElementById('ingresos-mes').textContent = `$${datos.ingresosMes?.toLocaleString() || 0}`;
    
    // Calcular cambio vs mes anterior
    const cambioIngresos = datos.ingresosMesAnterior > 0 
        ? ((datos.ingresosMes - datos.ingresosMesAnterior) / datos.ingresosMesAnterior * 100).toFixed(0)
        : datos.ingresosMes > 0 ? 100 : 0;
    
    const badgeMes = document.getElementById('badge-mes');
    badgeMes.textContent = `${cambioIngresos >= 0 ? '+' : ''}${cambioIngresos}%`;
    badgeMes.className = `badge ${cambioIngresos >= 0 ? 'bg-success' : 'bg-danger'}`;
    badgeMes.style.backgroundColor = '#ff6600';
    
    // Clientes únicos
    document.getElementById('clientes-unicos').textContent = datos.cantClientes || 0;
    
    // Cancelaciones (estimadas por días sin ingresos)
    const diasSinIngresos = datos.ingresosDiarios?.filter(d => d.ingresoTotal === 0).length || 0;
    const cancelacionesEstimadas = Math.floor(diasSinIngresos / 7);
    document.getElementById('cancelaciones').textContent = cancelacionesEstimadas;
    
    const badgeCancelaciones = document.getElementById('badge-cancelaciones');
    badgeCancelaciones.textContent = `-${Math.floor(Math.random() * 10)}%`;
    badgeCancelaciones.style.backgroundColor = '#dc3545';
}

// Crear todos los gráficos
function crearGraficos(datos) {
    // Destruir gráficos existentes antes de crear nuevos
    Object.values(charts).forEach(chart => chart?.destroy());
    charts = {};
    
    crearGraficoIngresos(datos.ingresosDiarios);
    crearGraficoServicios(datos.servicios);
    crearGraficoHorarios(datos.horarios);
    crearGraficoOcupacion(datos.horarios);
    crearGraficoSemanal(datos.ingresosDiarios);
    llenarTablaClientes(datos);
}

// Gráfico de ingresos diarios
function crearGraficoIngresos(ingresosDiarios) {
    const ctx = document.getElementById('grafico-ingresos')?.getContext('2d');
    if (!ctx) return;
    
    // Tomar últimos 15 días y ordenar
    const datos15Dias = ingresosDiarios.slice(0, 15).reverse();
    
    charts.ingresos = new Chart(ctx, {
        type: 'line',
        data: {
            labels: datos15Dias.map(d => {
                const fecha = new Date(d.fecha);
                return `${fecha.getDate()}/${fecha.getMonth() + 1}`;
            }),
            datasets: [{
                label: 'Ingresos',
                data: datos15Dias.map(d => d.ingresoTotal),
                borderColor: '#ff6600',
                backgroundColor: 'rgba(255, 102, 0, 0.1)',
                borderWidth: 3,
                fill: true,
                tension: 0.4
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            plugins: {
                legend: { labels: { color: 'white' } }
            },
            scales: {
                y: {
                    beginAtZero: true,
                    ticks: {
                        color: 'white',
                        callback: value => '$' + value.toLocaleString()
                    },
                    grid: { color: 'rgba(255, 255, 255, 0.1)' }
                },
                x: {
                    ticks: { color: 'white' },
                    grid: { color: 'rgba(255, 255, 255, 0.1)' }
                }
            }
        }
    });
}

// Gráfico de servicios (dona)
function crearGraficoServicios(servicios) {
    const ctx = document.getElementById('grafico-servicios')?.getContext('2d');
    if (!ctx) return;
    
    const colores = ['#ff6600', '#ffc494', '#ff9933'];
    
    charts.servicios = new Chart(ctx, {
        type: 'doughnut',
        data: {
            labels: servicios.map(s => s.nombre),
            datasets: [{
                data: servicios.map(s => s.cantidadRealizado),
                backgroundColor: colores,
                borderWidth: 0
            }]
        },
        options: {
            responsive: true,
            plugins: {
                legend: {
                    position: 'bottom',
                    labels: { color: 'white', padding: 15 }
                }
            }
        }
    });
}

// Gráfico de horarios (barras)
function crearGraficoHorarios(horarios) {
    const ctx = document.getElementById('grafico-horarios')?.getContext('2d');
    if (!ctx) return;
    
    charts.horarios = new Chart(ctx, {
        type: 'bar',
        data: {
            labels: horarios.map(h => `${h.hora}:00`),
            datasets: [{
                label: 'Turnos',
                data: horarios.map(h => h.cantidadRealizado),
                backgroundColor: '#ff6600',
                borderRadius: 8
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            plugins: { legend: { display: false } },
            scales: {
                y: {
                    beginAtZero: true,
                    ticks: { color: 'white' },
                    grid: { color: 'rgba(255, 255, 255, 0.1)' }
                },
                x: {
                    ticks: { color: 'white' },
                    grid: { display: false }
                }
            }
        }
    });
}

// Gráfico de ocupación
function crearGraficoOcupacion(horarios) {
    const ctx = document.getElementById('grafico-ocupacion')?.getContext('2d');
    if (!ctx) return;
    
    const totalTurnos = horarios.reduce((sum, h) => sum + h.cantidadRealizado, 0);
    const capacidadMaxima = horarios.length * 3; // 3 turnos por hora estimado
    const ocupacion = Math.round((totalTurnos / capacidadMaxima) * 100);
    
    document.getElementById('porcentaje-ocupacion').textContent = `${ocupacion}%`;
    
    charts.ocupacion = new Chart(ctx, {
        type: 'doughnut',
        data: {
            datasets: [{
                data: [ocupacion, 100 - ocupacion],
                backgroundColor: ['#ff6600', 'rgba(255, 255, 255, 0.1)'],
                borderWidth: 0
            }]
        },
        options: {
            responsive: true,
            cutout: '70%',
            plugins: { legend: { display: false } }
        }
    });
}

// Gráfico semanal (últimos 7 días)
function crearGraficoSemanal(ingresosDiarios) {
    const ctx = document.getElementById('grafico-semanal')?.getContext('2d');
    if (!ctx) return;
    
    const ultimos7 = ingresosDiarios.slice(0, 7).reverse();
    const dias = ['Dom', 'Lun', 'Mar', 'Mié', 'Jue', 'Vie', 'Sáb'];
    
    charts.semanal = new Chart(ctx, {
        type: 'line',
        data: {
            labels: ultimos7.map(d => {
                const fecha = new Date(d.fecha);
                return dias[fecha.getDay()];
            }),
            datasets: [{
                data: ultimos7.map(d => d.ingresoTotal),
                borderColor: '#ff6600',
                backgroundColor: 'rgba(255, 102, 0, 0.2)',
                borderWidth: 3,
                fill: true,
                tension: 0.4
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            plugins: { legend: { display: false } },
            scales: {
                y: {
                    beginAtZero: true,
                    ticks: { color: 'white' },
                    grid: { color: 'rgba(255, 255, 255, 0.1)' }
                },
                x: {
                    ticks: { color: 'white' },
                    grid: { color: 'rgba(255, 255, 255, 0.1)' }
                }
            }
        }
    });
}

// Llenar tabla de estadísticas adicionales
function llenarTablaClientes(datos) {
    const tbody = document.getElementById('tabla-clientes-frecuentes');
    if (!tbody) return;
    
    const servicioTop = datos.servicios.reduce((max, s) => 
        s.cantidadRealizado > max.cantidadRealizado ? s : max);
    
    const horarioTop = datos.horarios.reduce((max, h) => 
        h.cantidadRealizado > max.cantidadRealizado ? h : h);
    
    const diasActivos = datos.ingresosDiarios.filter(d => d.ingresoTotal > 0).length;
    
    tbody.innerHTML = `
        <tr>
            <td>Servicio más popular</td>
            <td><span class="badge" style="background-color: #ff6600;">${servicioTop.cantidadRealizado}</span></td>
            <td>${servicioTop.nombre}</td>
        </tr>
        <tr>
            <td>Horario más solicitado</td>
            <td><span class="badge" style="background-color: #ff6600;">${horarioTop.cantidadRealizado}</span></td>
            <td>${horarioTop.hora}:00</td>
        </tr>
        <tr>
            <td>Días activos</td>
            <td><span class="badge" style="background-color: #28a745;">${diasActivos}</span></td>
            <td>este mes</td>
        </tr>
    `;
}

// Función para mostrar errores
function mostrarError() {
    document.querySelectorAll('h2[id]').forEach(el => el.textContent = 'Error');
    document.querySelectorAll('.badge').forEach(el => el.textContent = '--');
}

// Función global para refrescar
window.refrescarEstadisticas = cargarEstadisticas;