import { useMemo } from "react"
import type { ReactNode } from "react"
import {
  Area, AreaChart, Bar, BarChart, CartesianGrid, Cell, Pie, PieChart,
  ResponsiveContainer, Tooltip, XAxis, YAxis,
} from "recharts"
import { useQueryClient } from "@tanstack/react-query"
import { usePanelEstadisticas } from "@/api/estadisticas"
import { useTodosTurnos } from "@/api/turnos"
import { KpiCard } from "@/components/admin/KpiCard"
import { useHotkeys } from "@/hooks/useHotkeys"
import { useRegisterShortcuts } from "@/hooks/useShortcutsRegistry"
import { formatCurrency, formatFecha } from "@/lib/format"

const COLORES_SERVICIO = ["#ff6600", "#ffc107", "#ef4444", "#ff9770", "#eab308", "#dc2626"]

const TOOLTIP_STYLE = {
  backgroundColor: "#2c2c2c",
  border: "1px solid rgba(255,102,0,0.3)",
  borderRadius: 12,
  color: "#fff",
}

const DIAS_SEMANA = [
  { getDay: 1, label: "Lun" },
  { getDay: 2, label: "Mar" },
  { getDay: 3, label: "Mié" },
  { getDay: 4, label: "Jue" },
  { getDay: 5, label: "Vie" },
  { getDay: 6, label: "Sáb" },
  { getDay: 0, label: "Dom" },
]

function pctCambio(actual: number, anterior: number): string {
  if (anterior <= 0) return actual > 0 ? "+100%" : "0%"
  const delta = ((actual - anterior) / anterior) * 100
  const signo = delta >= 0 ? "+" : "-"
  return `${signo}${Math.abs(Math.round(delta))}%`
}

function ChartCard({
  title,
  subtitle,
  children,
}: {
  title: string
  subtitle?: string
  children: ReactNode
}) {
  return (
    <div className="bg-card rounded-2xl border border-border p-6">
      <h2 className="text-lg font-semibold text-foreground mb-0.5">{title}</h2>
      {subtitle ? (
        <p className="text-sm text-muted-foreground mb-5">{subtitle}</p>
      ) : (
        <div className="mb-5" />
      )}
      {children}
    </div>
  )
}

function BarraHorizontal({
  label,
  value,
  valor,
  sub,
  max,
  color,
}: {
  label: string
  value: number
  valor: string
  sub: string
  max: number
  color: string
}) {
  const pct = max > 0 ? Math.round((value / max) * 100) : 0
  return (
    <div>
      <div className="flex items-center justify-between gap-3 mb-1.5">
        <span className="text-sm font-medium text-foreground truncate min-w-0">{label}</span>
        <span className="text-sm font-bold text-foreground shrink-0">{valor}</span>
      </div>
      <div className="flex items-center gap-3">
        <div className="h-2.5 flex-1 rounded-full bg-[#2c2c2c] overflow-hidden">
          <div
            className="h-full rounded-full transition-all duration-500"
            style={{ width: `${pct}%`, backgroundColor: color }}
          />
        </div>
        <span className="text-xs text-muted-foreground shrink-0 w-28 text-right tabular-nums">
          {sub}
        </span>
      </div>
    </div>
  )
}

export default function EstadisticasAdmin() {
  const { data: panel, isLoading } = usePanelEstadisticas()
  const { data: todosTurnos } = useTodosTurnos()
  const queryClient = useQueryClient()

  useHotkeys({ r: () => queryClient.invalidateQueries() })
  useRegisterShortcuts("admin", [{ key: "r", description: "Refrescar datos" }])

  const ingresos30 = useMemo(() => {
    const series = panel?.ingresosDiarios ?? []
    const mapped = series.slice(-30).map((d) => ({
      ...d,
      monto: d.ingresoTotal ?? 0,
    }))
    const hoy = new Date().toDateString()
    mapped.sort((a, b) => {
      const da = new Date(a.fecha ?? "").toDateString() === hoy ? 1 : 0
      const db = new Date(b.fecha ?? "").toDateString() === hoy ? 1 : 0
      return da - db
    })
    return mapped
  }, [panel])

  const horarios = useMemo(() => {
    const map = new Map<number, number>()
    for (const h of panel?.horarios ?? []) {
      map.set(h.hora ?? 0, h.cantidadRealizado ?? 0)
    }
    return Array.from({ length: 24 }, (_, i) => ({ hora: i, cantidad: map.get(i) ?? 0 }))
  }, [panel])

  const horaPico = useMemo(
    () =>
      horarios.reduce(
        (max, h) => (h.cantidad > max.cantidad ? h : max),
        { hora: 0, cantidad: 0 },
      ),
    [horarios],
  )

  const hayHorarios = horarios.some((h) => h.cantidad > 0)

  const servicios = useMemo(() => {
    return (panel?.servicios ?? []).sort(
      (a, b) => (b.cantidadRealizado ?? 0) - (a.cantidadRealizado ?? 0),
    )
  }, [panel])

  const totalIngresos = useMemo(
    () => servicios.reduce((acc, s) => acc + (s.cantidadRealizado ?? 0), 0),
    [servicios],
  )

  const porDiaSemana = useMemo(() => {
    const acc = new Map<number, { turnos: number; ingreso: number }>()
    for (const t of todosTurnos ?? []) {
      const d = new Date(t.fechaHora).getDay()
      const e = acc.get(d) ?? { turnos: 0, ingreso: 0 }
      e.turnos += 1
      e.ingreso += t.precioServicio ?? 0
      acc.set(d, e)
    }
    return DIAS_SEMANA.map(({ getDay, label }) => {
      const e = acc.get(getDay) ?? { turnos: 0, ingreso: 0 }
      return { label, turnos: e.turnos, ingreso: Math.round(e.ingreso) }
    })
  }, [todosTurnos])

  const evolucionMensual = useMemo(() => {
    const acc = new Map<string, { fecha: string; mes: string; ingresos: number }>()
    for (const t of todosTurnos ?? []) {
      const d = new Date(t.fechaHora)
      if (isNaN(d.getTime())) continue
      const key = `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}`
      const e =
        acc.get(key) ??
        {
          fecha: `${key}-01`,
          mes: new Date(`${key}-01T00:00:00`).toLocaleDateString("es-AR", {
            month: "short",
            year: "2-digit",
          }),
          ingresos: 0,
        }
      e.ingresos += t.precioServicio ?? 0
      acc.set(key, e)
    }
    return [...acc.values()]
      .sort((a, b) => a.fecha.localeCompare(b.fecha))
      .map((d) => ({ ...d, ingresos: Math.round(d.ingresos) }))
  }, [todosTurnos])

  const ingresosPorServicio = useMemo(() => {
    const acc = new Map<string, { nombre: string; turnos: number; total: number }>()
    for (const t of todosTurnos ?? []) {
      const key = t.tipoServicio || "Sin servicio"
      const e = acc.get(key) ?? { nombre: key, turnos: 0, total: 0 }
      e.turnos += 1
      e.total += t.precioServicio ?? 0
      acc.set(key, e)
    }
    return [...acc.values()].sort((a, b) => b.total - a.total)
  }, [todosTurnos])

  const totalIngresosServicios = useMemo(
    () => ingresosPorServicio.reduce((acc, s) => acc + s.total, 0),
    [ingresosPorServicio],
  )

  const topClientes = useMemo(() => {
    const acc = new Map<string, { nombre: string; turnos: number; total: number }>()
    for (const t of todosTurnos ?? []) {
      const key = t.telefonoCliente || "Sin teléfono"
      const e =
        acc.get(key) ??
        {
          nombre: `${t.nombreCliente ?? ""} ${t.apellidoCliente ?? ""}`.trim() || "Cliente",
          turnos: 0,
          total: 0,
        }
      e.turnos += 1
      e.total += t.precioServicio ?? 0
      acc.set(key, e)
    }
    return [...acc.values()]
      .filter((c) => c.turnos >= 2)
      .sort((a, b) => b.turnos - a.turnos || b.total - a.total)
      .slice(0, 6)
  }, [todosTurnos])

  const ticketPromedio = useMemo(() => {
    const total = (todosTurnos ?? []).reduce((acc, t) => acc + (t.precioServicio ?? 0), 0)
    const n = todosTurnos?.length ?? 0
    return n > 0 ? Math.round(total / n) : 0
  }, [todosTurnos])

  const totalFacturado = useMemo(
    () => (todosTurnos ?? []).reduce((acc, t) => acc + (t.precioServicio ?? 0), 0),
    [todosTurnos],
  )

  if (isLoading) {
    return (
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        {Array.from({ length: 4 }).map((_, i) => (
          <div key={i} className="bg-card rounded-2xl h-28 animate-pulse" />
        ))}
      </div>
    )
  }

  const turnosHoy = panel?.turnosHoy ?? 0
  const turnosAyer = panel?.turnosAyer ?? 0
  const ingresosMes = panel?.ingresosMes ?? 0
  const ingresosMesAnterior = panel?.ingresosMesAnterior ?? 0
  const cantClientes = panel?.cantClientes ?? 0

  return (
    <div className="space-y-8 page-enter">
      <div>
        <h1 className="text-2xl font-bold text-foreground">Panel de Estadísticas</h1>
        <p className="text-sm text-muted-foreground mt-1">
          Resumen del negocio en tiempo real.
        </p>
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <KpiCard
          label="Turnos hoy"
          value={String(turnosHoy)}
          hint={`Vs ayer: ${pctCambio(turnosHoy, turnosAyer)}`}
          tone="yellow"
          animated
        />
        <KpiCard
          label="Ingresos del mes"
          value={formatCurrency(ingresosMes)}
          hint={`Vs mes anterior: ${pctCambio(ingresosMes, ingresosMesAnterior)}`}
          tone="yellow"
          animated
        />
        <KpiCard
          label="Total facturado"
          value={formatCurrency(totalFacturado)}
          hint={`${todosTurnos?.length ?? 0} turnos · Ticket: ${formatCurrency(ticketPromedio)}`}
          tone="green"
          animated
        />
        <KpiCard
          label="Clientes registrados"
          value={String(cantClientes)}
          tone="green"
          animated
        />
      </div>

      <ChartCard
        title="Ingresos Diarios"
        subtitle="Últimos 30 días"
      >
        {ingresos30.length === 0 ? (
          <p className="text-sm text-muted-foreground">
            Aún no hay datos de ingresos diarios.
          </p>
        ) : (
          <div className="h-80">
            <ResponsiveContainer width="100%" height="100%">
              <AreaChart data={ingresos30} margin={{ top: 8, right: 12, left: 0, bottom: 0 }}>
                <defs>
                  <linearGradient id="gradIngresos" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="0%" stopColor="#ff6600" stopOpacity={0.4} />
                    <stop offset="100%" stopColor="#ff6600" stopOpacity={0.02} />
                  </linearGradient>
                </defs>
                <CartesianGrid strokeDasharray="3 3" stroke="rgba(255,255,255,0.06)" />
                <XAxis
                  dataKey="fecha"
                  tickFormatter={(v: string) => formatFecha(v)}
                  interval={3}
                  stroke="#666"
                  tick={{ fill: "#888", fontSize: 11 }}
                  tickLine={false}
                  axisLine={false}
                />
                <YAxis
                  tickFormatter={(v: number) => formatCurrency(v)}
                  stroke="#666"
                  tick={{ fill: "#888", fontSize: 11 }}
                  tickLine={false}
                  axisLine={false}
                  domain={[0, "auto"]}
                  width={70}
                />
                <Tooltip
                  formatter={(value) => [formatCurrency(Number(value ?? 0)), "Ingresos"]}
                  labelFormatter={(label) => formatFecha(String(label))}
                  contentStyle={TOOLTIP_STYLE}
                  labelStyle={{ color: "#999" }}
                />
                <Area
                  type="monotone"
                  dataKey="monto"
                  name="Ingresos"
                  stroke="#ff6600"
                  strokeWidth={2.5}
                  fill="url(#gradIngresos)"
                  dot={false}
                  activeDot={{ r: 5, fill: "#ff6600", stroke: "#fff", strokeWidth: 2 }}
                />
              </AreaChart>
            </ResponsiveContainer>
          </div>
        )}
      </ChartCard>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <ChartCard
          title="Turnos por Hora"
          subtitle={
            hayHorarios
              ? `Horario pico: ${horaPico.hora}:00 hs · ${horaPico.cantidad} turnos`
              : "Último año · sin turnos registrados"
          }
        >
          {!hayHorarios ? (
            <p className="text-sm text-muted-foreground">
              Aún no hay datos horarios.
            </p>
          ) : (
            <div className="h-72">
              <ResponsiveContainer width="100%" height="100%">
                <BarChart data={horarios} margin={{ top: 8, right: 12, left: 0, bottom: 0 }}>
                  <CartesianGrid strokeDasharray="3 3" stroke="rgba(255,255,255,0.06)" />
                  <XAxis
                    dataKey="hora"
                    tickFormatter={(v: number) => `${v}h`}
                    interval={2}
                    stroke="#666"
                    tick={{ fill: "#888", fontSize: 11 }}
                    tickLine={false}
                    axisLine={false}
                  />
                  <YAxis
                    allowDecimals={false}
                    stroke="#666"
                    tick={{ fill: "#888", fontSize: 11 }}
                    tickLine={false}
                    axisLine={false}
                    width={30}
                  />
                  <Tooltip
                    formatter={(value) => [`${value ?? 0} turnos`, null]}
                    labelFormatter={(label) => `${String(label)}:00 hs`}
                    contentStyle={TOOLTIP_STYLE}
                    labelStyle={{ color: "#999" }}
                  />
                  <Bar dataKey="cantidad" name="Turnos" radius={[4, 4, 0, 0]} maxBarSize={18}>
                    {horarios.map((h) => (
                      <Cell
                        key={h.hora}
                        fill={
                          h.hora === horaPico.hora
                            ? "#dc3545"
                            : h.cantidad > 0
                              ? "#ff6600"
                              : "rgba(255,102,0,0.15)"
                        }
                      />
                    ))}
                  </Bar>
                </BarChart>
              </ResponsiveContainer>
            </div>
          )}
        </ChartCard>

        <ChartCard
          title="Turnos e Ingresos por Día"
          subtitle="Todos los turnos registrados"
        >
          {porDiaSemana.every((d) => d.turnos === 0) ? (
            <p className="text-sm text-muted-foreground">
              Aún no hay turnos registrados.
            </p>
          ) : (
            <div className="h-72">
              <ResponsiveContainer width="100%" height="100%">
                <BarChart data={porDiaSemana} margin={{ top: 8, right: 12, left: 0, bottom: 0 }}>
                  <CartesianGrid strokeDasharray="3 3" stroke="rgba(255,255,255,0.06)" />
                  <XAxis
                    dataKey="label"
                    stroke="#666"
                    tick={{ fill: "#888", fontSize: 11 }}
                    tickLine={false}
                    axisLine={false}
                  />
                  <YAxis
                    yAxisId="izq"
                    allowDecimals={false}
                    stroke="#666"
                    tick={{ fill: "#888", fontSize: 11 }}
                    tickLine={false}
                    axisLine={false}
                    width={30}
                  />
                  <YAxis
                    yAxisId="der"
                    orientation="right"
                    tickFormatter={(v: number) => formatCurrency(v)}
                    stroke="#666"
                    tick={{ fill: "#888", fontSize: 11 }}
                    tickLine={false}
                    axisLine={false}
                    width={64}
                  />
                  <Tooltip
                    formatter={(value, name) =>
                      name === "Ingresos"
                        ? [formatCurrency(Number(value ?? 0)), "Ingresos"]
                        : [`${value ?? 0} turnos`, "Turnos"]
                    }
                    contentStyle={TOOLTIP_STYLE}
                    labelStyle={{ color: "#999" }}
                  />
                  <Bar
                    yAxisId="izq"
                    dataKey="turnos"
                    name="Turnos"
                    fill="#ff6600"
                    radius={[4, 4, 0, 0]}
                    maxBarSize={20}
                  />
                  <Bar
                    yAxisId="der"
                    dataKey="ingreso"
                    name="Ingresos"
                    fill="#ffc107"
                    radius={[4, 4, 0, 0]}
                    maxBarSize={20}
                  />
                </BarChart>
              </ResponsiveContainer>
            </div>
          )}
        </ChartCard>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <ChartCard title="Servicios Más Populares" subtitle="Último mes">
          {servicios.length === 0 ? (
            <p className="text-sm text-muted-foreground">
              Aún no hay datos de servicios en el último mes.
            </p>
          ) : (
            <div className="flex flex-col md:flex-row items-center gap-8">
              <div className="h-72 w-full max-w-[280px]">
                <ResponsiveContainer width="100%" height="100%">
                  <PieChart>
                    <Pie
                      data={servicios.map((s) => ({
                        name: s.nombre,
                        value: s.cantidadRealizado ?? 0,
                      }))}
                      cx="50%"
                      cy="50%"
                      innerRadius={60}
                      outerRadius={95}
                      paddingAngle={3}
                      dataKey="value"
                      stroke="#2c2c2c"
                    >
                      {servicios.map((_, i) => (
                        <Cell key={i} fill={COLORES_SERVICIO[i % COLORES_SERVICIO.length]} />
                      ))}
                    </Pie>
                    <Tooltip
                      formatter={(value) => [`${value ?? 0} turnos`, null]}
                      contentStyle={TOOLTIP_STYLE}
                      labelStyle={{ color: "#999" }}
                    />
                  </PieChart>
                </ResponsiveContainer>
              </div>
              <ul className="flex-1 w-full space-y-3">
                {servicios.map((s, i) => {
                  const cantidad = s.cantidadRealizado ?? 0
                  const pct = totalIngresos > 0 ? Math.round((cantidad / totalIngresos) * 100) : 0
                  return (
                    <li key={s.id} className="flex items-center gap-3">
                      <span
                        className="w-3 h-3 rounded-full shrink-0"
                        style={{ backgroundColor: COLORES_SERVICIO[i % COLORES_SERVICIO.length] }}
                      />
                      <span className="flex-1 text-sm text-foreground font-medium">
                        {s.nombre}
                      </span>
                      <span className="text-sm text-muted-foreground tabular-nums">
                        {pct}% · {cantidad}
                      </span>
                    </li>
                  )
                })}
              </ul>
            </div>
          )}
        </ChartCard>

        <ChartCard
          title="Ingresos por Servicio"
          subtitle="Total facturado por servicio"
        >
          {ingresosPorServicio.length === 0 ? (
            <p className="text-sm text-muted-foreground">
              Aún no hay turnos registrados.
            </p>
          ) : (
            <ul className="space-y-5">
              {ingresosPorServicio.slice(0, 6).map((s, i) => (
                <li key={s.nombre}>
                  <BarraHorizontal
                    label={s.nombre}
                    value={s.total}
                    valor={formatCurrency(s.total)}
                    sub={`${s.turnos} turnos · ${
                      totalIngresosServicios > 0
                        ? Math.round((s.total / totalIngresosServicios) * 100)
                        : 0
                    }%`}
                    max={ingresosPorServicio[0].total}
                    color={COLORES_SERVICIO[i % COLORES_SERVICIO.length]}
                  />
                </li>
              ))}
            </ul>
          )}
        </ChartCard>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <ChartCard
          title="Clientes Frecuentes"
          subtitle="Top por cantidad de turnos (mínimo 2 visitas)"
        >
          {topClientes.length === 0 ? (
            <p className="text-sm text-muted-foreground">
              Aún no hay clientes habituales.
            </p>
          ) : (
            <ul className="space-y-5">
              {topClientes.map((c, i) => (
                <li key={`${c.nombre}-${i}`}>
                  <BarraHorizontal
                    label={c.nombre}
                    value={c.turnos}
                    valor={`${c.turnos} turnos`}
                    sub={formatCurrency(c.total)}
                    max={topClientes[0].turnos}
                    color={COLORES_SERVICIO[i % COLORES_SERVICIO.length]}
                  />
                </li>
              ))}
            </ul>
          )}
        </ChartCard>

        <ChartCard
          title="Evolución Mensual de Ingresos"
          subtitle="Histórico mes a mes"
        >
          {evolucionMensual.length === 0 ? (
            <p className="text-sm text-muted-foreground">
              Aún no hay turnos registrados.
            </p>
          ) : (
            <div className="h-72">
              <ResponsiveContainer width="100%" height="100%">
                <BarChart data={evolucionMensual} margin={{ top: 8, right: 12, left: 0, bottom: 0 }}>
                  <CartesianGrid strokeDasharray="3 3" stroke="rgba(255,255,255,0.06)" />
                  <defs>
                    <linearGradient id="gradMensual" x1="0" y1="0" x2="0" y2="1">
                      <stop offset="0%" stopColor="#ff6600" stopOpacity={1} />
                      <stop offset="100%" stopColor="#ff6600" stopOpacity={0.6} />
                    </linearGradient>
                  </defs>
                  <XAxis
                    dataKey="mes"
                    stroke="#666"
                    tick={{ fill: "#888", fontSize: 11 }}
                    tickLine={false}
                    axisLine={false}
                  />
                  <YAxis
                    tickFormatter={(v: number) => formatCurrency(v)}
                    stroke="#666"
                    tick={{ fill: "#888", fontSize: 11 }}
                    tickLine={false}
                    axisLine={false}
                    domain={[0, "auto"]}
                    width={70}
                  />
                  <Tooltip
                    formatter={(value) => [formatCurrency(Number(value ?? 0)), "Ingresos"]}
                    contentStyle={TOOLTIP_STYLE}
                    labelStyle={{ color: "#999" }}
                  />
                  <Bar
                    dataKey="ingresos"
                    name="Ingresos"
                    fill="url(#gradMensual)"
                    radius={[6, 6, 0, 0]}
                    maxBarSize={40}
                  />
                </BarChart>
              </ResponsiveContainer>
            </div>
          )}
        </ChartCard>
      </div>
    </div>
  )
}
