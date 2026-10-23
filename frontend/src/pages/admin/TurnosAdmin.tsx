import { useCallback, useMemo, useRef, useState } from "react"
import { useQueryClient } from "@tanstack/react-query"
import {
  Pencil, Trash2, Search, List, Calendar, CalendarDays, Filter,
  RotateCcw, ChevronLeft, ChevronRight, X, GripVertical, ArrowRight,
} from "lucide-react"
import {
  useTodosTurnos, useActualizarTurno, useEliminarTurno,
} from "@/api/turnos"
import { useServicios } from "@/api/servicios"
import { StatusBadge } from "@/components/ui/StatusBadge"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { KpiCard } from "@/components/admin/KpiCard"
import { ClienteDetalle } from "@/components/admin/ClienteDetalle"
import {
  formatCurrency, formatFecha, formatHora, isHoy,
} from "@/lib/format"
import { useHotkeys } from "@/hooks/useHotkeys"
import { useRegisterShortcuts } from "@/hooks/useShortcutsRegistry"
import type { components } from "@/types/api"

type Turno = components["schemas"]["TurnoResponseDTO"]

const TIME_SLOTS = [
  "13:00", "13:30", "14:00", "14:30", "15:00", "15:30",
  "16:00", "16:30", "17:00", "17:30", "18:00", "18:30",
  "19:00", "19:30",
]

type Estado = "hoy" | "pendiente" | "confirmado"

function estadoDe(t: Turno): Estado {
  if (isHoy(t.fechaHora)) return "hoy"
  if (new Date(t.fechaHora) > new Date()) return "pendiente"
  return "confirmado"
}

function todayISO(): string {
  const d = new Date()
  const mm = String(d.getMonth() + 1).padStart(2, "0")
  const dd = String(d.getDate()).padStart(2, "0")
  return `${d.getFullYear()}-${mm}-${dd}`
}

function FiltroFecha({
  value,
  onChange,
  label,
  min,
}: {
  value: string
  onChange: (v: string) => void
  label: string
  min?: string
}) {
  return (
    <div className="relative">
      <CalendarDays className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
      <Input
        type="date"
        value={value}
        min={min}
        onChange={(e) => onChange(e.target.value)}
        aria-label={label}
        className="w-[9.5rem] pl-9 pr-2 text-sm [color-scheme:dark]"
      />
    </div>
  )
}

export default function TurnosAdmin() {
  const queryClient = useQueryClient()
  const { data: turnos, isLoading } = useTodosTurnos()
  const { data: servicios } = useServicios()
  const actualizar = useActualizarTurno()
  const eliminar = useEliminarTurno()

  const [vista, setVista] = useState<"lista" | "calendario">("lista")
  const [busqueda, setBusqueda] = useState("")
  const [filtroEstado, setFiltroEstado] = useState<Estado | "todos">("todos")
  const [filtroServicio, setFiltroServicio] = useState<"todos" | number>("todos")
  const [fechaDesde, setFechaDesde] = useState("")
  const [fechaHasta, setFechaHasta] = useState("")
  const [editando, setEditando] = useState<Turno | null>(null)
  const [eliminandoId, setEliminandoId] = useState<number | null>(null)
  const [diaSeleccionado, setDiaSeleccionado] = useState<string | null>(null)
  const [mesCalendario, setMesCalendario] = useState(() => {
    const now = new Date()
    return new Date(now.getFullYear(), now.getMonth(), 1)
  })

  const [clienteSel, setClienteSel] = useState<{ telefono: string; nombre: string } | null>(null)

  // Drag & drop state
  const [dragTurnoId, setDragTurnoId] = useState<number | null>(null)
  const [dragOverIso, setDragOverIso] = useState<string | null>(null)
  const [moverTurnoId, setMoverTurnoId] = useState<number | null>(null)
  const [moverFecha, setMoverFecha] = useState(todayISO())

  const searchRef = useRef<HTMLInputElement>(null)
  const hasModal = editando !== null || eliminandoId !== null || moverTurnoId !== null || clienteSel !== null

  const turnosOrdenados = useMemo(() => {
    if (!turnos) return []
    return [...turnos].sort(
      (a, b) => new Date(a.fechaHora).getTime() - new Date(b.fechaHora).getTime(),
    )
  }, [turnos])

  const kpis = useMemo(() => {
    const total = turnosOrdenados.length
    const pendientes = turnosOrdenados.filter((t) => estadoDe(t) === "pendiente").length
    const confirmados = total - pendientes
    const ingresosHoy = turnosOrdenados
      .filter((t) => isHoy(t.fechaHora))
      .reduce((acc, t) => acc + (t.precioServicio ?? 0), 0)
    return { total, confirmados, pendientes, ingresosHoy }
  }, [turnosOrdenados])

  const turnosFiltrados = useMemo(() => {
    const q = busqueda.trim().toLowerCase()
    return turnosOrdenados.filter((t) => {
      const estado = estadoDe(t)
      if (filtroEstado !== "todos" && estado !== filtroEstado) return false
      if (
        filtroServicio !== "todos" &&
        t.idServicio !== filtroServicio
      )
        return false
      if (q) {
        const match =
          `${t.nombreCliente} ${t.apellidoCliente} ${t.telefonoCliente}`
            .toLowerCase()
            .includes(q)
        if (!match) return false
      }
      const fecha = t.fechaHora.slice(0, 10)
      if (fechaDesde && fecha < fechaDesde) return false
      if (fechaHasta && fecha > fechaHasta) return false
      return true
    })
  }, [turnosOrdenados, busqueda, filtroEstado, filtroServicio, fechaDesde, fechaHasta])

  const turnosPorDia = useMemo(() => {
    const map = new Map<string, Turno[]>()
    for (const t of turnosOrdenados) {
      const key = t.fechaHora.slice(0, 10)
      const arr = map.get(key) ?? []
      arr.push(t)
      map.set(key, arr)
    }
    return map
  }, [turnosOrdenados])

  function limpiarFiltros() {
    setBusqueda("")
    setFiltroEstado("todos")
    setFiltroServicio("todos")
    setFechaDesde("")
    setFechaHasta("")
    setDiaSeleccionado(null)
  }

  function abrirCliente(telefono: string, nombre: string) {
    setDiaSeleccionado(null)
    setClienteSel({ telefono, nombre })
  }

  async function guardarEdicion() {
    if (!editando) return
    await actualizar.mutateAsync({
      id: editando.id,
      patch: { fechaHora: new Date(editando.fechaHora).toISOString() },
    })
    queryClient.invalidateQueries({ queryKey: ["turnos"] })
    queryClient.invalidateQueries({ queryKey: ["estadisticas"] })
    setEditando(null)
  }

  async function borrarTurno(id: number) {
    await eliminar.mutateAsync(id)
    queryClient.invalidateQueries({ queryKey: ["turnos"] })
    queryClient.invalidateQueries({ queryKey: ["estadisticas"] })
    setEliminandoId(null)
  }

  async function moverTurnoAFecha(turnoId: number, nuevaFechaISO: string) {
    const turno = turnosOrdenados.find((t) => t.id === turnoId)
    if (!turno) return
    const hora = turno.fechaHora.slice(11, 16)
    const fechaHora = `${nuevaFechaISO}T${hora}:00`
    await actualizar.mutateAsync({
      id: turnoId,
      patch: { fechaHora },
    })
    queryClient.invalidateQueries({ queryKey: ["turnos"] })
    queryClient.invalidateQueries({ queryKey: ["estadisticas"] })
    setDragTurnoId(null)
    setDragOverIso(null)
    setMoverTurnoId(null)
    setDiaSeleccionado(nuevaFechaISO)
  }

  // Calendar month navigation
  const mesAnterior = useCallback(() => {
    setMesCalendario(
      new Date(mesCalendario.getFullYear(), mesCalendario.getMonth() - 1, 1),
    )
  }, [mesCalendario])

  const mesSiguiente = useCallback(() => {
    setMesCalendario(
      new Date(mesCalendario.getFullYear(), mesCalendario.getMonth() + 1, 1),
    )
  }, [mesCalendario])

  const irAHoy = useCallback(() => {
    setMesCalendario(new Date())
  }, [])

  // Drag & drop handlers
  function handleDragStart(e: React.DragEvent, turnoId: number) {
    e.dataTransfer.effectAllowed = "move"
    e.dataTransfer.setData("text/plain", String(turnoId))
    setDragTurnoId(turnoId)
  }

  function handleDragOver(e: React.DragEvent, iso: string) {
    e.preventDefault()
    e.dataTransfer.dropEffect = "move"
    setDragOverIso(iso)
  }

  function handleDragLeave() {
    setDragOverIso(null)
  }

  function handleDrop(e: React.DragEvent, destinoISO: string) {
    e.preventDefault()
    const turnoId = Number(e.dataTransfer.getData("text/plain"))
    if (!turnoId) return
    const turno = turnosOrdenados.find((t) => t.id === turnoId)
    if (!turno) return
    const origenISO = turno.fechaHora.slice(0, 10)
    if (origenISO === destinoISO) {
      setDragTurnoId(null)
      setDragOverIso(null)
      return
    }
    if (destinoISO < todayISO()) return
    moverTurnoAFecha(turnoId, destinoISO)
  }

  function handleDragEnd() {
    setDragTurnoId(null)
    setDragOverIso(null)
  }

  // Touch fallback
  function confirmarMoverTurno() {
    if (moverTurnoId === null) return
    if (moverFecha < todayISO()) return
    moverTurnoAFecha(moverTurnoId, moverFecha)
  }

  const diasDelMes = useMemo(() => {
    const year = mesCalendario.getFullYear()
    const month = mesCalendario.getMonth()
    const first = new Date(year, month, 1)
    const offset = (first.getDay() + 6) % 7
    const daysInMonth = new Date(year, month + 1, 0).getDate()
    const cells: Array<{ day: number; iso: string } | null> = []
    for (let i = 0; i < offset; i++) cells.push(null)
    for (let d = 1; d <= daysInMonth; d++) {
      const iso = `${year}-${String(month + 1).padStart(2, "0")}-${String(d).padStart(2, "0")}`
      cells.push({ day: d, iso })
    }
    return cells
  }, [mesCalendario])

  const turnosDelDia = diaSeleccionado
    ? (turnosPorDia.get(diaSeleccionado) ?? [])
    : []

  const hotkeyMap = useMemo(
    () => ({
      Escape: () => {
        if (clienteSel) setClienteSel(null)
        else if (editando) setEditando(null)
        else if (eliminandoId !== null) setEliminandoId(null)
        else if (moverTurnoId !== null) setMoverTurnoId(null)
      },
      "/": () => {
        searchRef.current?.focus()
      },
      l: () => setVista("lista"),
      c: () => setVista("calendario"),
      ArrowLeft: () => {
        if (vista === "calendario") mesAnterior()
      },
      ArrowRight: () => {
        if (vista === "calendario") mesSiguiente()
      },
      "1": () => setFiltroEstado("todos"),
      "2": () => setFiltroEstado("confirmado"),
      "3": () => setFiltroEstado("pendiente"),
      "4": () => setFiltroEstado("hoy"),
      x: () => limpiarFiltros(),
      h: () => {
        if (vista === "calendario") irAHoy()
      },
      n: () => {
        if (vista !== "calendario") return
        const base = diaSeleccionado ?? todayISO()
        const days = diasDelMes.filter((c): c is { day: number; iso: string } => c !== null)
        const idx = days.findIndex((c) => c.iso === base)
        if (idx < 0) return
        if (idx < days.length - 1) setDiaSeleccionado(days[idx + 1].iso)
      },
      p: () => {
        if (vista !== "calendario") return
        const base = diaSeleccionado ?? todayISO()
        const days = diasDelMes.filter((c): c is { day: number; iso: string } => c !== null)
        const idx = days.findIndex((c) => c.iso === base)
        if (idx <= 0) return
        setDiaSeleccionado(days[idx - 1].iso)
      },
      j: () => {
        if (vista !== "calendario") return
        const days = diasDelMes.filter((c): c is { day: number; iso: string } => c !== null)
        if (days.length === 0) return
        if (!diaSeleccionado) { setDiaSeleccionado(days[0].iso); return }
        const idx = days.findIndex((c) => c.iso === diaSeleccionado)
        if (idx < 0) { setDiaSeleccionado(days[0].iso); return }
        if (idx < days.length - 1) setDiaSeleccionado(days[idx + 1].iso)
      },
      k: () => {
        if (vista !== "calendario") return
        const days = diasDelMes.filter((c): c is { day: number; iso: string } => c !== null)
        if (days.length === 0) return
        if (!diaSeleccionado) { setDiaSeleccionado(days[days.length - 1].iso); return }
        const idx = days.findIndex((c) => c.iso === diaSeleccionado)
        if (idx <= 0) return
        setDiaSeleccionado(days[idx - 1].iso)
      },
      e: () => {
        if (vista === "calendario" && diaSeleccionado && turnosDelDia.length > 0)
          setEditando(turnosDelDia[0])
      },
      Delete: () => {
        const t = vista === "calendario" && diaSeleccionado
          ? turnosDelDia[0]
          : turnosFiltrados[0]
        if (t) setEliminandoId(t.id)
      },
      Enter: () => {
        if (vista === "lista" && turnosFiltrados.length > 0) {
          const t = turnosFiltrados[0]
          abrirCliente(t.telefonoCliente, `${t.nombreCliente} ${t.apellidoCliente}`)
        }
      },
    }),
    [
      editando, eliminandoId, moverTurnoId, clienteSel,
      vista, mesAnterior, mesSiguiente, irAHoy,
      diaSeleccionado, diasDelMes, turnosDelDia, turnosFiltrados,
    ],
  )

  useHotkeys(hotkeyMap, !hasModal)

  useRegisterShortcuts("admin", [
    { key: "l", description: "Vista lista" },
    { key: "c", description: "Vista calendario" },
    { key: "/", description: "Buscar" },
    { key: "Escape", description: "Cerrar panel / modal" },
    { key: "1", description: "Filtro: todos" },
    { key: "2", description: "Filtro: confirmados" },
    { key: "3", description: "Filtro: pendientes" },
    { key: "4", description: "Filtro: hoy" },
    { key: "x", description: "Limpiar filtros" },
    { key: "h", description: "Ir a hoy (calendario)" },
    { key: "n", description: "Día siguiente (calendario)" },
    { key: "p", description: "Día anterior (calendario)" },
    { key: "j", description: "Seleccionar día siguiente" },
    { key: "k", description: "Seleccionar día anterior" },
    { key: "e", description: "Editar primer turno del día" },
    { key: "Delete", description: "Eliminar primer turno" },
    { key: "Enter", description: "Abrir cliente del turno" },
  ])

  const MESES = ["Enero", "Febrero", "Marzo", "Abril", "Mayo", "Junio", "Julio", "Agosto", "Septiembre", "Octubre", "Noviembre", "Diciembre"]

  return (
    <div className="space-y-6 page-enter">
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
        <div>
          <h1 className="text-2xl font-bold text-foreground">Gestión de Turnos</h1>
          <p className="text-sm text-muted-foreground mt-1">
            Administrá los turnos del negocio.
          </p>
        </div>
        <div className="flex items-center gap-2">
          <div className="flex bg-card rounded-xl p-1 border border-border">
            <button
              onClick={() => setVista("lista")}
              className={`flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium transition-colors active:scale-95 ${
                vista === "lista"
                  ? "bg-primary text-primary-foreground"
                  : "text-muted-foreground hover:text-foreground"
              }`}
            >
              <List className="w-4 h-4" />
              Lista
            </button>
            <button
              onClick={() => setVista("calendario")}
              className={`flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium transition-colors active:scale-95 ${
                vista === "calendario"
                  ? "bg-primary text-primary-foreground"
                  : "text-muted-foreground hover:text-foreground"
              }`}
            >
              <Calendar className="w-4 h-4" />
              Calendario
            </button>
          </div>
          <span className="hidden sm:inline text-xs text-muted-foreground whitespace-nowrap">
            <kbd className="px-1.5 py-0.5 rounded bg-muted text-[10px] font-mono">Shift+?</kbd> atajos
          </span>
        </div>
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <KpiCard label="Total turnos" value={String(kpis.total)} tone="blue" animated />
        <KpiCard label="Confirmados" value={String(kpis.confirmados)} tone="green" animated />
        <KpiCard label="Pendientes" value={String(kpis.pendientes)} tone="ocre" animated />
        <KpiCard
          label="Ingresos hoy"
          value={formatCurrency(kpis.ingresosHoy)}
          tone="cyan"
          animated
        />
      </div>

      <div className="flex flex-col md:flex-row gap-3 md:items-center md:justify-between">
        <div className="relative flex-1 max-w-sm">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
          <Input
            ref={searchRef}
            value={busqueda}
            onChange={(e) => setBusqueda(e.target.value)}
            placeholder="Buscar por cliente o teléfono"
            className="pl-9"
          />
        </div>
        <div className="flex flex-wrap items-center gap-2">
          <select
            value={filtroEstado}
            onChange={(e) => setFiltroEstado(e.target.value as Estado | "todos")}
            className="bg-card border-2 border-border text-foreground text-sm rounded-lg px-3 py-2.5 focus:outline-none focus:border-primary"
          >
            <option value="todos">Todos los estados</option>
            <option value="confirmado">Confirmado</option>
            <option value="pendiente">Pendiente</option>
            <option value="hoy">Hoy</option>
          </select>
          <select
            value={filtroServicio}
            onChange={(e) =>
              setFiltroServicio(
                e.target.value === "todos" ? "todos" : Number(e.target.value),
              )
            }
            className="bg-card border-2 border-border text-foreground text-sm rounded-lg px-3 py-2.5 focus:outline-none focus:border-primary"
          >
            <option value="todos">Todos los servicios</option>
            {servicios?.map((s) => (
              <option key={s.id} value={s.id}>
                {s.tipo}
              </option>
            ))}
          </select>
          <FiltroFecha
            value={fechaDesde}
            onChange={setFechaDesde}
            label="Desde"
          />
          <FiltroFecha
            value={fechaHasta}
            onChange={setFechaHasta}
            label="Hasta"
          />
          <Button variant="secondary" size="sm" onClick={limpiarFiltros}>
            <RotateCcw className="w-4 h-4 mr-1" />
            Limpiar
          </Button>
        </div>
      </div>

      <div className="flex items-center justify-between">
        <h2 className="text-lg font-semibold text-foreground flex items-center gap-2">
          Historial de Turnos
          <StatusBadge variant="info">
            {turnosFiltrados.length}
          </StatusBadge>
        </h2>
        <button className="inline-flex items-center gap-1.5 text-sm font-medium text-muted-foreground hover:text-foreground transition-colors">
          <Filter className="w-4 h-4" />
          Filtros Avanzados
        </button>
      </div>

      {isLoading ? (
        <div className="bg-card rounded-2xl h-48 animate-pulse border border-border" />
      ) : vista === "lista" ? (
        <div key="vista-lista" className="page-enter bg-card rounded-2xl border border-border overflow-hidden">
          <div className="overflow-x-auto hidden sm:block">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-border text-left">
                  <th className="p-3 text-muted-foreground font-medium">Fecha</th>
                  <th className="p-3 text-muted-foreground font-medium">Hora</th>
                  <th className="p-3 text-muted-foreground font-medium">Cliente</th>
                  <th className="p-3 text-muted-foreground font-medium">Servicio</th>
                  <th className="p-3 text-muted-foreground font-medium">Estado</th>
                  <th className="p-3 text-muted-foreground font-medium">Precio</th>
                  <th className="p-3 text-muted-foreground font-medium">Acciones</th>
                </tr>
              </thead>
              <tbody>
                {turnosFiltrados.length === 0 ? (
                  <tr>
                    <td colSpan={7} className="p-6 text-center text-muted-foreground">
                      No hay turnos que coincidan con los filtros.
                    </td>
                  </tr>
                ) : (
                  turnosFiltrados.map((t) => {
                    const estado = estadoDe(t)
                    const badge =
                      estado === "hoy"
                        ? { variant: "info" as const, label: "Hoy" }
                        : estado === "pendiente"
                          ? { variant: "warning" as const, label: "Pendiente" }
                          : { variant: "success" as const, label: "Confirmado" }
                    return (
                      <tr key={t.id} className="border-b border-border last:border-0 hover:bg-accent/50 transition-colors">
                        <td className="p-3 text-muted-foreground">
                          {formatFecha(t.fechaHora)}
                        </td>
                        <td className="p-3 text-muted-foreground">
                          {formatHora(t.fechaHora)}
                        </td>
                        <td className="p-3">
                          <button
                            onClick={() => abrirCliente(t.telefonoCliente, `${t.nombreCliente} ${t.apellidoCliente}`)}
                            className="text-left hover:bg-accent rounded-lg px-2 py-1 -mx-2 -my-1 transition-colors active:scale-95"
                          >
                            <p className="text-foreground font-medium">
                              {t.nombreCliente} {t.apellidoCliente}
                            </p>
                            <p className="text-xs text-muted-foreground">
                              {t.telefonoCliente}
                            </p>
                          </button>
                        </td>
                        <td className="p-3 text-muted-foreground">{t.tipoServicio}</td>
                        <td className="p-3">
                          <StatusBadge variant={badge.variant}>{badge.label}</StatusBadge>
                        </td>
                        <td className="p-3 text-foreground font-medium">
                          {formatCurrency(t.precioServicio)}
                        </td>
                        <td className="p-3">
                          <div className="flex items-center gap-1">
                            <button
                              onClick={() => setEditando(t)}
                              className="min-h-10 min-w-10 p-2.5 rounded-lg text-muted-foreground hover:text-foreground hover:bg-accent transition-colors active:scale-95"
                              aria-label="Editar turno"
                            >
                              <Pencil className="w-4 h-4" />
                            </button>
                            <button
                              onClick={() => setEliminandoId(t.id)}
                              className="min-h-10 min-w-10 p-2.5 rounded-lg text-muted-foreground hover:text-destructive hover:bg-destructive/10 transition-colors active:scale-95"
                              aria-label="Eliminar turno"
                            >
                              <Trash2 className="w-4 h-4" />
                            </button>
                          </div>
                        </td>
                      </tr>
                    )
                  })
                )}
              </tbody>
            </table>
          </div>
          {/* Mobile card list */}
          <div className="sm:hidden divide-y divide-border">
            {turnosFiltrados.length === 0 ? (
              <p className="p-6 text-center text-muted-foreground text-sm">
                No hay turnos que coincidan con los filtros.
              </p>
            ) : (
              turnosFiltrados.map((t) => {
                const estado = estadoDe(t)
                const badge =
                  estado === "hoy"
                    ? { variant: "info" as const, label: "Hoy" }
                    : estado === "pendiente"
                      ? { variant: "warning" as const, label: "Pendiente" }
                      : { variant: "success" as const, label: "Confirmado" }
                return (
                  <div key={t.id} className="p-4 space-y-2">
                    <div className="flex items-center justify-between">
                      <button
                        onClick={() => abrirCliente(t.telefonoCliente, `${t.nombreCliente} ${t.apellidoCliente}`)}
                        className="text-left hover:bg-accent rounded-lg px-2 py-1 -mx-2 -my-1 transition-colors active:scale-95"
                      >
                        <p className="text-foreground font-medium">
                          {t.nombreCliente} {t.apellidoCliente}
                        </p>
                      </button>
                      <StatusBadge variant={badge.variant}>{badge.label}</StatusBadge>
                    </div>
                    <div className="flex items-center gap-3 text-xs text-muted-foreground">
                      <span>{formatFecha(t.fechaHora)}</span>
                      <span>{formatHora(t.fechaHora)}</span>
                    </div>
                    <div className="flex items-center justify-between">
                      <div className="text-xs text-muted-foreground">
                        {t.tipoServicio} · {formatCurrency(t.precioServicio)}
                      </div>
                      <div className="flex items-center gap-1">
                        <button
                          onClick={() => setEditando(t)}
                          className="min-h-10 min-w-10 p-2.5 rounded-lg text-muted-foreground hover:text-foreground hover:bg-accent transition-colors active:scale-95"
                          aria-label="Editar turno"
                        >
                          <Pencil className="w-4 h-4" />
                        </button>
                        <button
                          onClick={() => setEliminandoId(t.id)}
                          className="min-h-10 min-w-10 p-2.5 rounded-lg text-muted-foreground hover:text-destructive hover:bg-destructive/10 transition-colors active:scale-95"
                          aria-label="Eliminar turno"
                        >
                          <Trash2 className="w-4 h-4" />
                        </button>
                      </div>
                    </div>
                  </div>
                )
              })
            )}
          </div>
        </div>
      ) : (
        <div key="vista-calendario" className="page-enter space-y-4">
          <div className="bg-card rounded-2xl border border-border p-4">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-lg font-semibold text-foreground">
                {MESES[mesCalendario.getMonth()]} {mesCalendario.getFullYear()}
              </h3>
              <div className="flex gap-1">
                <button
                  onClick={mesAnterior}
                  className="min-h-10 min-w-10 p-2 rounded-lg text-muted-foreground hover:text-foreground hover:bg-accent active:scale-95 transition-all"
                  aria-label="Mes anterior"
                >
                  <ChevronLeft className="w-5 h-5" />
                </button>
                <button
                  onClick={irAHoy}
                  className="px-3 py-2 rounded-lg text-sm font-medium text-muted-foreground hover:text-foreground hover:bg-accent active:scale-95 transition-all"
                >
                  Hoy
                </button>
                <button
                  onClick={mesSiguiente}
                  className="min-h-10 min-w-10 p-2 rounded-lg text-muted-foreground hover:text-foreground hover:bg-accent active:scale-95 transition-all"
                  aria-label="Mes siguiente"
                >
                  <ChevronRight className="w-5 h-5" />
                </button>
              </div>
            </div>
            <div className="grid grid-cols-7 gap-1 text-center mb-1">
              {["L", "M", "X", "J", "V", "S", "D"].map((d) => (
                <span key={d} className="text-xs text-muted-foreground font-medium py-1">
                  {d}
                </span>
              ))}
            </div>
            <div key={`${mesCalendario.getFullYear()}-${mesCalendario.getMonth()}`} className="page-enter grid grid-cols-7 gap-1">
              {diasDelMes.map((cell, i) => {
                if (!cell) return <div key={`e-${i}`} />
                const turnos = turnosPorDia.get(cell.iso) ?? []
                const esSeleccionado = diaSeleccionado === cell.iso
                const esHoyDia = cell.iso === todayISO()
                const esDestinoDrop = dragOverIso === cell.iso && dragTurnoId !== null
                return (
                  <button
                    key={cell.iso}
                    onClick={() => setDiaSeleccionado(cell.iso)}
                    onDragOver={(e) => handleDragOver(e, cell.iso)}
                    onDragLeave={handleDragLeave}
                    onDrop={(e) => handleDrop(e, cell.iso)}
                    className={`relative min-h-20 sm:min-h-24 rounded-lg p-1 flex flex-col items-stretch gap-0.5 overflow-hidden transition-all duration-150 ${
                      esDestinoDrop
                        ? "ring-2 ring-primary ring-offset-1 ring-offset-background scale-105"
                        : esSeleccionado
                          ? "bg-primary/10 ring-2 ring-primary"
                          : esHoyDia
                            ? "border border-primary"
                            : turnos.length > 0
                              ? "hover:bg-accent"
                              : "hover:bg-accent"
                    }`}
                  >
                    <span
                      className={`text-xs text-left pl-1 ${
                        esSeleccionado
                          ? "text-primary font-bold"
                          : esHoyDia
                            ? "text-primary font-bold"
                            : "text-muted-foreground"
                      }`}
                    >
                      {cell.day}
                    </span>
                    {turnos.slice(0, 3).map((t) => (
                      <span
                        key={t.id}
                        className={`text-[9px] sm:text-[10px] leading-tight px-1 py-0.5 rounded truncate text-left ${
                          esSeleccionado
                            ? "bg-primary text-primary-foreground"
                            : "bg-primary/20 text-foreground"
                        }`}
                      >
                        {formatHora(t.fechaHora)} {t.tipoServicio}
                      </span>
                    ))}
                    {turnos.length > 3 && (
                      <span className="text-[9px] sm:text-[10px] pl-1 text-muted-foreground text-left leading-tight">
                        +{turnos.length - 3} más
                      </span>
                    )}
                  </button>
                )
              })}
            </div>
          </div>

          <div className="bg-card rounded-2xl border border-border overflow-hidden">
            <div className="p-4 border-b border-border flex items-center justify-between">
              <h3 className="font-semibold text-foreground">
                {diaSeleccionado ? formatFecha(`${diaSeleccionado}T00:00:00`) : "Turnos del día"}
              </h3>
              <StatusBadge variant="info">{turnosDelDia.length}</StatusBadge>
            </div>
            {turnosDelDia.length === 0 ? (
              <div className="p-4 text-sm text-muted-foreground text-center">
                {dragTurnoId !== null ? "Soltá aquí para mover el turno" : "Seleccioná un día o no hay turnos en esta fecha."}
              </div>
            ) : (
              <>
                {/* Desktop table */}
                <div className="overflow-x-auto hidden sm:block">
                  <table className="w-full text-sm">
                    <thead>
                      <tr className="border-b border-border text-left">
                        <th className="p-3 text-muted-foreground font-medium w-8"></th>
                        <th className="p-3 text-muted-foreground font-medium">Hora</th>
                        <th className="p-3 text-muted-foreground font-medium">Cliente</th>
                        <th className="p-3 text-muted-foreground font-medium">Servicio</th>
                        <th className="p-3 text-muted-foreground font-medium">Precio</th>
                        <th className="p-3 text-muted-foreground font-medium">Acciones</th>
                      </tr>
                    </thead>
                    <tbody>
                      {turnosDelDia.map((t) => (
                        <tr
                          key={t.id}
                          draggable
                          onDragStart={(e) => handleDragStart(e, t.id)}
                          onDragEnd={handleDragEnd}
                          className={`border-b border-border last:border-0 transition-colors ${
                            dragTurnoId === t.id ? "opacity-50" : "hover:bg-accent/50"
                          }`}
                        >
                          <td className="p-3 text-muted-foreground cursor-grab active:cursor-grabbing">
                            <GripVertical className="w-4 h-4" />
                          </td>
                          <td className="p-3 text-muted-foreground">{formatHora(t.fechaHora)}</td>
                          <td className="p-3">
                            <button
                              onClick={() => abrirCliente(t.telefonoCliente, `${t.nombreCliente} ${t.apellidoCliente}`)}
                              className="text-left hover:bg-accent rounded-lg px-2 py-1 -mx-2 -my-1 transition-colors active:scale-95"
                            >
                              <p className="text-foreground font-medium">
                                {t.nombreCliente} {t.apellidoCliente}
                              </p>
                              <p className="text-xs text-muted-foreground">{t.telefonoCliente}</p>
                            </button>
                          </td>
                          <td className="p-3 text-muted-foreground">{t.tipoServicio}</td>
                          <td className="p-3 text-foreground font-medium">
                            {formatCurrency(t.precioServicio)}
                          </td>
                          <td className="p-3">
                            <div className="flex items-center gap-1">
                              <button
                                onClick={() => setEditando(t)}
                                className="min-h-10 min-w-10 p-2.5 rounded-lg text-muted-foreground hover:text-foreground hover:bg-accent transition-colors active:scale-95"
                                aria-label="Editar turno"
                              >
                                <Pencil className="w-4 h-4" />
                              </button>
                              <button
                                onClick={() => {
                                  setMoverTurnoId(t.id)
                                  setMoverFecha(t.fechaHora.slice(0, 10))
                                }}
                                className="min-h-10 min-w-10 p-2.5 rounded-lg text-muted-foreground hover:text-primary hover:bg-accent transition-colors active:scale-95"
                                aria-label="Mover turno"
                                title="Mover a otra fecha"
                              >
                                <ArrowRight className="w-4 h-4" />
                              </button>
                            </div>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
                {/* Mobile card list */}
                <div className="sm:hidden divide-y divide-border">
                  {turnosDelDia.map((t) => (
                    <div key={t.id} className="p-4 space-y-2">
                      <div className="flex items-center justify-between">
                        <button
                          onClick={() => abrirCliente(t.telefonoCliente, `${t.nombreCliente} ${t.apellidoCliente}`)}
                          className="text-left hover:bg-accent rounded-lg px-2 py-1 -mx-2 -my-1 transition-colors active:scale-95"
                        >
                          <p className="text-foreground font-medium">
                            {t.nombreCliente} {t.apellidoCliente}
                          </p>
                        </button>
                        <span className="text-xs text-muted-foreground">{formatHora(t.fechaHora)}</span>
                      </div>
                      <div className="flex items-center justify-between">
                        <span className="text-xs text-muted-foreground">
                          {t.tipoServicio} · {formatCurrency(t.precioServicio)}
                        </span>
                        <div className="flex items-center gap-1">
                          <button
                            onClick={() => setEditando(t)}
                            className="min-h-10 min-w-10 p-2.5 rounded-lg text-muted-foreground hover:text-foreground hover:bg-accent transition-colors active:scale-95"
                            aria-label="Editar turno"
                          >
                            <Pencil className="w-4 h-4" />
                          </button>
                          <button
                            onClick={() => {
                              setMoverTurnoId(t.id)
                              setMoverFecha(t.fechaHora.slice(0, 10))
                            }}
                            className="min-h-10 min-w-10 p-2.5 rounded-lg text-muted-foreground hover:text-primary hover:bg-accent transition-colors active:scale-95"
                            aria-label="Mover turno"
                            title="Mover a otra fecha"
                          >
                            <ArrowRight className="w-4 h-4" />
                          </button>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              </>
            )}
          </div>
        </div>
      )}

      {clienteSel && (
        <ClienteDetalle
          telefono={clienteSel.telefono}
          nombre={clienteSel.nombre}
          turnos={turnosOrdenados}
          onVolver={() => setClienteSel(null)}
          onEditar={(t) => { setClienteSel(null); setEditando(t) }}
          onEliminar={(id) => { setClienteSel(null); setEliminandoId(id) }}
          onMover={(t) => { setClienteSel(null); setMoverTurnoId(t.id); setMoverFecha(t.fechaHora.slice(0, 10)) }}
        />
      )}

      {editando && (
        <EditarModal
          turno={editando}
          guardando={actualizar.isPending}
          onCancel={() => setEditando(null)}
          onChangeFecha={(iso) => setEditando({ ...editando, fechaHora: iso })}
          onGuardar={guardarEdicion}
        />
      )}

      {eliminandoId !== null && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm animate-in fade-in-0">
          <div className="bg-card rounded-2xl border border-border p-6 max-w-sm w-full space-y-4 animate-in fade-in-0 zoom-in-95 duration-200">
            <h3 className="text-lg font-semibold text-foreground">¿Eliminar turno?</h3>
            <p className="text-sm text-muted-foreground">
              Esta acción no se puede deshacer.
            </p>
            <div className="flex justify-end gap-2">
              <Button variant="secondary" onClick={() => setEliminandoId(null)}>
                Cancelar
              </Button>
              <Button
                variant="destructive"
                onClick={() => borrarTurno(eliminandoId)}
                disabled={eliminar.isPending}
              >
                Eliminar
              </Button>
            </div>
          </div>
        </div>
      )}

      {moverTurnoId !== null && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm animate-in fade-in-0">
          <div className="bg-card rounded-2xl border border-border p-6 max-w-sm w-full space-y-4 animate-in fade-in-0 zoom-in-95 duration-200">
            <div className="flex items-center justify-between">
              <h3 className="text-lg font-semibold text-foreground">Mover turno</h3>
              <button
                onClick={() => setMoverTurnoId(null)}
                className="p-2 rounded-lg text-muted-foreground hover:text-foreground hover:bg-accent"
                aria-label="Cerrar"
              >
                <X className="w-5 h-5" />
              </button>
            </div>
            <div className="space-y-1.5">
              <label className="text-sm text-foreground">Nueva fecha</label>
              <Input
                type="date"
                value={moverFecha}
                min={todayISO()}
                onChange={(e) => setMoverFecha(e.target.value)}
                className="w-full [color-scheme:dark]"
              />
              {moverFecha < todayISO() && (
                <p className="text-xs text-destructive">
                  No se pueden mover turnos a fechas pasadas.
                </p>
              )}
            </div>
            <div className="flex justify-end gap-2">
              <Button variant="secondary" onClick={() => setMoverTurnoId(null)}>
                Cancelar
              </Button>
              <Button
                onClick={confirmarMoverTurno}
                disabled={actualizar.isPending || moverFecha < todayISO()}
              >
                {actualizar.isPending ? "Moviendo…" : "Mover"}
              </Button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

function EditarModal({
  turno,
  guardando,
  onCancel,
  onChangeFecha,
  onGuardar,
}: {
  turno: Turno
  guardando: boolean
  onCancel: () => void
  onChangeFecha: (iso: string) => void
  onGuardar: () => void
}) {
  const [fecha, hora] = useMemo(() => {
    const d = new Date(turno.fechaHora)
    const f = `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}-${String(d.getDate()).padStart(2, "0")}`
    const h = TIME_SLOTS.includes(
      `${String(d.getHours()).padStart(2, "0")}:${String(d.getMinutes()).padStart(2, "0")}`,
    )
      ? `${String(d.getHours()).padStart(2, "0")}:${String(d.getMinutes()).padStart(2, "0")}`
      : TIME_SLOTS[0]
    return [f, h]
  }, [turno])

  const [nuevaFecha, setNuevaFecha] = useState(fecha)
  const [nuevaHora, setNuevaHora] = useState(hora)
  const [error, setError] = useState(false)

  function guardar() {
    const valor = `${nuevaFecha}T${nuevaHora}:00`
    if (new Date(valor) <= new Date()) {
      setError(true)
      return
    }
    onChangeFecha(valor)
    onGuardar()
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm animate-in fade-in-0" role="dialog" aria-modal="true">
      <div className="bg-card rounded-2xl border border-border p-6 max-w-sm w-full space-y-4 animate-in fade-in-0 zoom-in-95 duration-200">
        <div className="flex items-center justify-between">
          <h3 className="text-lg font-semibold text-foreground">Editar turno</h3>
          <button
            onClick={onCancel}
            className="p-2 rounded-lg text-muted-foreground hover:text-foreground hover:bg-accent"
            aria-label="Cerrar"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        <div className="space-y-1">
          <p className="text-xs text-muted-foreground">
            {turno.nombreCliente} {turno.apellidoCliente} · {turno.tipoServicio}
          </p>
        </div>

        <div className="grid grid-cols-2 gap-3">
          <div className="space-y-1.5">
            <label className="text-sm text-foreground">Fecha</label>
            <Input
              type="date"
              value={nuevaFecha}
              min={todayISO()}
              onChange={(e) => {
                setNuevaFecha(e.target.value)
                setError(false)
              }}
              className="[color-scheme:dark]"
            />
          </div>
          <div className="space-y-1.5">
            <label className="text-sm text-foreground">Hora</label>
            <select
              value={nuevaHora}
              onChange={(e) => {
                setNuevaHora(e.target.value)
                setError(false)
              }}
              className="w-full bg-input border-2 border-primary text-foreground text-sm rounded-xl px-4 py-2.5 focus:outline-none"
            >
              {TIME_SLOTS.map((h) => (
                <option key={h} value={h}>
                  {h}
                </option>
              ))}
            </select>
          </div>
        </div>

        {error && (
          <p className="text-sm text-destructive">
            La fecha debe ser futura.
          </p>
        )}

        <div className="flex justify-end gap-2">
          <Button variant="secondary" onClick={onCancel}>
            Cancelar
          </Button>
          <Button onClick={guardar} disabled={guardando}>
            {guardando ? "Guardando…" : "Guardar"}
          </Button>
        </div>
      </div>
    </div>
  )
}
