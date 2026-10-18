import { memo, useCallback, useMemo } from "react"
import { usePanelEstadisticas, useUltimosTurnos } from "@/api/estadisticas"
import { StatusBadge } from "@/components/ui/StatusBadge"
import { formatCurrency, formatFecha, formatHora, isHoy } from "@/lib/format"
import { useRegisterShortcuts } from "@/hooks/useShortcutsRegistry"
import { useHotkeys } from "@/hooks/useHotkeys"

const KpiCard = memo(function KpiCard({
  label,
  value,
  hint,
}: {
  label: string
  value: string
  hint?: string
}) {
  return (
    <div className="bg-card rounded-2xl p-5 border border-border">
      <p className="text-xs uppercase tracking-wide text-muted-foreground font-medium">
        {label}
      </p>
      <p className="text-2xl font-bold text-foreground mt-1">{value}</p>
      {hint && <p className="text-xs text-muted-foreground mt-1">{hint}</p>}
    </div>
  )
})

function SectionTitle({ children }: { children: React.ReactNode }) {
  return (
    <h2 className="text-lg font-semibold text-foreground mb-3">{children}</h2>
  )
}

export default function AdminDashboard() {
  const { data: panel, isLoading: loadingPanel, isError: errorPanel, refetch: refetchPanel } =
    usePanelEstadisticas()
  const { data: turnos, isLoading: loadingTurnos, isError: errorTurnos, refetch: refetchTurnos } =
    useUltimosTurnos(10)

  const refetchAll = useCallback(() => {
    refetchPanel()
    refetchTurnos()
  }, [refetchPanel, refetchTurnos])

  const shortcuts = useMemo(
    () => [
      { key: "Shift+r", description: "Recargar datos del panel" },
    ],
    [],
  )
  useRegisterShortcuts("admin", shortcuts)

  useHotkeys(
    useMemo(() => ({ "Shift+r": refetchAll }), [refetchAll]),
    true,
    { ignoreWhenDialogOpen: true },
  )

  const turnosOrdenados = useMemo(() => {
    if (!turnos) return []
    return [...turnos].sort(
      (a, b) =>
        new Date(b.fechaHora).getTime() - new Date(a.fechaHora).getTime(),
    )
  }, [turnos])

  const maxServicio = useMemo(() => {
    if (!panel?.servicios?.length) return 0
    return Math.max(...panel.servicios.map((s) => s.cantidadRealizado ?? 0))
  }, [panel])

  if (loadingPanel) {
    return (
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        {Array.from({ length: 4 }).map((_, i) => (
          <div key={i} className="bg-card rounded-2xl h-24 animate-pulse" />
        ))}
      </div>
    )
  }

  if (errorPanel) {
    return (
      <div className="bg-destructive/10 text-destructive text-sm p-4 rounded-lg">
        No se pudieron cargar las estadísticas. Intentá de nuevo más tarde.
      </div>
    )
  }

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
          value={String(panel?.turnosHoy ?? 0)}
          hint={`Ayer: ${panel?.turnosAyer ?? 0}`}
        />
        <KpiCard
          label="Ingresos del mes"
          value={formatCurrency(panel?.ingresosMes ?? 0)}
          hint={`Mes anterior: ${formatCurrency(panel?.ingresosMesAnterior ?? 0)}`}
        />
        <KpiCard
          label="Total clientes"
          value={String(panel?.cantClientes ?? 0)}
        />
        <KpiCard
          label="Turnos registrados"
          value={String(turnos?.length ?? 0)}
        />
      </div>

      <div>
        <SectionTitle>Servicios más solicitados</SectionTitle>
        {panel?.servicios?.length ? (
          <div className="space-y-3">
            {panel.servicios.map((s) => {
              const cantidad = s.cantidadRealizado ?? 0
              const pct = maxServicio > 0 ? (cantidad / maxServicio) * 100 : 0
              return (
                <div key={s.id} className="bg-card rounded-xl p-4 border border-border">
                  <div className="flex justify-between items-center mb-2">
                    <span className="text-sm font-medium text-foreground">
                      {s.nombre}
                    </span>
                    <span className="text-sm text-primary font-bold">
                      {cantidad}
                    </span>
                  </div>
                  <div className="h-2 rounded-full bg-background overflow-hidden">
                    <div
                      className="h-full rounded-full bg-primary transition-all duration-500"
                      style={{ width: `${pct}%` }}
                    />
                  </div>
                </div>
              )
            })}
          </div>
        ) : (
          <p className="text-sm text-muted-foreground">
            Aún no hay servicios registrados este mes.
          </p>
        )}
      </div>

      <div>
        <SectionTitle>Últimos turnos</SectionTitle>
        {loadingTurnos ? (
          <div className="bg-card rounded-2xl h-32 animate-pulse" />
        ) : errorTurnos ? (
          <div className="bg-destructive/10 text-destructive text-sm p-4 rounded-lg">
            No se pudieron cargar los turnos.
          </div>
        ) : turnosOrdenados.length === 0 ? (
          <p className="text-sm text-muted-foreground">
            No hay turnos registrados todavía.
          </p>
        ) : (
          <div className="bg-card rounded-2xl border border-border overflow-hidden">
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-border text-left">
                    <th className="p-3 text-muted-foreground font-medium">Cliente</th>
                    <th className="p-3 text-muted-foreground font-medium">Servicio</th>
                    <th className="p-3 text-muted-foreground font-medium">Fecha</th>
                    <th className="p-3 text-muted-foreground font-medium">Hora</th>
                    <th className="p-3 text-muted-foreground font-medium">Precio</th>
                    <th className="p-3 text-muted-foreground font-medium">Estado</th>
                  </tr>
                </thead>
                <tbody>
                  {turnosOrdenados.map((t) => {
                    const esHoy = isHoy(t.fechaHora)
                    return (
                      <tr
                        key={t.id}
                        className="border-b border-border last:border-0"
                      >
                        <td className="p-3 text-foreground">
                          {t.nombreCliente} {t.apellidoCliente}
                        </td>
                        <td className="p-3 text-muted-foreground">
                          {t.tipoServicio}
                        </td>
                        <td className="p-3 text-muted-foreground">
                          {formatFecha(t.fechaHora)}
                        </td>
                        <td className="p-3 text-muted-foreground">
                          {formatHora(t.fechaHora)}
                        </td>
                        <td className="p-3 text-foreground font-medium">
                          {formatCurrency(t.precioServicio)}
                        </td>
                        <td className="p-3">
                          <StatusBadge variant={esHoy ? "info" : "success"}>
                            {esHoy ? "Hoy" : "Agendado"}
                          </StatusBadge>
                        </td>
                      </tr>
                    )
                  })}
                </tbody>
              </table>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
