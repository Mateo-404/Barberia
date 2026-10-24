import type { components } from "@/types/api"

type Turno = components["schemas"]["TurnoResponseDTO"]
type Servicio = components["schemas"]["ServicioResponseDTO"]

interface Db {
  servicios: Servicio[]
  turnos: Turno[]
  nextId: number
}

export const DEMO_KEY = "barberia-demo-db"

const SERVICIOS: Servicio[] = [
  { id: 1, tipo: "Corte clásico", precio: 5000 },
  { id: 2, tipo: "Corte + barba", precio: 7000 },
  { id: 3, tipo: "Afeitado", precio: 3500 },
]

const CLIENTES = [
  ["Juan", "Pérez"], ["Lucas", "Gómez"], ["Mateo", "Díaz"], ["Santi", "Ruiz"],
  ["Tomás", "Sosa"], ["Nico", "Vega"], ["Bruno", "Mora"], ["Facu", "Luna"],
  ["Ale", "Rojas"], ["Gonza", "Paz"],
]

const SLOTS = [
  "13:00", "13:30", "14:00", "14:30", "15:00", "15:30", "16:00",
  "16:30", "17:00", "17:30", "18:00", "18:30", "19:00", "19:30",
]

// PRNG determinista: la demo se ve igual para todos los visitantes.
function mulberry32(seed: number) {
  return () => {
    seed |= 0
    seed = (seed + 0x6d2b79f5) | 0
    let t = Math.imul(seed ^ (seed >>> 15), 1 | seed)
    t = (t + Math.imul(t ^ (t >>> 7), 61 | t)) ^ t
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296
  }
}

const pad = (n: number) => String(n).padStart(2, "0")
export const toLocalIso = (d: Date, slot: string) =>
  `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${slot}:00`

function seed(): Db {
  const rnd = mulberry32(404)
  const turnos: Turno[] = []
  let id = 1
  const add = (day: Date, count: number) => {
    const slots = [...SLOTS].sort(() => rnd() - 0.5).slice(0, count)
    for (const slot of slots) {
      const [nombre, apellido] = CLIENTES[Math.floor(rnd() * CLIENTES.length)]
      const s = SERVICIOS[Math.floor(rnd() * SERVICIOS.length)]
      turnos.push({
        id: id++,
        fechaHora: toLocalIso(day, slot),
        telefonoCliente: `11555${String(1000 + Math.floor(rnd() * 8999))}`,
        nombreCliente: nombre,
        apellidoCliente: apellido,
        emailCliente: `${nombre}.${apellido}@email.com`.toLowerCase(),
        idServicio: s.id!,
        tipoServicio: s.tipo!,
        precioServicio: s.precio!,
      })
    }
  }
  for (let off = -35; off <= 6; off++) {
    const day = new Date()
    day.setDate(day.getDate() + off)
    if (day.getDay() === 0) continue
    add(day, off === 0 ? 4 : off === -1 ? 3 : 1 + Math.floor(rnd() * 3))
  }
  return { servicios: SERVICIOS, turnos, nextId: id }
}

let cache: Db | null = null

export function getDb(): Db {
  if (cache) return cache
  try {
    const raw = localStorage.getItem(DEMO_KEY)
    if (raw) return (cache = JSON.parse(raw) as Db)
  } catch { /* storage bloqueado: seguimos en memoria */ }
  return (cache = seed())
}

export function saveDb() {
  try {
    localStorage.setItem(DEMO_KEY, JSON.stringify(cache))
  } catch { /* noop */ }
}

export function resetDb() {
  cache = null
  try {
    localStorage.removeItem(DEMO_KEY)
  } catch { /* noop */ }
}
