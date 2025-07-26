import { ENDPOINTS, HORARIOS_LABORALES, MENSAJES } from "./config.js";

// <-- Servicios -->
const obtenerServicios = async () => {
  try {
    const response = await fetch(ENDPOINTS.servicios);
    if (!response.ok) throw new Error("Error al obtener los servicios");
    return await response.json();
  } catch (error) {
    console.error(error);
    return [];
  }
};

// <-- Turnos -->
const obtenerTurnosOcupados = async () => {
  try {
    const response = await fetch(ENDPOINTS.turnos + "/findDateTimes");
    if (!response.ok) throw new Error("Error al obtener los turnos ocupados");
    return await response.json();
  } catch (error) {
    console.error(error);
    return [];
  }
};

/**
 * Formatea un arreglo como ['2025-08-15T10:00'] a:
 * {
 *   '2025-08-15': ['10:00'],
 *   '2025-08-16': ['11:30']
 * }
 */
function mapearTurnosOcupados(arr) {
  const mapa = {};
  arr.forEach((item) => {
    const [fecha, hora] = item.split("T");
    if (!mapa[fecha]) mapa[fecha] = [];
    mapa[fecha].push(hora);
  });
  return mapa;
}

// Iniciar app una vez que DOM y datos estén listos
document.addEventListener("DOMContentLoaded", async function () {
  // Cargar servicios en el select
  const servicios = await obtenerServicios();
  const selectServicio = document.getElementById("servicio");
  servicios.forEach((servicio) => {
    const option = document.createElement("option");
    option.value = servicio.id;
    option.textContent = servicio.tipo;
    selectServicio.appendChild(option);
  });

  //* Habria que optimizar para que solo haga el fetch cuando está en la pantalla de fecha
  const turnosOcupadosRaw = await obtenerTurnosOcupados();
  const turnosOcupadosPorFecha = mapearTurnosOcupados(turnosOcupadosRaw);

  const secciones = document.querySelectorAll(".seccion");
  let pasoActual = 0;

  /**
   * Muestra los horarios disponibles para la fecha seleccionada,
   * ocultando los que ya están ocupados y mostrando un mensaje si no hay disponibles.
   */
  window.cargarHorariosDisponibles = function () {
    // Establecer la fecha mínima como hoy
    const inputFecha = document.getElementById('fecha');
    const hoy = new Date();
    const yyyy = hoy.getFullYear();
    const mm = String(hoy.getMonth() + 1).padStart(2, '0');
    const dd = String(hoy.getDate()).padStart(2, '0');
    inputFecha.min = `${yyyy}-${mm}-${dd}`;
    
    // Llamada a la API
    const contenedor = document.getElementById("horarios");
    contenedor.innerHTML = "";

    const fechaSeleccionada = document.getElementById("fecha").value;
    if (!fechaSeleccionada) return;

    const ocupados = turnosOcupadosPorFecha[fechaSeleccionada] || [];
    const horariosLibres = HORARIOS_LABORALES.filter(
      (hora) => !ocupados.includes(hora)
    );

    if (horariosLibres.length === 0) {
      const mensaje = document.createElement("p");
      mensaje.textContent = MENSAJES.fechaSinTurnos;
      mensaje.className = "text-warning mt-3";
      contenedor.appendChild(mensaje);
      return;
    }

    horariosLibres.forEach((hora) => {
      const btn = document.createElement("button");
      btn.type = "button";
      btn.className = "input-opcion";
      btn.textContent = hora;

      btn.onclick = () => {
        document
          .querySelectorAll(".input-opcion")
          .forEach((b) => b.classList.remove("selected"));
        btn.classList.add("selected");
        btn.dataset.selected = true;
      };

      contenedor.appendChild(btn);
    });
  };

  // Navegación entre pasos
  window.siguientePaso = function (paso) {
    const seccionActiva = document.querySelector(".seccion.activa");
    const inputs = seccionActiva.querySelectorAll("input, select");
    let valido = true; 

    // Verificamos que hayan inputs requeridos y que estén completos
    inputs.forEach((input) => {
      if (input.hasAttribute("required") && !input.value) {
        input.classList.add("is-invalid");
        valido = false;
      } else {
        input.classList.remove("is-invalid");
      }
    });
    
    
    // Validar selección de horario en el paso de horarios (paso 2)
    if (seccionActiva.id === "paso-2") {
      const horarioSeleccionado = seccionActiva.querySelector(".input-opcion.selected");
      if (!horarioSeleccionado) {
        valido = false;
        alert("Por favor, seleccioná un horario.");
        return;
      }
    }
    
    if (!valido) {
      alert("maquinola, no estan completos los campos requeridos");
      return;
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

  /**
   * Captura los datos del formulario y los muestra en consola
   */
  function obtenerHorarioSeleccionado() {
    const seleccionado = document.querySelector(".input-opcion.selected");
    return seleccionado ? seleccionado.textContent : null;
  }

  document.getElementById("form-turno").addEventListener("submit", function (e) {
      e.preventDefault();

      // Formato de los datos a enviar
      const data = {
        fechaHora: `${
          document.getElementById("fecha").value
        }T${obtenerHorarioSeleccionado()}`, // ejemplo: "2025-08-16T14:00"
        cliente: {
          nombre: document.getElementById("nombre").value,
          apellido: document.getElementById("apellido").value,
          telefono: document.getElementById("telefono").value,
          email: document.getElementById("email").value,
        },
        servicio: {
          id: parseInt(document.getElementById("servicio").value),
        },
      };

      console.log("🧾 Datos del turno a enviar:", data);
      

      // Fetch a la API para reservar el turno
      fetch(ENDPOINTS.turnos, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(data),
      })
        .then(async (res) => {
          if (!res.ok) throw new Error("Error al reservar turno");

          // Verificamos si hay contenido en la respuesta
          const text = await res.text();
          if (text) {
            const json = JSON.parse(text);
            console.log("✅ JSON recibido:", json);
          } else {
            console.log(
              "✅ Turno reservado. No se recibió cuerpo en la respuesta."
            );
          }

          alert(MENSAJES.turnoConfirmado);
        })
        .catch((err) => {
          console.error(err);
          alert("Hubo un error al reservar el turno.");
        });
    });
});
