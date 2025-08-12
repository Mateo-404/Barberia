document.addEventListener('DOMContentLoaded', function() {
    // Datos simulados - reemplazar con datos reales de la API
    const datosSimulados = {
        turnosHoy: 8,
        turnosAyer: 7,
        ingresosMes: 48500,
        ingresosMesAnterior: 44900,
        clientesUnicos: 142,
        clientesMesAnterior: 123,
        cancelaciones: 4,
        cancelacionesAnterior: 6,
        
        ingresosDiarios: [
            { fecha: '2024-07-10', ingresos: 1200 },
            { fecha: '2024-07-11', ingresos: 1500 },
            { fecha: '2024-07-12', ingresos: 1800 },
            { fecha: '2024-07-13', ingresos: 1300 },
            { fecha: '2024-07-14', ingresos: 2100 },
            { fecha: '2024-07-15', ingresos: 1900 },
            { fecha: '2024-07-16', ingresos: 1600 },
            { fecha: '2024-07-17', ingresos: 1700 },
            { fecha: '2024-07-18', ingresos: 2200 },
            { fecha: '2024-07-19', ingresos: 1400 },
            { fecha: '2024-07-20', ingresos: 1800 },
            { fecha: '2024-07-21', ingresos: 2000 },
            { fecha: '2024-07-22', ingresos: 1500 },
            { fecha: '2024-07-23', ingresos: 1900 },
            { fecha: '2024-07-24', ingresos: 2300 },
            { fecha: '2024-07-25', ingresos: 1600 },
            { fecha: '2024-07-26', ingresos: 1800 },
            { fecha: '2024-07-27', ingresos: 2100 },
            { fecha: '2024-07-28', ingresos: 1700 },
            { fecha: '2024-07-29', ingresos: 2000 },
            { fecha: '2024-07-30', ingresos: 1900 },
            { fecha: '2024-08-01', ingresos: 2200 },
            { fecha: '2024-08-02', ingresos: 1800 },
            { fecha: '2024-08-03', ingresos: 1600 },
            { fecha: '2024-08-04', ingresos: 2100 },
            { fecha: '2024-08-05', ingresos: 1900 },
            { fecha: '2024-08-06', ingresos: 2000 },
            { fecha: '2024-08-07', ingresos: 1700 },
            { fecha: '2024-08-08', ingresos: 2300 },
            { fecha: '2024-08-09', ingresos: 1500 }
        ],
        
        servicios: [
            { nombre: 'Corte', cantidad: 45, color: '#ff6600' },
            { nombre: 'Barba', cantidad: 28, color: '#ffc494' },
            { nombre: 'Corte + Barba', cantidad: 35, color: '#ff9933' },
            { nombre: 'Otros', cantidad: 12, color: '#ffb366' }
        ],
        
        horarios: [
            { hora: '09:00', cantidad: 8 },
            { hora: '10:00', cantidad: 12 },
            { hora: '11:00', cantidad: 15 },
            { hora: '12:00', cantidad: 10 },
            { hora: '14:00', cantidad: 14 },
            { hora: '15:00', cantidad: 18 },
            { hora: '16:00', cantidad: 20 },
            { hora: '17:00', cantidad: 16 },
            { hora: '18:00', cantidad: 13 },
            { hora: '19:00', cantidad: 9 }
        ],
        
        clientesFrecuentes: [
            { nombre: 'Juan Pérez', turnos: 12, ultimoServicio: 'Corte + Barba' },
            { nombre: 'Carlos López', turnos: 10, ultimoServicio: 'Corte' },
            { nombre: 'Miguel Rodríguez', turnos: 9, ultimoServicio: 'Barba' },
            { nombre: 'Diego Martínez', turnos: 8, ultimoServicio: 'Corte + Barba' },
            { nombre: 'Andrés García', turnos: 7, ultimoServicio: 'Corte' }
        ]
    };

    // Inicializar dashboard
    actualizarResumen();
    crearGraficoIngresos();
    crearGraficoServicios();
    crearGraficoHorarios();
    crearGraficoOcupacion();
    crearGraficoSemanal();
    llenarTablaClientesFrecuentes();

    function actualizarResumen() {
        // Actualizar métricas principales
        document.getElementById('turnos-hoy').textContent = datosSimulados.turnosHoy;
        document.getElementById('ingresos-mes').textContent = `$${datosSimulados.ingresosMes.toLocaleString()}`;
        document.getElementById('clientes-unicos').textContent = datosSimulados.clientesUnicos;
        document.getElementById('cancelaciones').textContent = datosSimulados.cancelaciones;

        // Calcular y mostrar porcentajes de cambio
        const cambioTurnos = ((datosSimulados.turnosHoy - datosSimulados.turnosAyer) / datosSimulados.turnosAyer * 100).toFixed(0);
        const cambioIngresos = ((datosSimulados.ingresosMes - datosSimulados.ingresosMesAnterior) / datosSimulados.ingresosMesAnterior * 100).toFixed(0);
        const cambioCancelaciones = ((datosSimulados.cancelacionesAnterior - datosSimulados.cancelaciones) / datosSimulados.cancelacionesAnterior * 100).toFixed(0);

        document.getElementById('badge-hoy').textContent = `${cambioTurnos > 0 ? '+' : ''}${cambioTurnos}%`;
        document.getElementById('badge-mes').textContent = `${cambioIngresos > 0 ? '+' : ''}${cambioIngresos}%`;
        document.getElementById('badge-cancelaciones').textContent = `${cambioCancelaciones > 0 ? '+' : ''}${cambioCancelaciones}%`;
    }

    function crearGraficoIngresos() {
        const ctx = document.getElementById('grafico-ingresos').getContext('2d');
        
        new Chart(ctx, {
            type: 'line',
            data: {
                labels: datosSimulados.ingresosDiarios.map(d => {
                    const fecha = new Date(d.fecha);
                    return fecha.getDate() + '/' + (fecha.getMonth() + 1);
                }),
                datasets: [{
                    label: 'Ingresos Diarios',
                    data: datosSimulados.ingresosDiarios.map(d => d.ingresos),
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
                    legend: {
                        labels: { color: 'white' }
                    }
                },
                scales: {
                    y: {
                        beginAtZero: true,
                        ticks: {
                            color: 'white',
                            callback: function(value) {
                                return '$' + value.toLocaleString();
                            }
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

    function crearGraficoServicios() {
        const ctx = document.getElementById('grafico-servicios').getContext('2d');
        
        new Chart(ctx, {
            type: 'doughnut',
            data: {
                labels: datosSimulados.servicios.map(s => s.nombre),
                datasets: [{
                    data: datosSimulados.servicios.map(s => s.cantidad),
                    backgroundColor: datosSimulados.servicios.map(s => s.color),
                    borderWidth: 0
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: true,
                plugins: {
                    legend: {
                        position: 'bottom',
                        labels: { 
                            color: 'white',
                            padding: 20,
                            usePointStyle: true
                        }
                    }
                }
            }
        });
    }

    function crearGraficoHorarios() {
        const ctx = document.getElementById('grafico-horarios').getContext('2d');
        
        new Chart(ctx, {
            type: 'bar',
            data: {
                labels: datosSimulados.horarios.map(h => h.hora),
                datasets: [{
                    label: 'Turnos',
                    data: datosSimulados.horarios.map(h => h.cantidad),
                    backgroundColor: '#ff6600',
                    borderRadius: 8
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: { display: false }
                },
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

    function crearGraficoOcupacion() {
        const ctx = document.getElementById('grafico-ocupacion').getContext('2d');
        const ocupacion = 78;
        
        new Chart(ctx, {
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
                maintainAspectRatio: true,
                cutout: '70%',
                plugins: {
                    legend: { display: false }
                }
            }
        });
    }

    function crearGraficoSemanal() {
        const ctx = document.getElementById('grafico-semanal').getContext('2d');
        const diasSemana = ['Lun', 'Mar', 'Mié', 'Jue', 'Vie', 'Sáb'];
        const turnosSemana = [12, 15, 18, 20, 25, 30];
        
        new Chart(ctx, {
            type: 'line',
            data: {
                labels: diasSemana,
                datasets: [{
                    label: 'Turnos',
                    data: turnosSemana,
                    borderColor: '#ff6600',
                    backgroundColor: 'rgba(255, 102, 0, 0.2)',
                    borderWidth: 3,
                    fill: true,
                    tension: 0.4,
                    pointBackgroundColor: '#ff6600',
                    pointBorderColor: '#fff',
                    pointBorderWidth: 2
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: { display: false }
                },
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

    function llenarTablaClientesFrecuentes() {
        const tbody = document.getElementById('tabla-clientes-frecuentes');
        tbody.innerHTML = '';
        
        datosSimulados.clientesFrecuentes.forEach(cliente => {
            const row = tbody.insertRow();
            row.innerHTML = `
                <td>${cliente.nombre}</td>
                <td><span class="badge" style="background-color: #ff6600;">${cliente.turnos}</span></td>
                <td>${cliente.ultimoServicio}</td>
            `;
        });
    }

    // Función para actualizar datos en tiempo real (llamar desde la API)
    window.actualizarDatos = function(nuevosDatos) {
        Object.assign(datosSimulados, nuevosDatos);
        actualizarResumen();
        // Recrear gráficos con nuevos datos si es necesario
    };
});