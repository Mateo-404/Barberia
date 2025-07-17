// scripts.js

let pasoActual = 1;
let horarioSeleccionado = null;

function siguientePaso(paso) {
  document.getElementById(`paso-${paso}`).classList.remove('activa');
  document.getElementById(`paso-${paso + 1}`).classList.add('activa');
}

function volverPaso(paso) {
  document.getElementById(`paso-${paso + 1}`).classList.remove('activa');
  document.getElementById(`paso-${paso}`).classList.add('activa');
}

// Cargar horarios disponibles dinámicamente
async function cargarHorariosDisponibles() {
  const fecha = document.getElementById('fecha').value;
  const horariosDiv = document.getElementById('horarios');
  horariosDiv.innerHTML = '';

  if (!fecha) return;

  try {
    const response = await fetch(`/turnos/ocupados?fecha=${fecha}`);
    const datos = await response.json(); // ["13:00", "14:30", ...]

    const horariosTrabajo = generarHorarios("13:00", "20:00", 30);
    horariosTrabajo.forEach(hora => {
      if (!datos.includes(hora)) {
        const btn = document.createElement('button');
        btn.textContent = hora;
        btn.type = 'button';
        btn.classList.add('btn');
        btn.onclick = () => seleccionarHorario(btn, hora);
        horariosDiv.appendChild(btn);
      }
    });
  } catch (error) {
    console.error('Error cargando horarios:', error);
  }
}

// Generador de horarios en intervalos de 30 minutos
function generarHorarios(inicio, fin, intervaloMin) {
  const horarios = [];
  let [h, m] = inicio.split(':').map(Number);
  const [hf, mf] = fin.split(':').map(Number);

  while (h < hf || (h === hf && m < mf)) {
    horarios.push(`${String(h).padStart(2, '0')}:${String(m).padStart(2, '0')}`);
    m += intervaloMin;
    if (m >= 60) {
      h += 1;
      m = m % 60;
    }
  }

  return horarios;
}

function seleccionarHorario(boton, hora) {
  document.querySelectorAll('#horarios button').forEach(b => b.classList.remove('selected'));
  boton.classList.add('selected');
  horarioSeleccionado = hora;
}

// Enviar formulario
document.getElementById('form-turno').addEventListener('submit', async function (e) {
  e.preventDefault();

  const data = {
    cliente: {
      nombre: document.getElementById('nombre').value,
      apellido: document.getElementById('apellido').value,
      telefono: document.getElementById('telefono').value,
      email: document.getElementById('email').value
    },
    fechaHora: `${document.getElementById('fecha').value}T${horarioSeleccionado}`,
    servicioId: parseInt(document.getElementById('servicio').value)
  };

  try {
    const res = await fetch('/turnos', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(data)
    });

    if (res.ok) {
      alert('Turno reservado con éxito.');
      window.location.href = 'index.html';
    } else {
      alert('Error al reservar turno');
    }
  } catch (err) {
    console.error(err);
    alert('Error de conexión.');
  }
});
