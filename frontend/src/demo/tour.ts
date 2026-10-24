import { driver, type DriveStep, type Driver } from "driver.js"
import "driver.js/dist/driver.css"
import "./tour.css"
import { DEMO_ADMIN } from "@/mocks/handlers"

const SEEN_KEY = "barberia-demo-tour-seen"

const $ = (sel: string) => document.querySelector<HTMLElement>(`[data-tour="${sel}"]`)
const sleep = (ms: number) => new Promise((r) => setTimeout(r, ms))

async function waitFor(sel: string, timeout = 4000) {
  const t0 = Date.now()
  while (Date.now() - t0 < timeout) {
    const el = $(sel)
    if (el) return el
    await sleep(80)
  }
  return null
}

// React escucha `input`; el setter nativo evita que React ignore el valor.
function fill(el: HTMLInputElement | null, value: string) {
  if (!el) return
  const set = Object.getOwnPropertyDescriptor(HTMLInputElement.prototype, "value")!.set!
  set.call(el, value)
  el.dispatchEvent(new Event("input", { bubbles: true }))
}

const input = (scope: string, sel: string) =>
  $(scope)?.querySelector<HTMLInputElement>(sel) ?? null

function nextOpenDay() {
  const d = new Date()
  do d.setDate(d.getDate() + 1)
  while (d.getDay() === 0)
  const p = (n: number) => String(n).padStart(2, "0")
  return `${d.getFullYear()}-${p(d.getMonth() + 1)}-${p(d.getDate())}`
}

interface Step {
  /** `data-tour` del elemento a resaltar (sin valor = popover centrado). */
  anchor?: string
  title: string
  text: string
  /** Condición que indica que el visitante ya hizo la acción por su cuenta. */
  done?: () => boolean
  /** Hace la acción por el visitante; devuelve el ancla del siguiente paso. */
  act?: () => Promise<string | null>
  nextLabel?: string
}

const click = (sel: string) => $(sel)?.click()

const STEPS: Step[] = [
  {
    title: "¡Bienvenido a la demo! 👋",
    text: "Te muestro cómo reserva un cliente y cómo lo ve el barbero. Podés hacer cada paso vos o dejar que lo haga por vos. Los datos son simulados y viven solo en tu navegador.",
    nextLabel: "Empezar",
  },
  {
    anchor: "servicios",
    title: "1 · Elegí un servicio",
    text: "Tocá uno de los servicios. Si preferís, avanzo yo.",
    done: () => !!$("servicios")?.querySelector("button.border-primary"),
    act: async () => {
      if (!$("servicios")?.querySelector("button.border-primary")) {
        $("servicios")?.querySelector<HTMLElement>("button")?.click()
        await sleep(150)
      }
      click("siguiente")
      return (await waitFor("fecha")) ? "fecha" : null
    },
  },
  {
    anchor: "fecha",
    title: "2 · Elegí el día",
    text: "Seleccioná una fecha. Se cargan los horarios libres de ese día.",
    done: () => !!$("horarios"),
    act: async () => {
      fill(input("fecha", "input[type=date]"), nextOpenDay())
      return (await waitFor("horarios")) ? "horarios" : null
    },
  },
  {
    anchor: "horarios",
    title: "3 · Elegí un horario",
    text: "Los horarios tachados ya están ocupados: el sistema los lee de los turnos existentes, así no hay doble reserva.",
    done: () => !!$("horarios")?.querySelector("button.bg-primary"),
    act: async () => {
      if (!$("horarios")?.querySelector("button.bg-primary")) {
        $("horarios")?.querySelector<HTMLElement>("button:not(:disabled)")?.click()
        await sleep(150)
      }
      click("siguiente")
      return (await waitFor("datos")) ? "datos" : null
    },
  },
  {
    anchor: "datos",
    title: "4 · Tus datos",
    text: "Nombre, apellido y teléfono. El formulario valida en vivo (React Hook Form + Zod).",
    done: () => {
      const v = (n: string) => input("datos", `input[name=${n}]`)?.value.trim()
      return !!(v("nombreCliente") && v("apellidoCliente") && v("telefonoCliente"))
    },
    act: async () => {
      fill(input("datos", "input[name=nombreCliente]"), "Alex")
      fill(input("datos", "input[name=apellidoCliente]"), "Visitante")
      fill(input("datos", "input[name=telefonoCliente]"), "1155550000")
      await sleep(200)
      return "reservar"
    },
    nextLabel: "Completar por mí",
  },
  {
    anchor: "reservar",
    title: "5 · Confirmá",
    text: "Al reservar, el turno se guarda y ese horario queda ocupado para el próximo cliente.",
    done: () => !!$("exito"),
    act: async () => {
      click("reservar")
      return (await waitFor("exito")) ? "exito" : null
    },
    nextLabel: "Reservar",
  },
  {
    anchor: "exito",
    title: "¡Turno reservado! ✅",
    text: "Ahora pasemos al lado del barbero: el panel de administración.",
    act: async () => {
      window.location.hash = "#/login"
      return (await waitFor("login-form")) ? "login-form" : null
    },
    nextLabel: "Ir al panel",
  },
  {
    anchor: "login-form",
    title: "Acceso del barbero",
    text: `Es un login real con JWT. En la demo las credenciales son ${DEMO_ADMIN.email} / ${DEMO_ADMIN.password}.`,
    done: () => !!$("kpis"),
    act: async () => {
      fill(input("login-form", "#email"), DEMO_ADMIN.email)
      fill(input("login-form", "#contrasenia"), DEMO_ADMIN.password)
      await sleep(150)
      $("login-form")?.querySelector<HTMLButtonElement>("button[type=submit]")?.click()
      return (await waitFor("kpis", 6000)) ? "kpis" : null
    },
    nextLabel: "Entrar por mí",
  },
  {
    anchor: "kpis",
    title: "Resumen del día",
    text: "Turnos de hoy, ingresos del mes y clientes. Fijate que el turno que acabás de reservar ya cuenta.",
  },
  {
    anchor: "admin-nav",
    title: "Secciones del panel",
    text: "Panel, Estadísticas (ingresos, horarios pico, clientes habituales) y Turnos (lista, calendario, editar y cancelar). Te llevo a Turnos.",
    act: async () => {
      window.location.hash = "#/admin/turnos"
      await sleep(600)
      return "admin-nav"
    },
    nextLabel: "Ver turnos",
  },
  {
    title: "Todo con el teclado ⌨️",
    text: "Apretá Shift+? en cualquier pantalla para ver los atajos. Eso es todo: explorá libremente. Podés reiniciar los datos o repetir la visita con los botones de abajo.",
    nextLabel: "Terminar",
  },
]

let active: Driver | null = null

export function startTour(from = 0) {
  active?.destroy()
  if (window.location.hash !== "#/" && window.location.hash !== "" && from === 0) {
    window.location.hash = "#/"
  }
  let poll = 0
  let busy = false
  let index = from

  const toDriveStep = (s: Step, i: number): DriveStep => ({
    element: s.anchor ? () => $(s.anchor!) ?? document.body : undefined,
    popover: {
      title: s.title,
      description: s.text,
      showButtons: ["next", "close"],
      nextBtnText: s.nextLabel ?? (s.act ? "Hacerlo por mí" : "Siguiente"),
      progressText: `${i + 1} / ${STEPS.length}`,
      onNextClick: async () => {
        if (busy) return
        busy = true
        try {
          const target = s.done?.() ? null : await s.act?.()
          if (s.act && !s.done?.() && target === null && i < STEPS.length - 1 && s.anchor) {
            // la acción no llegó a destino: dejamos al visitante reintentar
            return
          }
          advance()
        } finally {
          busy = false
        }
      },
    },
  })

  const d = driver({
    showProgress: true,
    allowClose: true,
    overlayOpacity: 0.65,
    stagePadding: 8,
    stageRadius: 14,
    popoverClass: "barber-tour",
    nextBtnText: "Siguiente",
    doneBtnText: "Terminar",
    steps: STEPS.map(toDriveStep),
    onDestroyed: () => {
      window.clearInterval(poll)
      active = null
      try { localStorage.setItem(SEEN_KEY, "1") } catch { /* noop */ }
    },
  })

  function advance() {
    index += 1
    if (index >= STEPS.length) d.destroy()
    else d.moveTo(index)
  }

  // Si el visitante hace la acción solo, pasamos al siguiente paso.
  poll = window.setInterval(() => {
    if (busy) return
    const s = STEPS[index]
    if (s?.done?.() && STEPS[index + 1] && s.act) {
      busy = true
      void (async () => {
        await sleep(450)
        // pasos con avance encadenado (p. ej. tocar "Siguiente" de la app)
        if (!s.anchor || ["servicios", "horarios"].includes(s.anchor)) await s.act!()
        advance()
        busy = false
      })()
    }
  }, 400)

  active = d
  d.drive(from)
}

export function shouldAutoStart() {
  try { return !localStorage.getItem(SEEN_KEY) } catch { return true }
}
