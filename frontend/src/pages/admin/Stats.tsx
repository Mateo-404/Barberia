import { useMemo } from "react"
import type { ReactNode } from "react"
import { useTranslation } from "react-i18next"
import {
  Area, AreaChart, Bar, BarChart, CartesianGrid, Cell, Pie, PieChart,
  ResponsiveContainer, Tooltip, XAxis, YAxis,
} from "recharts"
import { useQueryClient } from "@tanstack/react-query"
import { useDashboard } from "@/api/stats"
import { useAppointments } from "@/api/appointments"
import { KpiCard } from "@/components/admin/KpiCard"
import { useHotkeys } from "@/hooks/useHotkeys"
import { useRegisterShortcuts } from "@/hooks/useShortcutsRegistry"
import { errorMessage } from "@/lib/error-message"
import { formatCurrency, formatDate } from "@/lib/format"

const SERIES = ["var(--chart-1)", "var(--chart-2)", "var(--chart-3)", "var(--chart-4)", "var(--chart-5)"]
const GRID = "var(--border)"
const AXIS = { stroke: "var(--muted-foreground)", tick: { fill: "var(--muted-foreground)", fontSize: 11 }, tickLine: false, axisLine: false } as const
const TOOLTIP_STYLE = {
  backgroundColor: "var(--card)",
  border: "1px solid var(--border)",
  borderRadius: 12,
  color: "var(--foreground)",
}
const TOOLTIP_LABEL = { color: "var(--muted-foreground)" }

/** Monday-first weekday numbers as returned by Date#getDay(). */
const WEEK = [1, 2, 3, 4, 5, 6, 0] as const

function percentChange(current: number, previous: number): string {
  if (previous <= 0) return current > 0 ? "+100%" : "0%"
  const delta = ((current - previous) / previous) * 100
  return `${delta >= 0 ? "+" : "-"}${Math.abs(Math.round(delta))}%`
}

function ChartCard({ title, subtitle, children }: { title: string; subtitle?: string; children: ReactNode }) {
  return (
    <div className="bg-card rounded-2xl border border-border p-6">
      <h2 className="text-lg font-semibold text-foreground mb-0.5">{title}</h2>
      {subtitle ? <p className="text-sm text-muted-foreground mb-5">{subtitle}</p> : <div className="mb-5" />}
      {children}
    </div>
  )
}

function Empty({ children }: { children: ReactNode }) {
  return <p className="text-sm text-muted-foreground">{children}</p>
}

function HorizontalBar({
  label, value, display, sub, max, color,
}: { label: string; value: number; display: string; sub: string; max: number; color: string }) {
  const pct = max > 0 ? Math.round((value / max) * 100) : 0
  return (
    <div>
      <div className="flex items-center justify-between gap-3 mb-1.5">
        <span className="text-sm font-medium text-foreground truncate min-w-0">{label}</span>
        <span className="text-sm font-bold text-foreground shrink-0">{display}</span>
      </div>
      <div className="flex items-center gap-3">
        <div className="h-2.5 flex-1 rounded-full bg-muted overflow-hidden">
          <div className="h-full rounded-full transition-all duration-500" style={{ width: `${pct}%`, backgroundColor: color }} />
        </div>
        <span className="text-xs text-muted-foreground shrink-0 w-28 text-right tabular-nums">{sub}</span>
      </div>
    </div>
  )
}

interface Tally {
  name: string
  count: number
  total: number
}

export default function Stats() {
  const { t, i18n } = useTranslation("admin")
  const { t: tc } = useTranslation()
  const { t: ts } = useTranslation("shortcuts")
  const { data: dashboard, isLoading, isError, error, refetch } = useDashboard()
  const { data: appointments } = useAppointments()
  const queryClient = useQueryClient()

  useHotkeys({ r: () => queryClient.invalidateQueries() })
  useRegisterShortcuts("admin", [{ key: "r", description: ts("stats.refresh") }])

  // The API lists days newest first; charts read oldest → newest.
  const dailyRevenue = useMemo(() => [...(dashboard?.dailyRevenue ?? [])].reverse(), [dashboard])

  const hours = useMemo(() => {
    const byHour = new Map((dashboard?.hours ?? []).map((h) => [h.hour, h.count]))
    return Array.from({ length: 24 }, (_, hour) => ({ hour, count: byHour.get(hour) ?? 0 }))
  }, [dashboard])
  const peak = useMemo(() => hours.reduce((max, h) => (h.count > max.count ? h : max), { hour: 0, count: 0 }), [hours])
  const hasHours = hours.some((h) => h.count > 0)

  const popular = useMemo(() => [...(dashboard?.services ?? [])].sort((a, b) => b.count - a.count), [dashboard])
  const popularTotal = popular.reduce((sum, s) => sum + s.count, 0)

  const all = useMemo(() => appointments ?? [], [appointments])
  const totalRevenue = all.reduce((sum, a) => sum + a.servicePrice, 0)
  const averageTicket = all.length > 0 ? Math.round(totalRevenue / all.length) : 0

  const weekdayNames = useMemo(() => {
    const fmt = new Intl.DateTimeFormat(i18n.language, { weekday: "short" })
    return Object.fromEntries(WEEK.map((d) => [d, fmt.format(new Date(2024, 0, 7 + d))]))
  }, [i18n.language])

  const byWeekday = useMemo(() => {
    const acc = new Map<number, { count: number; revenue: number }>()
    for (const a of all) {
      const day = new Date(a.startsAt).getDay()
      const entry = acc.get(day) ?? { count: 0, revenue: 0 }
      acc.set(day, { count: entry.count + 1, revenue: entry.revenue + a.servicePrice })
    }
    return WEEK.map((day) => ({
      label: weekdayNames[day],
      count: acc.get(day)?.count ?? 0,
      revenue: Math.round(acc.get(day)?.revenue ?? 0),
    }))
  }, [all, weekdayNames])

  const monthly = useMemo(() => {
    const acc = new Map<string, number>()
    for (const a of all) {
      const key = a.startsAt.slice(0, 7)
      acc.set(key, (acc.get(key) ?? 0) + a.servicePrice)
    }
    const fmt = new Intl.DateTimeFormat(i18n.language, { month: "short", year: "2-digit" })
    return [...acc.entries()]
      .sort(([a], [b]) => a.localeCompare(b))
      .map(([key, revenue]) => ({ month: fmt.format(new Date(`${key}-01T00:00:00`)), revenue: Math.round(revenue) }))
  }, [all, i18n.language])

  const tally = (keyOf: (a: (typeof all)[number]) => string, nameOf: (a: (typeof all)[number]) => string) => {
    const acc = new Map<string, Tally>()
    for (const a of all) {
      const key = keyOf(a)
      const entry = acc.get(key) ?? { name: nameOf(a), count: 0, total: 0 }
      acc.set(key, { ...entry, count: entry.count + 1, total: entry.total + a.servicePrice })
    }
    return [...acc.values()]
  }

  const revenueByService = useMemo(
    () => tally((a) => a.serviceName ?? "", (a) => a.serviceName || t("stats.noService")).sort((a, b) => b.total - a.total),
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [all, t],
  )
  const revenueByServiceTotal = revenueByService.reduce((sum, s) => sum + s.total, 0)

  const topCustomers = useMemo(
    () =>
      tally(
        (a) => a.customerPhone ?? "",
        (a) => `${a.customerFirstName ?? ""} ${a.customerLastName ?? ""}`.trim() || t("stats.customer"),
      )
        .filter((c) => c.count >= 2)
        .sort((a, b) => b.count - a.count || b.total - a.total)
        .slice(0, 6),
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [all, t],
  )

  if (isError) {
    return (
      <div className="space-y-8 page-enter">
        <h1 className="text-2xl font-bold text-foreground">{t("stats.title")}</h1>
        <div role="alert" className="bg-card rounded-2xl border border-destructive/40 p-6 space-y-3">
          <h2 className="text-lg font-semibold text-foreground">{t("stats.loadError")}</h2>
          <p className="text-sm text-muted-foreground">{errorMessage(tc, error)}</p>
          <button
            type="button"
            onClick={() => void refetch()}
            className="rounded-xl bg-primary px-4 py-2 text-sm font-semibold text-primary-foreground hover:opacity-90"
          >
            {tc("retry")}
          </button>
        </div>
      </div>
    )
  }

  if (isLoading || !dashboard) {
    return (
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        {Array.from({ length: 4 }).map((_, i) => (
          <div key={i} className="bg-card rounded-2xl h-28 animate-pulse" />
        ))}
      </div>
    )
  }

  const color = (i: number) => SERIES[i % SERIES.length]
  const appointmentsLabel = (count: number) => t("stats.appointmentsCount", { count })

  return (
    <div className="space-y-8 page-enter">
      <div>
        <h1 className="text-2xl font-bold text-foreground">{t("stats.title")}</h1>
        <p className="text-sm text-muted-foreground mt-1">{t("stats.subtitle")}</p>
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <KpiCard
          label={t("dashboard.kpi.today")}
          value={String(dashboard.appointmentsToday)}
          hint={t("stats.vsYesterday", { change: percentChange(dashboard.appointmentsToday, dashboard.appointmentsYesterday) })}
          tone="yellow"
          animated
        />
        <KpiCard
          label={t("dashboard.kpi.revenue")}
          value={formatCurrency(dashboard.revenueThisMonth)}
          hint={t("stats.vsLastMonth", { change: percentChange(dashboard.revenueThisMonth, dashboard.revenueLastMonth) })}
          tone="yellow"
          animated
        />
        <KpiCard
          label={t("stats.kpi.billed")}
          value={formatCurrency(totalRevenue)}
          hint={t("stats.kpi.billedHint", { count: all.length, ticket: formatCurrency(averageTicket) })}
          tone="green"
          animated
        />
        <KpiCard label={t("stats.kpi.customers")} value={String(dashboard.customerCount)} tone="green" animated />
      </div>

      <ChartCard title={t("stats.daily.title")} subtitle={t("stats.daily.subtitle")}>
        {dailyRevenue.length === 0 ? (
          <Empty>{t("stats.daily.empty")}</Empty>
        ) : (
          <div className="h-80">
            <ResponsiveContainer width="100%" height="100%">
              <AreaChart data={dailyRevenue} margin={{ top: 8, right: 12, left: 0, bottom: 0 }}>
                <defs>
                  <linearGradient id="dailyFill" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="0%" stopColor="var(--primary)" stopOpacity={0.4} />
                    <stop offset="100%" stopColor="var(--primary)" stopOpacity={0.02} />
                  </linearGradient>
                </defs>
                <CartesianGrid strokeDasharray="3 3" stroke={GRID} />
                <XAxis dataKey="date" tickFormatter={(v: string) => formatDate(v)} interval={3} {...AXIS} />
                <YAxis tickFormatter={(v: number) => formatCurrency(v)} domain={[0, "auto"]} width={70} {...AXIS} />
                <Tooltip
                  formatter={(value) => [formatCurrency(Number(value ?? 0)), t("stats.revenue")]}
                  labelFormatter={(label) => formatDate(String(label))}
                  contentStyle={TOOLTIP_STYLE}
                  labelStyle={TOOLTIP_LABEL}
                />
                <Area
                  type="monotone"
                  dataKey="total"
                  name={t("stats.revenue")}
                  stroke="var(--primary)"
                  strokeWidth={2.5}
                  fill="url(#dailyFill)"
                  dot={false}
                  activeDot={{ r: 5, fill: "var(--primary)", stroke: "var(--background)", strokeWidth: 2 }}
                />
              </AreaChart>
            </ResponsiveContainer>
          </div>
        )}
      </ChartCard>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <ChartCard
          title={t("stats.hours.title")}
          subtitle={hasHours ? t("stats.hours.peak", { hour: peak.hour, count: peak.count }) : t("stats.hours.none")}
        >
          {!hasHours ? (
            <Empty>{t("stats.hours.empty")}</Empty>
          ) : (
            <div className="h-72">
              <ResponsiveContainer width="100%" height="100%">
                <BarChart data={hours} margin={{ top: 8, right: 12, left: 0, bottom: 0 }}>
                  <CartesianGrid strokeDasharray="3 3" stroke={GRID} />
                  <XAxis dataKey="hour" tickFormatter={(v: number) => `${v}h`} interval={2} {...AXIS} />
                  <YAxis allowDecimals={false} width={30} {...AXIS} />
                  <Tooltip
                    formatter={(value) => [appointmentsLabel(Number(value ?? 0)), null]}
                    labelFormatter={(label) => `${String(label)}:00`}
                    contentStyle={TOOLTIP_STYLE}
                    labelStyle={TOOLTIP_LABEL}
                  />
                  <Bar dataKey="count" radius={[4, 4, 0, 0]} maxBarSize={18}>
                    {hours.map((h) => (
                      <Cell
                        key={h.hour}
                        fill={h.hour === peak.hour ? "var(--chart-4)" : h.count > 0 ? "var(--primary)" : "var(--accent)"}
                      />
                    ))}
                  </Bar>
                </BarChart>
              </ResponsiveContainer>
            </div>
          )}
        </ChartCard>

        <ChartCard title={t("stats.weekday.title")} subtitle={t("stats.weekday.subtitle")}>
          {byWeekday.every((d) => d.count === 0) ? (
            <Empty>{t("stats.noAppointments")}</Empty>
          ) : (
            <div className="h-72">
              <ResponsiveContainer width="100%" height="100%">
                <BarChart data={byWeekday} margin={{ top: 8, right: 12, left: 0, bottom: 0 }}>
                  <CartesianGrid strokeDasharray="3 3" stroke={GRID} />
                  <XAxis dataKey="label" {...AXIS} />
                  <YAxis yAxisId="count" allowDecimals={false} width={30} {...AXIS} />
                  <YAxis yAxisId="revenue" orientation="right" tickFormatter={(v: number) => formatCurrency(v)} width={64} {...AXIS} />
                  <Tooltip
                    formatter={(value, name) =>
                      name === "revenue"
                        ? [formatCurrency(Number(value ?? 0)), t("stats.revenue")]
                        : [appointmentsLabel(Number(value ?? 0)), t("stats.weekday.appointments")]
                    }
                    contentStyle={TOOLTIP_STYLE}
                    labelStyle={TOOLTIP_LABEL}
                  />
                  <Bar yAxisId="count" dataKey="count" name="count" fill="var(--chart-1)" radius={[4, 4, 0, 0]} maxBarSize={20} />
                  <Bar yAxisId="revenue" dataKey="revenue" name="revenue" fill="var(--chart-3)" radius={[4, 4, 0, 0]} maxBarSize={20} />
                </BarChart>
              </ResponsiveContainer>
            </div>
          )}
        </ChartCard>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <ChartCard title={t("stats.popular.title")} subtitle={t("stats.popular.subtitle")}>
          {popular.length === 0 ? (
            <Empty>{t("stats.popular.empty")}</Empty>
          ) : (
            <div className="flex flex-col md:flex-row items-center gap-8">
              <div className="h-72 w-full max-w-[280px]">
                <ResponsiveContainer width="100%" height="100%">
                  <PieChart>
                    <Pie
                      data={popular.map((s) => ({ name: s.name, value: s.count }))}
                      cx="50%"
                      cy="50%"
                      innerRadius={60}
                      outerRadius={95}
                      paddingAngle={3}
                      dataKey="value"
                      stroke="var(--card)"
                    >
                      {popular.map((_, i) => (
                        <Cell key={i} fill={color(i)} />
                      ))}
                    </Pie>
                    <Tooltip
                      formatter={(value) => [appointmentsLabel(Number(value ?? 0)), null]}
                      contentStyle={TOOLTIP_STYLE}
                      labelStyle={TOOLTIP_LABEL}
                    />
                  </PieChart>
                </ResponsiveContainer>
              </div>
              <ul className="flex-1 w-full space-y-3">
                {popular.map((s, i) => (
                  <li key={s.id} className="flex items-center gap-3">
                    <span className="w-3 h-3 rounded-full shrink-0" style={{ backgroundColor: color(i) }} />
                    <span className="flex-1 text-sm text-foreground font-medium">{s.name}</span>
                    <span className="text-sm text-muted-foreground tabular-nums">
                      {popularTotal > 0 ? Math.round((s.count / popularTotal) * 100) : 0}% · {s.count}
                    </span>
                  </li>
                ))}
              </ul>
            </div>
          )}
        </ChartCard>

        <ChartCard title={t("stats.byService.title")} subtitle={t("stats.byService.subtitle")}>
          {revenueByService.length === 0 ? (
            <Empty>{t("stats.noAppointments")}</Empty>
          ) : (
            <ul className="space-y-5">
              {revenueByService.slice(0, 6).map((s, i) => (
                <li key={s.name}>
                  <HorizontalBar
                    label={s.name}
                    value={s.total}
                    display={formatCurrency(s.total)}
                    sub={`${appointmentsLabel(s.count)} · ${revenueByServiceTotal > 0 ? Math.round((s.total / revenueByServiceTotal) * 100) : 0}%`}
                    max={revenueByService[0].total}
                    color={color(i)}
                  />
                </li>
              ))}
            </ul>
          )}
        </ChartCard>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <ChartCard title={t("stats.regulars.title")} subtitle={t("stats.regulars.subtitle")}>
          {topCustomers.length === 0 ? (
            <Empty>{t("stats.regulars.empty")}</Empty>
          ) : (
            <ul className="space-y-5">
              {topCustomers.map((c, i) => (
                <li key={`${c.name}-${i}`}>
                  <HorizontalBar
                    label={c.name}
                    value={c.count}
                    display={appointmentsLabel(c.count)}
                    sub={formatCurrency(c.total)}
                    max={topCustomers[0].count}
                    color={color(i)}
                  />
                </li>
              ))}
            </ul>
          )}
        </ChartCard>

        <ChartCard title={t("stats.monthly.title")} subtitle={t("stats.monthly.subtitle")}>
          {monthly.length === 0 ? (
            <Empty>{t("stats.noAppointments")}</Empty>
          ) : (
            <div className="h-72">
              <ResponsiveContainer width="100%" height="100%">
                <BarChart data={monthly} margin={{ top: 8, right: 12, left: 0, bottom: 0 }}>
                  <CartesianGrid strokeDasharray="3 3" stroke={GRID} />
                  <XAxis dataKey="month" {...AXIS} />
                  <YAxis tickFormatter={(v: number) => formatCurrency(v)} domain={[0, "auto"]} width={70} {...AXIS} />
                  <Tooltip
                    formatter={(value) => [formatCurrency(Number(value ?? 0)), t("stats.revenue")]}
                    contentStyle={TOOLTIP_STYLE}
                    labelStyle={TOOLTIP_LABEL}
                  />
                  <Bar dataKey="revenue" name={t("stats.revenue")} fill="var(--primary)" radius={[6, 6, 0, 0]} maxBarSize={40} />
                </BarChart>
              </ResponsiveContainer>
            </div>
          )}
        </ChartCard>
      </div>
    </div>
  )
}
