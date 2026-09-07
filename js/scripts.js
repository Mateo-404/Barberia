import { ENDPOINTS, HORARIOS_LABORALES } from "./config.js";

//<-- Constantes globales -->
let pasoActual = 0; // Controla el paso actual del formulario
const secciones = document.querySelectorAll(".seccion");
const regexTelefono = /^[0-9]{7,15}$/;
const regexEmail = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
const data = {
  fechaHora: "",
  cliente: {
    nombre: "",
    apellido: "",
    telefono: "",
    email: "",
  },
  servicio: {
    id: 0,
  },
};


// <-- API -->
const API = {
  async obtenerServicios() {
    try {
      const response = await fetch(ENDPOINTS.servicios);
      if (!response.ok) throw new Error("API: Error al obtener los servicios");
      return await response.json();
    } catch (error) {
      console.error(error);
      mostrarError("Hubo un error al cargar los servicios.");
      return [];
    }
  },
  async obtenerTurnosOcupados() {
    try {
      const response = await fetch(ENDPOINTS.turnos + "/findDateTimes");
      if (!response.ok) throw new Error("API: Error al obtener los turnos ocupados");
      return await response.json();
    } catch (error) {
      console.error(error);
      mostrarError("Hubo un error al cargar los turnos ocupados.");
      return [];
    }
  },
  async reservarTurno(data) {
    console.log("URL a la que va el POST:", ENDPOINTS.turnos);
    try {
      const res = await fetch(ENDPOINTS.turnos, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(data),
      });
      if (!res.ok) throw new Error("Error al reservar turno");
      const text = await res.text();
      // Si la respuesta del servidor es JSON, lo parseamos
      if (text) {
        const json = JSON.parse(text);
        console.log("✅ JSON recibido:", json);
      } else {
        console.log("✅ Turno reservado. No se recibió cuerpo en la respuesta.");
        document.getElementById("form-turno").reset();
      }
    } catch (err) {
      console.error(err);
      mostrarError("Hubo un error al reservar el turno.");
    }
  }
};


//<-- Utilidades (formato, validación, etc.) -->
// 1. Seccion: Servicios
const mostrarServicios = (servicios) => {
  const selectServicio = document.getElementById("servicio");
  
  servicios.forEach((servicio) => {
    const option = document.createElement("option");
    option.value = servicio.id;
    option.textContent = servicio.tipo;
    selectServicio.appendChild(option);
  });
}

// 2. Seccion: Fechas
// Convierte un array de 'YYYY-MM-DDTHH:mm' a un objeto agrupado por fecha con arrays de horas.
function mapearTurnosOcupados(arr) {
  const mapa = {};
  arr.forEach((item) => {
    const [fecha, hora] = item.split("T");
    if (!mapa[fecha]) mapa[fecha] = [];
    mapa[fecha].push(hora);
  });
  return mapa;
}

// Validaciones
const tieneValor = (input) => {
  // Si el input está oculto (por estar en otro paso), lo ignoramos
  if (input.offsetParent === null) return true;

  if (input.hasAttribute("required") && !input.value) {
    input.classList.add("is-invalid");
    return false;
  } else {
    input.classList.remove("is-invalid");
    return true;
  }
};
const validarTelefono = (telefono) => {
  if (!regexTelefono.test(telefono)) {
    mostrarError("Por favor, ingresá un teléfono válido (solo números, 7 a 15 dígitos).");
    return false;
  }
  return true;
};
const validarEmail = (email) => {
  if (!regexEmail.test(email)) {
    mostrarError("Por favor, ingresá un correo electrónico válido.");
    return false;
  }
  return true;
};

const mostrarError = (mensaje) => {
  const modalMensaje = document.getElementById("modalErrorMensaje");
  modalMensaje.textContent = mensaje;

  const modal = new bootstrap.Modal(document.getElementById("modalError"));
  modal.show();
}



//<-- Eventos y DOM -->
document.addEventListener("DOMContentLoaded", async function () {
  // 1. Pantalla: Seleccionar Servicio
  const servicios = await API.obtenerServicios();
  mostrarServicios(servicios);

  // 2. Pantalla: Seleccionar Fecha
  const turnosOcupadosRaw = await API.obtenerTurnosOcupados();
  const turnosOcupadosPorFecha = mapearTurnosOcupados(turnosOcupadosRaw);

  /**
   * Muestra los horarios disponibles para la fecha seleccionada,
   * ocultando los que ya están ocupados y mostrando un mensaje si no hay disponibles.
   */
  window.cargarHorariosDisponibles = function () {
    const inputFecha = document.getElementById("fecha");
    const contenedor = document.getElementById("horarios");

    // 1. Establecer la fecha mínima como hoy (solo la primera vez)
    if (!inputFecha.min) {
      const hoy = new Date().toISOString().split("T")[0];
      inputFecha.min = hoy;
    }

    // 2. Obtener fecha seleccionada
    const fecha = inputFecha.value;
    contenedor.innerHTML = "";
    if (!fecha) return;

    // 3. Filtrar horarios disponibles
    const ocupados = turnosOcupadosPorFecha[fecha] || [];
    const libres = HORARIOS_LABORALES.filter(h => !ocupados.includes(h));

    // 4. Mostrar mensaje si no hay horarios
    if (libres.length === 0) {
      contenedor.innerHTML = `<p class="text-warning mt-3">${"No quedan turnos disponibles para esta fecha, seleccione otra"}</p>`;
      return;
    }

    // 5. Crear botones de horarios disponibles
    const fragment = document.createDocumentFragment();
    libres.forEach(hora => {
      const btn = document.createElement("button");
      btn.type = "button";
      btn.className = "input-opcion";
      btn.textContent = hora;

      btn.onclick = () => {
        document.querySelectorAll(".input-opcion").forEach(b => b.classList.remove("selected"));
        btn.classList.add("selected");
        btn.dataset.selected = true;
      };

      fragment.appendChild(btn);
    });

    contenedor.appendChild(fragment);
  };

  // Botones de Navegación
  window.siguientePaso = function (paso) {
    const seccionActiva = document.querySelector(".seccion.activa");
    const inputs = seccionActiva.querySelectorAll("input, select");

    // <-- Validaciones -->
    // 1. Inputs completos
    for (let input of inputs) {
      if (!tieneValor(input)) {
        mostrarError("Por favor, completá todos los campos obligatorios.");
        return;
      }
    }
    
    // 2. Selección de horario
    if (seccionActiva.id === "paso-2") {
      const horarioSeleccionado = seccionActiva.querySelector(".input-opcion.selected");
      if (!horarioSeleccionado) {
        mostrarError("Por favor, seleccioná un horario.");
        return;
      }
    }

    secciones[paso - 1].classList.remove("activa");
    secciones[paso].classList.add("activa");
    pasoActual = paso;
  };
  
  window.volverPaso = function (paso) {
    secciones[pasoActual].classList.remove("activa");
    secciones[paso - 1].classList.add("activa");
    pasoActual = paso - 1;
  };

  window.enviarForm = async function () {
    const telefono = document.getElementById("telefono").value.trim();
    const email = document.getElementById("email").value.trim();
    const nombre = document.getElementById("nombre").value.trim();
    const apellido = document.getElementById("apellido").value.trim();
    const servicioId = parseInt(document.getElementById("servicio").value);
    const fecha = document.getElementById("fecha").value;
    const hora = document.querySelector(".selected")?.textContent || "";
    
    // Validación de datos
    //! Revisar si no se puede reemplazar por funcion
    if (!nombre || !apellido || !telefono || !email || !fecha || !hora || !servicioId) {
      mostrarError("Por favor, completá todos los campos obligatorios.");
      return;
    }

    if (!validarTelefono(telefono) || !validarEmail(email)) return;
    
    data.fechaHora = `${fecha}T${hora}`;
    data.cliente.nombre = nombre;
    data.cliente.apellido = apellido;
    data.cliente.telefono = telefono;
    data.cliente.email = email;
    data.servicio.id = servicioId;
    
    //* Log de datos para depuración
    console.log("🧾 Datos del turno a enviar:", data);
    
    await API.reservarTurno(data);

    window.location.href = "reserva-confirmada.html";
  }
});
