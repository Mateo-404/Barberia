import { useMemo } from "react"
import {
  ArrowLeft, Phone, Mail, Calendar as CalendarIcon,
} from "lucide-react"
import {
  PieChart, Pie, Cell, Tooltip, ResponsiveContainer,
  BarChart, Bar, XAxis, YAxis, CartesianGrid,
} from "recharts"
import { KpiCard } from "@/components/admin/KpiCard"
import { StatusBadge } from "@/components/ui/StatusBadge"
import { Button } from "@/components/ui/button"
import {
  formatCurrency, formatFecha, formatHora, isHoy,
} from "@/lib/format"
import type { components } from "@/types/api"
import { Pencil, Trash2, ArrowRight } from "lucide-react"

type Turno = components["schemas"]["TurnoResponseDTO"]
type Estado = "hoy" | "pendiente" | "confirmado"

const PALETA = ["#ff6600", "#ffc107", "#dc3545", "#f97316", "#999999"]

function estadoDe(t: Turno): Estado {
  if (isHoy(t.fechaHora)) return "hoy"
  if (new Date(t.fechaHora) > new Date()) return "pendiente"
  return "confirmado"
}

export function ClienteDetalle({
  telefono,
  nombre,
  turnos,
  onVolver,
  onEditar,
  onEliminar,
  onMover,
}: {
  telefono: string
  nombre: string
  turnos: Turno[]
  onVolver: () => void
  onEditar: (turno: Turno) => void
  onEliminar: (id: number) => void
  onMover: (turno: Turno) => void
}) {
  const clienteTurnos = useMemo(
    () =>
      [...turnos]
        .filter((t) => t.telefonoCliente === telefono)
        .sort((a, b) => new Date(b.fechaHora).getTime() - new Date(a.fechaHora).getTime()),
    [turnos, telefono],
  )

  const email = clienteTurnos[0]?.emailCliente ?? ""

  const primerTurno = useMemo(() => {
    const sorted = [...clienteTurnos].sort(
      (a, b) => new Date(a.fechaHora).getTime() - new Date(b.fechaHora).getTime(),
    )
    return sorted[0]
  }, [clienteTurnos])

  const ultimoTurno = clienteTurnos[0]

  const proximaReserva = useMemo(
    () =>
      clienteTurnos.find(
        (t) => new Date(t.fechaHora) > new Date() && estadoDe(t) !== "confirmado",
      ),
    [clienteTurnos],
  )

  const totalInvertido = useMemo(
    () => clienteTurnos.reduce((acc, t) => acc + (t.precioServicio ?? 0), 0),
    [clienteTurnos],
  )

  const ticketPromedio = clienteTurnos.length > 0
    ? Math.round(totalInvertido / clienteTurnos.length)
    : 0

  const serviciosDistintos = useMemo(() => {
    const set = new Set(clienteTurnos.map((t) => t.tipoServicio))
    return set.size
  }, [clienteTurnos])

  const distribucionServicios = useMemo(() => {
    const map = new Map<string, number>()
    for (const t of clienteTurnos) {
      map.set(t.tipoServicio, (map.get(t.tipoServicio) ?? 0) + 1)
    }
    return Array.from(map.entries())
      .map(([name, value]) => ({ name, value }))
      .sort((a, b) => b.value - a.value)
  }, [clienteTurnos])

  const gastoPorVisita = useMemo(() => {
    return [...clienteTurnos]
      .sort((a, b) => new Date(a.fechaHora).getTime() - new Date(b.fechaHora).getTime())
      .map((t, i) => ({
        turno: `#${i + 1}`,
        fecha: formatFecha(t.fechaHora),
        monto: t.precioServicio ?? 0,
      }))
  }, [clienteTurnos])

  return (
    <div className="fixed inset-0 z-50 bg-background overflow-y-auto animate-in fade-in-0 duration-200">
      <div className="max-w-5xl mx-auto p-4 sm:p-6 space-y-6 page-enter">
        <div className="flex flex-col sm:flex-row sm:items-center gap-4">
          <Button variant="secondary" onClick={onVolver} className="self-start">
            <ArrowLeft className="w-4 h-4 mr-2" />
            Volver a Turnos
          </Button>
          <div className="flex-1 min-w-0">
            <h1 className="text-2xl font-bold text-foreground truncate">
              {nombre}
            </h1>
            <div className="flex flex-wrap items-center gap-3 mt-1 text-sm text-muted-foreground">
              {telefono && (
                <a href={`tel:${telefono}`} className="flex items-center gap-1 hover:text-foreground transition-colors">
                  <Phone className="w-3.5 h-3.5" />
                  {telefono}
                </a>
              )}
              {email && (
                <a href={`mailto:${email}`} className="flex items-center gap-1 hover:text-foreground transition-colors">
                  <Mail className="w-3.5 h-3.5" />
                  {email}
                </a>
              )}
              {primerTurno && (
                <span className="flex items-center gap-1">
                  <CalendarIcon className="w-3.5 h-3.5" />
                  Primera visita: {formatFecha(primerTurno.fechaHora)}
                </span>
              )}
              {ultimoTurno && ultimoTurno !== primerTurno && (
                <span className="flex items-center gap-1">
                  <CalendarIcon className="w-3.5 h-3.5" />
                  Última visita: {formatFecha(ultimoTurno.fechaHora)}
                </span>
              )}
            </div>
          </div>
        </div>

        <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-5 gap-4">
          <KpiCard label="Total turnos" value={String(clienteTurnos.length)} tone="blue" animated />
          <KpiCard label="Total invertido" value={formatCurrency(totalInvertido)} tone="green" animated />
          <KpiCard label="Ticket promedio" value={formatCurrency(ticketPromedio)} tone="ocre" animated />
          <KpiCard label="Servicios distintos" value={String(serviciosDistintos)} tone="cyan" animated />
          <KpiCard
            label="Próxima reserva"
            value={proximaReserva ? formatFecha(proximaReserva.fechaHora) : "—"}
            tone="yellow"
          />
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div className="bg-card rounded-2xl border border-border p-5">
            <h2 className="text-lg font-semibold text-foreground mb-4">
              Servicios utilizados
            </h2>
            {distribucionServicios.length === 0 ? (
              <p className="text-sm text-muted-foreground">Sin datos.</p>
            ) : (
              <div className="flex flex-col sm:flex-row items-center gap-4">
                <div className="h-48 w-full max-w-[200px]">
                  <ResponsiveContainer width="100%" height="100%">
                    <PieChart>
                      <Pie
                        data={distribucionServicios}
                        cx="50%"
                        cy="50%"
                        innerRadius={40}
                        outerRadius={70}
                        paddingAngle={3}
                        dataKey="value"
                        stroke="#2c2c2c"
                      >
                        {distribucionServicios.map((_, i) => (
                          <Cell key={i} fill={PALETA[i % PALETA.length]} />
                        ))}
                      </Pie>
                      <Tooltip
                        formatter={(value) => [`${value ?? 0} turnos`, null]}
                        contentStyle={{
                          backgroundColor: "#2c2c2c",
                          border: "1px solid rgba(255,102,0,0.3)",
                          borderRadius: 12,
                          color: "#fff",
                        }}
                        labelStyle={{ color: "#999" }}
                      />
                    </PieChart>
                  </ResponsiveContainer>
                </div>
                <ul className="flex-1 w-full space-y-2">
                  {distribucionServicios.map((s, i) => (
                    <li key={s.name} className="flex items-center gap-2">
                      <span
                        className="w-3 h-3 rounded-full shrink-0"
                        style={{ backgroundColor: PALETA[i % PALETA.length] }}
                      />
                      <span className="flex-1 text-sm text-foreground font-medium">{s.name}</span>
                      <span className="text-sm text-muted-foreground">{s.value}</span>
                    </li>
                  ))}
                </ul>
              </div>
            )}
          </div>

          <div className="bg-card rounded-2xl border border-border p-5">
            <h2 className="text-lg font-semibold text-foreground mb-4">
              Gasto por visita
            </h2>
            {gastoPorVisita.length === 0 ? (
              <p className="text-sm text-muted-foreground">Sin datos.</p>
            ) : (
              <div className="h-52">
                <ResponsiveContainer width="100%" height="100%">
                  <BarChart data={gastoPorVisita} margin={{ top: 4, right: 8, left: 0, bottom: 0 }}>
                    <CartesianGrid strokeDasharray="3 3" stroke="rgba(255,255,255,0.08)" />
                    <XAxis
                      dataKey="turno"
                      stroke="#999999"
                      tick={{ fill: "#999999", fontSize: 11 }}
                      tickLine={false}
                      axisLine={false}
                    />
                    <YAxis
                      tickFormatter={(v: number) => formatCurrency(v)}
                      stroke="#999999"
                      tick={{ fill: "#999999", fontSize: 11 }}
                      tickLine={false}
                      axisLine={false}
                      width={65}
                    />
                    <Tooltip
                      formatter={(value) => [formatCurrency(Number(value ?? 0)), "Monto"]}
                      labelFormatter={(label, payload) => {
                        const item = payload?.[0]?.payload as { fecha?: string } | undefined
                        return item?.fecha ?? String(label)
                      }}
                      contentStyle={{
                        backgroundColor: "#2c2c2c",
                        border: "1px solid rgba(255,102,0,0.3)",
                        borderRadius: 12,
                        color: "#fff",
                      }}
                      labelStyle={{ color: "#999" }}
                    />
                    <Bar dataKey="monto" radius={[6, 6, 0, 0]}>
                      {gastoPorVisita.map((_, i) => (
                        <Cell key={i} fill={PALETA[i % PALETA.length]} />
                      ))}
                    </Bar>
                  </BarChart>
                </ResponsiveContainer>
              </div>
            )}
          </div>
        </div>

        <div className="bg-card rounded-2xl border border-border overflow-hidden">
          <div className="p-4 border-b border-border">
            <h2 className="font-semibold text-foreground">
              Historial de turnos
              <StatusBadge variant="info">{clienteTurnos.length}</StatusBadge>
            </h2>
          </div>
          <div className="overflow-x-auto hidden sm:block">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-border text-left">
                  <th className="p-3 text-muted-foreground font-medium">Fecha</th>
                  <th className="p-3 text-muted-foreground font-medium">Hora</th>
                  <th className="p-3 text-muted-foreground font-medium">Servicio</th>
                  <th className="p-3 text-muted-foreground font-medium">Estado</th>
                  <th className="p-3 text-muted-foreground font-medium">Precio</th>
                  <th className="p-3 text-muted-foreground font-medium">Acciones</th>
                </tr>
              </thead>
              <tbody>
                {clienteTurnos.map((t) => {
                  const estado = estadoDe(t)
                  const badge =
                    estado === "hoy"
                      ? { variant: "info" as const, label: "Hoy" }
                      : estado === "pendiente"
                        ? { variant: "warning" as const, label: "Pendiente" }
                        : { variant: "success" as const, label: "Confirmado" }
                  return (
                    <tr key={t.id} className="border-b border-border last:border-0 hover:bg-accent/50 transition-colors">
                      <td className="p-3 text-muted-foreground">{formatFecha(t.fechaHora)}</td>
                      <td className="p-3 text-muted-foreground">{formatHora(t.fechaHora)}</td>
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
                            onClick={() => onEditar(t)}
                            className="min-h-10 min-w-10 p-2.5 rounded-lg text-muted-foreground hover:text-foreground hover:bg-accent transition-colors active:scale-95"
                            aria-label="Editar turno"
                          >
                            <Pencil className="w-4 h-4" />
                          </button>
                          {estado === "hoy" && (
                            <button
                              onClick={() => onMover(t)}
                              className="min-h-10 min-w-10 p-2.5 rounded-lg text-muted-foreground hover:text-primary hover:bg-accent transition-colors active:scale-95"
                              aria-label="Mover turno"
                              title="Mover a otra fecha"
                            >
                              <ArrowRight className="w-4 h-4" />
                            </button>
                          )}
                          <button
                            onClick={() => onEliminar(t.id)}
                            className="min-h-10 min-w-10 p-2.5 rounded-lg text-muted-foreground hover:text-destructive hover:bg-destructive/10 transition-colors active:scale-95"
                            aria-label="Eliminar turno"
                          >
                            <Trash2 className="w-4 h-4" />
                          </button>
                        </div>
                      </td>
                    </tr>
                  )
                })}
              </tbody>
            </table>
          </div>
          <div className="sm:hidden divide-y divide-border">
            {clienteTurnos.map((t) => {
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
                    <p className="text-foreground font-medium">{t.tipoServicio}</p>
                    <StatusBadge variant={badge.variant}>{badge.label}</StatusBadge>
                  </div>
                  <div className="flex items-center gap-3 text-xs text-muted-foreground">
                    <span>{formatFecha(t.fechaHora)}</span>
                    <span>{formatHora(t.fechaHora)}</span>
                  </div>
                  <div className="flex items-center justify-between">
                    <div className="text-xs text-muted-foreground">
                      {formatCurrency(t.precioServicio)}
                    </div>
                    <div className="flex items-center gap-1">
                      <button
                        onClick={() => onEditar(t)}
                        className="min-h-10 min-w-10 p-2.5 rounded-lg text-muted-foreground hover:text-foreground hover:bg-accent transition-colors active:scale-95"
                        aria-label="Editar turno"
                      >
                        <Pencil className="w-4 h-4" />
                      </button>
                      {estado === "hoy" && (
                        <button
                          onClick={() => onMover(t)}
                          className="min-h-10 min-w-10 p-2.5 rounded-lg text-muted-foreground hover:text-primary hover:bg-accent transition-colors active:scale-95"
                          aria-label="Mover turno"
                          title="Mover a otra fecha"
                        >
                          <ArrowRight className="w-4 h-4" />
                        </button>
                      )}
                      <button
                        onClick={() => onEliminar(t.id)}
                        className="min-h-10 min-w-10 p-2.5 rounded-lg text-muted-foreground hover:text-destructive hover:bg-destructive/10 transition-colors active:scale-95"
                        aria-label="Eliminar turno"
                      >
                        <Trash2 className="w-4 h-4" />
                      </button>
                    </div>
                  </div>
                </div>
              )
            })}
          </div>
        </div>
      </div>
    </div>
  )
}
