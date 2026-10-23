import { memo, useCallback, useMemo } from "react"
import { useTranslation } from "react-i18next"
import { useDashboard } from "@/api/stats"
import { useLatestAppointments } from "@/api/appointments"
import { StatusBadge } from "@/components/ui/StatusBadge"
import { formatCurrency, formatDate, formatTime, isToday } from "@/lib/format"
import { useRegisterShortcuts } from "@/hooks/useShortcutsRegistry"
import { useHotkeys } from "@/hooks/useHotkeys"

const KpiCard = memo(function KpiCard({ label, value, hint }: { label: string; value: string; hint?: string }) {
  return (
    <div className="bg-card rounded-2xl p-5 border border-border">
      <p className="text-xs text-muted-foreground font-medium">{label}</p>
      <p className="text-2xl font-bold text-foreground mt-1">{value}</p>
      {hint && <p className="text-xs text-muted-foreground mt-1">{hint}</p>}
    </div>
  )
})

function SectionTitle({ children }: { children: React.ReactNode }) {
  return <h2 className="text-lg font-semibold text-foreground mb-3">{children}</h2>
}

export default function Dashboard() {
  const { t } = useTranslation("admin")
  const { t: ts } = useTranslation("shortcuts")
  const { data: dashboard, isLoading, isError, refetch: refetchDashboard } = useDashboard()
  const { data: latest, isLoading: loadingLatest, isError: latestFailed, refetch: refetchLatest } =
    useLatestAppointments(10)

  const refetchAll = useCallback(() => {
    void refetchDashboard()
    void refetchLatest()
  }, [refetchDashboard, refetchLatest])

  const shortcuts = useMemo(() => [{ key: "Shift+r", description: ts("admin.reload") }], [ts])
  useRegisterShortcuts("admin", shortcuts)
  useHotkeys(useMemo(() => ({ "Shift+r": refetchAll }), [refetchAll]), true, { ignoreWhenDialogOpen: true })

  const maxServiceCount = useMemo(
    () => Math.max(0, ...(dashboard?.services ?? []).map((s) => s.count)),
    [dashboard],
  )

  if (isLoading) {
    return (
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        {Array.from({ length: 4 }).map((_, i) => (
          <div key={i} className="bg-card rounded-2xl h-24 animate-pulse" />
        ))}
      </div>
    )
  }

  if (isError || !dashboard) {
    return (
      <div role="alert" className="bg-destructive/10 text-destructive text-sm p-4 rounded-lg">
        {t("dashboard.loadError")}
      </div>
    )
  }

  return (
    <div className="space-y-8 page-enter">
      <div>
        <h1 className="text-2xl font-bold text-foreground">{t("dashboard.title")}</h1>
        <p className="text-sm text-muted-foreground mt-1">{t("dashboard.subtitle")}</p>
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <KpiCard
          label={t("dashboard.kpi.today")}
          value={String(dashboard.appointmentsToday)}
          hint={t("dashboard.kpi.yesterday", { count: dashboard.appointmentsYesterday })}
        />
        <KpiCard
          label={t("dashboard.kpi.revenue")}
          value={formatCurrency(dashboard.revenueThisMonth)}
          hint={t("dashboard.kpi.lastMonth", { amount: formatCurrency(dashboard.revenueLastMonth) })}
        />
        <KpiCard label={t("dashboard.kpi.customers")} value={String(dashboard.customerCount)} />
        <KpiCard label={t("dashboard.kpi.recorded")} value={String(latest?.length ?? 0)} />
      </div>

      <div>
        <SectionTitle>{t("dashboard.topServices")}</SectionTitle>
        {dashboard.services.length ? (
          <div className="space-y-3">
            {dashboard.services.map((s) => (
              <div key={s.id} className="bg-card rounded-xl p-4 border border-border">
                <div className="flex justify-between items-center mb-2">
                  <span className="text-sm font-medium text-foreground">{s.name}</span>
                  <span className="text-sm text-primary font-bold">{s.count}</span>
                </div>
                <div className="h-2 rounded-full bg-background overflow-hidden">
                  <div
                    className="h-full rounded-full bg-primary transition-all duration-500"
                    style={{ width: `${maxServiceCount > 0 ? (s.count / maxServiceCount) * 100 : 0}%` }}
                  />
                </div>
              </div>
            ))}
          </div>
        ) : (
          <p className="text-sm text-muted-foreground">{t("dashboard.noServices")}</p>
        )}
      </div>

      <div>
        <SectionTitle>{t("dashboard.latest")}</SectionTitle>
        {loadingLatest ? (
          <div className="bg-card rounded-2xl h-32 animate-pulse" />
        ) : latestFailed ? (
          <div role="alert" className="bg-destructive/10 text-destructive text-sm p-4 rounded-lg">
            {t("dashboard.latestError")}
          </div>
        ) : !latest?.length ? (
          <p className="text-sm text-muted-foreground">{t("dashboard.noAppointments")}</p>
        ) : (
          <div className="bg-card rounded-2xl border border-border overflow-hidden">
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-border text-left">
                    {(["customer", "service", "date", "time", "price", "status"] as const).map((col) => (
                      <th key={col} className="p-3 text-muted-foreground font-medium">
                        {t(`dashboard.columns.${col}`)}
                      </th>
                    ))}
                  </tr>
                </thead>
                <tbody>
                  {latest.map((a) => (
                    <tr key={a.id} className="border-b border-border last:border-0">
                      <td className="p-3 text-foreground">
                        {a.customerFirstName} {a.customerLastName}
                      </td>
                      <td className="p-3 text-muted-foreground">{a.serviceName}</td>
                      <td className="p-3 text-muted-foreground">{formatDate(a.startsAt)}</td>
                      <td className="p-3 text-muted-foreground">{formatTime(a.startsAt)}</td>
                      <td className="p-3 text-foreground font-medium">{formatCurrency(a.servicePrice)}</td>
                      <td className="p-3">
                        <StatusBadge variant={isToday(a.startsAt) ? "info" : "success"}>
                          {isToday(a.startsAt) ? t("status.today") : t("status.scheduled")}
                        </StatusBadge>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
