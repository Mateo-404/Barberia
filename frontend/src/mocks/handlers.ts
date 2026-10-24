import { http, HttpResponse, delay } from "msw"
import type { components } from "@/types/api"
import { getDb, saveDb } from "./db"

type Turno = components["schemas"]["TurnoResponseDTO"]

const BASE = import.meta.env.VITE_API_BASE_URL ?? "http://localhost:8080"
export const DEMO_ADMIN = { email: "admin@admin.com", password: "admin1234" }

const problem = (status: number, title: string, detail: string, type?: string) =>
  HttpResponse.json({ title, detail, type }, { status })

// JWT de juguete (sin firma real) para que AuthContext pueda leer `exp`.
function fakeJwt() {
  const b64 = (o: object) => btoa(JSON.stringify(o)).replace(/=/g, "")
  const exp = Math.floor(Date.now() / 1000) + 8 * 3600
  return `${b64({ alg: "none" })}.${b64({ sub: DEMO_ADMIN.email, exp })}.demo`
}

const isAuth = (req: Request) => req.headers.get("Authorization")?.startsWith("Bearer ")
const unauthorized = () =>
  problem(401, "No autorizado", "Token inválido", "/errors/invalid-token")

const day = (iso: string) => iso.slice(0, 10)
const localDay = (offset: number) => {
  const d = new Date()
  d.setDate(d.getDate() + offset)
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}-${String(d.getDate()).padStart(2, "0")}`
}

function panel() {
  const { turnos } = getDb()
  const now = new Date()
  const monthKey = (d: Date) => `${d.getFullYear()}-${d.getMonth()}`
  const prev = new Date(now.getFullYear(), now.getMonth() - 1, 1)
  const realizados = turnos.filter((t) => new Date(t.fechaHora) <= now)
  const sum = (ts: Turno[]) => ts.reduce((a, t) => a + t.precioServicio, 0)

  const ingresosDiarios = Array.from({ length: 30 }, (_, i) => {
    const d = localDay(i - 29)
    return {
      fecha: `${d}T00:00:00`,
      ingresoTotal: sum(realizados.filter((t) => day(t.fechaHora) === d)),
    }
  })
  const count = <K extends string | number>(key: (t: Turno) => K) => {
    const m = new Map<K, number>()
    realizados.forEach((t) => m.set(key(t), (m.get(key(t)) ?? 0) + 1))
    return m
  }
  return {
    turnosHoy: turnos.filter((t) => day(t.fechaHora) === localDay(0)).length,
    turnosAyer: turnos.filter((t) => day(t.fechaHora) === localDay(-1)).length,
    ingresosMes: sum(realizados.filter((t) => monthKey(new Date(t.fechaHora)) === monthKey(now))),
    ingresosMesAnterior: sum(realizados.filter((t) => monthKey(new Date(t.fechaHora)) === monthKey(prev))),
    cantClientes: new Set(turnos.map((t) => t.telefonoCliente)).size,
    ingresosDiarios,
    servicios: [...count((t) => t.idServicio)].map(([id, n]) => ({
      id,
      nombre: turnos.find((t) => t.idServicio === id)!.tipoServicio,
      cantidadRealizado: n,
    })),
    horarios: [...count((t) => Number(t.fechaHora.slice(11, 13)))]
      .map(([hora, n]) => ({ hora, cantidadRealizado: n }))
      .sort((a, b) => a.hora - b.hora),
  }
}

export const handlers = [
  http.get(`${BASE}/ping`, () => HttpResponse.text("pong")),

  http.post(`${BASE}/administradores/login`, async ({ request }) => {
    await delay(250)
    const body = (await request.json()) as { email?: string; contrasenia?: string }
    if (body.email !== DEMO_ADMIN.email || body.contrasenia !== DEMO_ADMIN.password) {
      return problem(401, "No autorizado", "Credenciales inválidas")
    }
    return HttpResponse.json({
      id: 1, nombre: "Admin", apellido: "Demo", email: DEMO_ADMIN.email, token: fakeJwt(),
    })
  }),

  http.get(`${BASE}/servicios`, () => HttpResponse.json(getDb().servicios)),

  http.get(`${BASE}/turnos/findDateTimes`, () =>
    HttpResponse.json(getDb().turnos.map((t) => t.fechaHora.slice(0, 16))),
  ),

  http.post(`${BASE}/turnos`, async ({ request }) => {
    await delay(300)
    const db = getDb()
    const b = (await request.json()) as {
      fechaHora: string; idServicio: number; telefonoCliente: string
      nombreCliente: string; apellidoCliente: string; emailCliente?: string
    }
    const servicio = db.servicios.find((s) => s.id === b.idServicio)
    if (!servicio) return problem(404, "No encontrado", "El servicio no existe")
    if (db.turnos.some((t) => t.fechaHora.slice(0, 16) === b.fechaHora.slice(0, 16))) {
      return problem(409, "Conflicto", "Ese horario ya está reservado. Elegí otro.")
    }
    const turno: Turno = {
      id: db.nextId++,
      fechaHora: b.fechaHora,
      telefonoCliente: b.telefonoCliente,
      nombreCliente: b.nombreCliente,
      apellidoCliente: b.apellidoCliente,
      emailCliente: b.emailCliente,
      idServicio: servicio.id!,
      tipoServicio: servicio.tipo!,
      precioServicio: servicio.precio!,
    }
    db.turnos.push(turno)
    saveDb()
    return HttpResponse.json(turno, { status: 201 })
  }),

  http.get(`${BASE}/turnos`, ({ request }) =>
    isAuth(request) ? HttpResponse.json(getDb().turnos) : unauthorized(),
  ),

  http.get(`${BASE}/turnos/:page/:size`, ({ request, params }) => {
    if (!isAuth(request)) return unauthorized()
    const page = Number(params.page)
    const size = Number(params.size)
    const sorted = [...getDb().turnos].sort((a, b) => b.fechaHora.localeCompare(a.fechaHora))
    const content = sorted.slice(page * size, page * size + size).map((t) => ({
      id: t.id,
      fechaHora: t.fechaHora,
      cliente: {
        nombre: t.nombreCliente, apellido: t.apellidoCliente,
        email: t.emailCliente, telefono: t.telefonoCliente,
      },
      servicio: { id: t.idServicio, tipo: t.tipoServicio, precio: t.precioServicio },
    }))
    return HttpResponse.json({
      content, totalElements: sorted.length, totalPages: Math.ceil(sorted.length / size),
      number: page, size, first: page === 0, last: (page + 1) * size >= sorted.length,
      numberOfElements: content.length, empty: content.length === 0,
    })
  }),

  http.patch(`${BASE}/turnos/:id`, async ({ request, params }) => {
    if (!isAuth(request)) return unauthorized()
    const db = getDb()
    const t = db.turnos.find((x) => x.id === Number(params.id))
    if (!t) return problem(404, "No encontrado", "El turno no existe")
    const patch = (await request.json()) as { fechaHora?: string; servicio?: { id: number } }
    if (patch.fechaHora) {
      if (db.turnos.some((x) => x.id !== t.id && x.fechaHora.slice(0, 16) === patch.fechaHora!.slice(0, 16))) {
        return problem(409, "Conflicto", "Ese horario ya está reservado.")
      }
      t.fechaHora = patch.fechaHora
    }
    if (patch.servicio) {
      const s = db.servicios.find((x) => x.id === patch.servicio!.id)
      if (s) Object.assign(t, { idServicio: s.id, tipoServicio: s.tipo, precioServicio: s.precio })
    }
    saveDb()
    return HttpResponse.json(t)
  }),

  http.delete(`${BASE}/turnos/:id`, ({ request, params }) => {
    if (!isAuth(request)) return unauthorized()
    const db = getDb()
    db.turnos = db.turnos.filter((x) => x.id !== Number(params.id))
    saveDb()
    return new HttpResponse(null, { status: 204 })
  }),

  http.get(`${BASE}/estadisticas/panel`, ({ request }) =>
    isAuth(request) ? HttpResponse.json(panel()) : unauthorized(),
  ),
]
