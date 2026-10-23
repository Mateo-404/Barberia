import { useMemo } from "react"
import { useTranslation } from "react-i18next"
import { ArrowLeft, Phone, Mail, Calendar as CalendarIcon } from "lucide-react"
import {
  PieChart, Pie, Cell, Tooltip, ResponsiveContainer,
  BarChart, Bar, XAxis, YAxis, CartesianGrid,
} from "recharts"
import { KpiCard } from "@/components/admin/KpiCard"
import { StatusBadge } from "@/components/ui/StatusBadge"
import { Button } from "@/components/ui/button"
import { AppointmentActions } from "@/components/admin/AppointmentActions"
import { formatCurrency, formatDate, formatTime } from "@/lib/format"
import { appointmentStatus, STATUS_BADGE_VARIANT } from "@/lib/appointment-status"
import type { Appointment } from "@/api/appointments"

const PALETTE = ["var(--chart-1)", "var(--chart-2)", "var(--chart-3)", "var(--chart-4)", "var(--chart-5)"]
const TOOLTIP_STYLE = {
  backgroundColor: "var(--card)",
  border: "1px solid var(--border)",
  borderRadius: 12,
  color: "var(--foreground)",
}
const AXIS_TICK = { fill: "var(--muted-foreground)", fontSize: 11 }

/** Full-screen history and spending breakdown of one customer. */
export function CustomerDetail({
  phone,
  name,
  appointments,
  onBack,
  onEdit,
  onDelete,
  onMove,
}: {
  phone: string
  name: string
  appointments: Appointment[]
  onBack: () => void
  onEdit: (a: Appointment) => void
  onDelete: (id: number) => void
  onMove: (a: Appointment) => void
}) {
  const { t } = useTranslation("admin")

  // Most recent first.
  const history = useMemo(
    () => appointments.filter((a) => a.customerPhone === phone).sort((a, b) => b.startsAt.localeCompare(a.startsAt)),
    [appointments, phone],
  )
  const chronological = useMemo(() => [...history].reverse(), [history])

  const email = history[0]?.customerEmail ?? ""
  const first = chronological[0]
  const last = history[0]
  const nextBooking = history.find((a) => appointmentStatus(a.startsAt) !== "past")
  const totalSpent = history.reduce((sum, a) => sum + a.servicePrice, 0)
  const average = history.length > 0 ? Math.round(totalSpent / history.length) : 0
  const distinctServices = new Set(history.map((a) => a.serviceName)).size

  const serviceMix = useMemo(() => {
    const counts = new Map<string, number>()
    for (const a of history) {
      const key = a.serviceName ?? ""
      counts.set(key, (counts.get(key) ?? 0) + 1)
    }
    return [...counts.entries()].map(([label, value]) => ({ name: label, value })).sort((a, b) => b.value - a.value)
  }, [history])

  const spendPerVisit = useMemo(
    () => chronological.map((a, i) => ({ visit: `#${i + 1}`, date: formatDate(a.startsAt), amount: a.servicePrice })),
    [chronological],
  )

  const rowActions = (a: Appointment) => (
    <AppointmentActions
      appointment={a}
      onEdit={onEdit}
      onMove={appointmentStatus(a.startsAt) === "today" ? onMove : undefined}
      onDelete={onDelete}
    />
  )

  return (
    <div className="fixed inset-0 z-50 bg-background overflow-y-auto animate-in fade-in-0 duration-200">
      <div className="max-w-5xl mx-auto p-4 sm:p-6 space-y-6 page-enter">
        <div className="flex flex-col sm:flex-row sm:items-center gap-4">
          <Button variant="secondary" onClick={onBack} className="self-start">
            <ArrowLeft className="w-4 h-4 mr-2" />
            {t("customer.back")}
          </Button>
          <div className="flex-1 min-w-0">
            <h1 className="text-2xl font-bold text-foreground truncate">{name}</h1>
            <div className="flex flex-wrap items-center gap-3 mt-1 text-sm text-muted-foreground">
              {phone && (
                <a href={`tel:${phone}`} className="flex items-center gap-1 hover:text-foreground transition-colors">
                  <Phone className="w-3.5 h-3.5" />
                  {phone}
                </a>
              )}
              {email && (
                <a href={`mailto:${email}`} className="flex items-center gap-1 hover:text-foreground transition-colors">
                  <Mail className="w-3.5 h-3.5" />
                  {email}
                </a>
              )}
              {first && (
                <span className="flex items-center gap-1">
                  <CalendarIcon className="w-3.5 h-3.5" />
                  {t("customer.firstVisit", { date: formatDate(first.startsAt) })}
                </span>
              )}
              {last && last !== first && (
                <span className="flex items-center gap-1">
                  <CalendarIcon className="w-3.5 h-3.5" />
                  {t("customer.lastVisit", { date: formatDate(last.startsAt) })}
                </span>
              )}
            </div>
          </div>
        </div>

        <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-5 gap-4">
          <KpiCard label={t("customer.kpi.total")} value={String(history.length)} tone="blue" animated />
          <KpiCard label={t("customer.kpi.spent")} value={formatCurrency(totalSpent)} tone="green" animated />
          <KpiCard label={t("customer.kpi.average")} value={formatCurrency(average)} tone="ocre" animated />
          <KpiCard label={t("customer.kpi.services")} value={String(distinctServices)} tone="cyan" animated />
          <KpiCard
            label={t("customer.kpi.next")}
            value={nextBooking ? formatDate(nextBooking.startsAt) : "—"}
            tone="yellow"
          />
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div className="bg-card rounded-2xl border border-border p-5">
            <h2 className="text-lg font-semibold text-foreground mb-4">{t("customer.servicesUsed")}</h2>
            {serviceMix.length === 0 ? (
              <p className="text-sm text-muted-foreground">{t("customer.noData")}</p>
            ) : (
              <div className="flex flex-col sm:flex-row items-center gap-4">
                <div className="h-48 w-full max-w-[200px]">
                  <ResponsiveContainer width="100%" height="100%">
                    <PieChart>
                      <Pie
                        data={serviceMix}
                        cx="50%"
                        cy="50%"
                        innerRadius={40}
                        outerRadius={70}
                        paddingAngle={3}
                        dataKey="value"
                        stroke="var(--card)"
                      >
                        {serviceMix.map((_, i) => (
                          <Cell key={i} fill={PALETTE[i % PALETTE.length]} />
                        ))}
                      </Pie>
                      <Tooltip
                        formatter={(value) => [t("customer.appointmentsCount", { count: Number(value ?? 0) }), null]}
                        contentStyle={TOOLTIP_STYLE}
                        labelStyle={{ color: "var(--muted-foreground)" }}
                      />
                    </PieChart>
                  </ResponsiveContainer>
                </div>
                <ul className="flex-1 w-full space-y-2">
                  {serviceMix.map((s, i) => (
                    <li key={s.name} className="flex items-center gap-2">
                      <span className="w-3 h-3 rounded-full shrink-0" style={{ backgroundColor: PALETTE[i % PALETTE.length] }} />
                      <span className="flex-1 text-sm text-foreground font-medium">{s.name}</span>
                      <span className="text-sm text-muted-foreground">{s.value}</span>
                    </li>
                  ))}
                </ul>
              </div>
            )}
          </div>

          <div className="bg-card rounded-2xl border border-border p-5">
            <h2 className="text-lg font-semibold text-foreground mb-4">{t("customer.spendPerVisit")}</h2>
            {spendPerVisit.length === 0 ? (
              <p className="text-sm text-muted-foreground">{t("customer.noData")}</p>
            ) : (
              <div className="h-52">
                <ResponsiveContainer width="100%" height="100%">
                  <BarChart data={spendPerVisit} margin={{ top: 4, right: 8, left: 0, bottom: 0 }}>
                    <CartesianGrid strokeDasharray="3 3" stroke="var(--border)" />
                    <XAxis dataKey="visit" stroke="var(--muted-foreground)" tick={AXIS_TICK} tickLine={false} axisLine={false} />
                    <YAxis
                      tickFormatter={(v: number) => formatCurrency(v)}
                      stroke="var(--muted-foreground)"
                      tick={AXIS_TICK}
                      tickLine={false}
                      axisLine={false}
                      width={65}
                    />
                    <Tooltip
                      formatter={(value) => [formatCurrency(Number(value ?? 0)), t("customer.amount")]}
                      labelFormatter={(label, payload) => {
                        const item = payload?.[0]?.payload as { date?: string } | undefined
                        return item?.date ?? String(label)
                      }}
                      contentStyle={TOOLTIP_STYLE}
                      labelStyle={{ color: "var(--muted-foreground)" }}
                    />
                    <Bar dataKey="amount" radius={[6, 6, 0, 0]}>
                      {spendPerVisit.map((_, i) => (
                        <Cell key={i} fill={PALETTE[i % PALETTE.length]} />
                      ))}
                    </Bar>
                  </BarChart>
                </ResponsiveContainer>
              </div>
            )}
          </div>
        </div>

        <div className="bg-card rounded-2xl border border-border overflow-hidden">
          <div className="p-4 border-b border-border flex items-center gap-2">
            <h2 className="font-semibold text-foreground">{t("customer.history")}</h2>
            <StatusBadge variant="info">{history.length}</StatusBadge>
          </div>
          <div className="overflow-x-auto hidden sm:block">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-border text-left">
                  {(["date", "time", "service", "status", "price", "actions"] as const).map((col) => (
                    <th key={col} className="p-3 text-muted-foreground font-medium">{t(`appointments.columns.${col}`)}</th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {history.map((a) => {
                  const status = appointmentStatus(a.startsAt)
                  return (
                    <tr key={a.id} className="border-b border-border last:border-0 hover:bg-accent/50 transition-colors">
                      <td className="p-3 text-muted-foreground">{formatDate(a.startsAt)}</td>
                      <td className="p-3 text-muted-foreground">{formatTime(a.startsAt)}</td>
                      <td className="p-3 text-muted-foreground">{a.serviceName}</td>
                      <td className="p-3"><StatusBadge variant={STATUS_BADGE_VARIANT[status]}>{t(`status.${status}`)}</StatusBadge></td>
                      <td className="p-3 text-foreground font-medium">{formatCurrency(a.servicePrice)}</td>
                      <td className="p-3">{rowActions(a)}</td>
                    </tr>
                  )
                })}
              </tbody>
            </table>
          </div>
          <div className="sm:hidden divide-y divide-border">
            {history.map((a) => {
              const status = appointmentStatus(a.startsAt)
              return (
                <div key={a.id} className="p-4 space-y-2">
                  <div className="flex items-center justify-between">
                    <p className="text-foreground font-medium">{a.serviceName}</p>
                    <StatusBadge variant={STATUS_BADGE_VARIANT[status]}>{t(`status.${status}`)}</StatusBadge>
                  </div>
                  <div className="flex items-center gap-3 text-xs text-muted-foreground">
                    <span>{formatDate(a.startsAt)}</span>
                    <span>{formatTime(a.startsAt)}</span>
                  </div>
                  <div className="flex items-center justify-between">
                    <div className="text-xs text-muted-foreground">{formatCurrency(a.servicePrice)}</div>
                    {rowActions(a)}
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
