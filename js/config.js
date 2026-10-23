const API_BASE_URL = "https://barberia-production-547f.up.railway.app";

export const ENDPOINTS = {
  turnos: `${API_BASE_URL}/turnos`,
  clientes: `${API_BASE_URL}/clientes`,
  servicios: `${API_BASE_URL}/servicios`,
  administradores: `${API_BASE_URL}/administradores`,
  estadisticas: `${API_BASE_URL}/estadisticas`
};

export const HORARIOS_LABORALES = [
  '13:00', '13:30', '14:00', '14:30','15:00', '15:30', '16:00', '16:30',
  '17:00', '17:30', '18:00', '18:30', '19:00', '19:30', '20:00'
];
