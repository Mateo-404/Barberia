import { useCallback, useMemo, useRef, useState } from "react"
import {
  Search, List, Calendar, CalendarDays, RotateCcw, ChevronLeft, ChevronRight, GripVertical,
} from "lucide-react"
import { useTranslation } from "react-i18next"
import { useAppointments, useUpdateAppointment, useDeleteAppointment, type Appointment } from "@/api/appointments"
import { useServices } from "@/api/services"
import { StatusBadge } from "@/components/ui/StatusBadge"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Dialog } from "@/components/ui/dialog"
import { KpiCard } from "@/components/admin/KpiCard"
import { CustomerDetail } from "@/components/admin/CustomerDetail"
import { AppointmentActions } from "@/components/admin/AppointmentActions"
import { formatCurrency, formatDate, formatTime, isToday } from "@/lib/format"
import { timeSlots } from "@/lib/business-config"
import { isFutureDateTime, todayInBusinessTz } from "@/lib/datetime"
import { appointmentStatus, STATUS_BADGE_VARIANT, type AppointmentStatus } from "@/lib/appointment-status"
import { errorMessage } from "@/lib/error-message"
import { useHotkeys } from "@/hooks/useHotkeys"
import { useRegisterShortcuts } from "@/hooks/useShortcutsRegistry"

type View = "list" | "calendar"
type DayCell = { day: number; iso: string }

const customerName = (a: Appointment) => `${a.customerFirstName ?? ""} ${a.customerLastName ?? ""}`.trim()

function DateFilter({
  value, onChange, label,
}: { value: string; onChange: (v: string) => void; label: string }) {
  return (
    <div className="relative">
      <CalendarDays className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
      <Input
        type="date"
        value={value}
        onChange={(e) => onChange(e.target.value)}
        aria-label={label}
        className="w-[9.5rem] pl-9 pr-2 text-sm [color-scheme:dark]"
      />
    </div>
  )
}

function StatusLabel({ startsAt }: { startsAt: string }) {
  const { t } = useTranslation("admin")
  const status = appointmentStatus(startsAt)
  return <StatusBadge variant={STATUS_BADGE_VARIANT[status]}>{t(`status.${status}`)}</StatusBadge>
}

export default function Appointments() {
  const { t, i18n } = useTranslation("admin")
  const { t: tc } = useTranslation()
  const { t: ts } = useTranslation("shortcuts")
  const { data: appointments, isLoading } = useAppointments()
  const { data: services } = useServices()
  const update = useUpdateAppointment()
  const remove = useDeleteAppointment()

  const [view, setView] = useState<View>("list")
  const [search, setSearch] = useState("")
  const [statusFilter, setStatusFilter] = useState<AppointmentStatus | "all">("all")
  const [serviceFilter, setServiceFilter] = useState<"all" | number>("all")
  const [dateFrom, setDateFrom] = useState("")
  const [dateTo, setDateTo] = useState("")
  const [editing, setEditing] = useState<Appointment | null>(null)
  const [deletingId, setDeletingId] = useState<number | null>(null)
  const [selectedDay, setSelectedDay] = useState<string | null>(null)
  const [month, setMonth] = useState(() => {
    const now = new Date()
    return new Date(now.getFullYear(), now.getMonth(), 1)
  })
  const [customer, setCustomer] = useState<{ phone: string; name: string } | null>(null)

  const [draggingId, setDraggingId] = useState<number | null>(null)
  const [dragOverDay, setDragOverDay] = useState<string | null>(null)
  const [movingId, setMovingId] = useState<number | null>(null)
  const [moveDate, setMoveDate] = useState(todayInBusinessTz())

  const searchRef = useRef<HTMLInputElement>(null)
  const hasModal = editing !== null || deletingId !== null || movingId !== null || customer !== null
  const today = todayInBusinessTz()

  const sorted = useMemo(
    () => [...(appointments ?? [])].sort((a, b) => a.startsAt.localeCompare(b.startsAt)),
    [appointments],
  )

  const kpis = useMemo(() => {
    const upcoming = sorted.filter((a) => appointmentStatus(a.startsAt) === "upcoming").length
    return {
      total: sorted.length,
      past: sorted.length - upcoming,
      upcoming,
      revenueToday: sorted.filter((a) => isToday(a.startsAt)).reduce((sum, a) => sum + a.servicePrice, 0),
    }
  }, [sorted])

  const filtered = useMemo(() => {
    const q = search.trim().toLowerCase()
    return sorted.filter((a) => {
      if (statusFilter !== "all" && appointmentStatus(a.startsAt) !== statusFilter) return false
      if (serviceFilter !== "all" && a.serviceId !== serviceFilter) return false
      if (q && !`${customerName(a)} ${a.customerPhone ?? ""}`.toLowerCase().includes(q)) return false
      const day = a.startsAt.slice(0, 10)
      return !(dateFrom && day < dateFrom) && !(dateTo && day > dateTo)
    })
  }, [sorted, search, statusFilter, serviceFilter, dateFrom, dateTo])

  const byDay = useMemo(() => {
    const map = new Map<string, Appointment[]>()
    for (const a of sorted) {
      const key = a.startsAt.slice(0, 10)
      map.set(key, [...(map.get(key) ?? []), a])
    }
    return map
  }, [sorted])

  const dayAppointments = selectedDay ? (byDay.get(selectedDay) ?? []) : []

  function clearFilters() {
    setSearch("")
    setStatusFilter("all")
    setServiceFilter("all")
    setDateFrom("")
    setDateTo("")
    setSelectedDay(null)
  }

  function openCustomer(a: Appointment) {
    setSelectedDay(null)
    setCustomer({ phone: a.customerPhone ?? "", name: customerName(a) })
  }

  function startMove(a: Appointment) {
    setMovingId(a.id)
    setMoveDate(a.startsAt.slice(0, 10))
  }

  async function saveEdit(startsAt: string) {
    if (!editing) return
    try {
      await update.mutateAsync({ id: editing.id, patch: { startsAt } })
      setEditing(null)
    } catch {
      // The error shows inside the dialog (update.isError).
    }
  }

  async function confirmDelete(id: number) {
    try {
      await remove.mutateAsync(id)
      setDeletingId(null)
    } catch {
      // The error shows inside the dialog (remove.isError).
    }
  }

  async function moveToDay(id: number, day: string) {
    const appointment = sorted.find((a) => a.id === id)
    if (!appointment) return
    try {
      await update.mutateAsync({ id, patch: { startsAt: `${day}T${appointment.startsAt.slice(11, 16)}:00` } })
      setDraggingId(null)
      setDragOverDay(null)
      setMovingId(null)
      setSelectedDay(day)
    } catch {
      // The error shows inside the dialog (update.isError).
    }
  }

  const shiftMonth = useCallback(
    (delta: number) => setMonth((m) => new Date(m.getFullYear(), m.getMonth() + delta, 1)),
    [],
  )
  const goToToday = useCallback(() => setMonth(new Date()), [])

  const monthCells = useMemo(() => {
    const year = month.getFullYear()
    const index = month.getMonth()
    const offset = (new Date(year, index, 1).getDay() + 6) % 7
    const cells: Array<DayCell | null> = Array.from({ length: offset }, () => null)
    for (let d = 1; d <= new Date(year, index + 1, 0).getDate(); d++) {
      cells.push({ day: d, iso: `${year}-${String(index + 1).padStart(2, "0")}-${String(d).padStart(2, "0")}` })
    }
    return cells
  }, [month])
  const monthDays = useMemo(() => monthCells.filter((c): c is DayCell => c !== null), [monthCells])

  const monthLabel = useMemo(() => {
    const label = new Intl.DateTimeFormat(i18n.language, { month: "long", year: "numeric" }).format(month)
    return label.charAt(0).toUpperCase() + label.slice(1)
  }, [month, i18n.language])

  // Monday-first narrow weekday names for the active locale.
  const weekdays = useMemo(() => {
    const fmt = new Intl.DateTimeFormat(i18n.language, { weekday: "narrow" })
    return Array.from({ length: 7 }, (_, i) => fmt.format(new Date(2024, 0, 1 + i)))
  }, [i18n.language])

  function stepSelectedDay(delta: number) {
    if (view !== "calendar" || monthDays.length === 0) return
    const base = selectedDay ?? (delta > 0 ? "" : today)
    const index = monthDays.findIndex((c) => c.iso === base)
    const next = index < 0 ? (delta > 0 ? 0 : monthDays.length - 1) : index + delta
    if (next >= 0 && next < monthDays.length) setSelectedDay(monthDays[next].iso)
  }

  function onDrop(e: React.DragEvent, day: string) {
    e.preventDefault()
    const id = Number(e.dataTransfer.getData("text/plain"))
    const appointment = sorted.find((a) => a.id === id)
    if (!appointment) return
    if (appointment.startsAt.slice(0, 10) === day) {
      setDraggingId(null)
      setDragOverDay(null)
      return
    }
    if (day >= today) void moveToDay(id, day)
  }

  const hotkeyMap = useMemo(
    () => ({
      Escape: () => {
        if (customer) setCustomer(null)
        else if (editing) setEditing(null)
        else if (deletingId !== null) setDeletingId(null)
        else if (movingId !== null) setMovingId(null)
      },
      "/": () => searchRef.current?.focus(),
      l: () => setView("list"),
      c: () => setView("calendar"),
      ArrowLeft: () => view === "calendar" && shiftMonth(-1),
      ArrowRight: () => view === "calendar" && shiftMonth(1),
      "1": () => setStatusFilter("all"),
      "2": () => setStatusFilter("past"),
      "3": () => setStatusFilter("upcoming"),
      "4": () => setStatusFilter("today"),
      x: clearFilters,
      h: () => view === "calendar" && goToToday(),
      n: () => stepSelectedDay(1),
      p: () => stepSelectedDay(-1),
      j: () => stepSelectedDay(1),
      k: () => stepSelectedDay(-1),
      e: () => {
        if (view === "calendar" && dayAppointments.length > 0) setEditing(dayAppointments[0])
      },
      Delete: () => {
        const first = view === "calendar" && selectedDay ? dayAppointments[0] : filtered[0]
        if (first) setDeletingId(first.id)
      },
      Enter: () => {
        if (view === "list" && filtered.length > 0) openCustomer(filtered[0])
      },
    }),
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [customer, editing, deletingId, movingId, view, selectedDay, monthDays, dayAppointments, filtered],
  )
  useHotkeys(hotkeyMap, !hasModal)

  useRegisterShortcuts("admin", [
    { key: "l", description: ts("appointments.list") },
    { key: "c", description: ts("appointments.calendar") },
    { key: "/", description: ts("appointments.search") },
    { key: "Escape", description: ts("appointments.close") },
    { key: "1", description: ts("appointments.filterAll") },
    { key: "2", description: ts("appointments.filterPast") },
    { key: "3", description: ts("appointments.filterUpcoming") },
    { key: "4", description: ts("appointments.filterToday") },
    { key: "x", description: ts("appointments.clearFilters") },
    { key: "h", description: ts("appointments.goToday") },
    { key: "n", description: ts("appointments.nextDay") },
    { key: "p", description: ts("appointments.previousDay") },
    { key: "j", description: ts("appointments.selectNext") },
    { key: "k", description: ts("appointments.selectPrevious") },
    { key: "e", description: ts("appointments.editFirst") },
    { key: "Delete", description: ts("appointments.deleteFirst") },
    { key: "Enter", description: ts("appointments.openCustomer") },
  ])

  const select = "bg-card border-2 border-border text-foreground text-sm rounded-lg px-3 py-2.5 focus:outline-none focus:border-primary"
  const nameButton = "text-left hover:bg-accent rounded-lg px-2 py-1 -mx-2 -my-1 transition-colors active:scale-95"
  const tab = (active: boolean) =>
    `flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium transition-colors active:scale-95 ${
      active ? "bg-primary text-primary-foreground" : "text-muted-foreground hover:text-foreground"
    }`

  const customerCell = (a: Appointment, withPhone: boolean) => (
    <button onClick={() => openCustomer(a)} className={nameButton}>
      <p className="text-foreground font-medium">{customerName(a)}</p>
      {withPhone && <p className="text-xs text-muted-foreground">{a.customerPhone}</p>}
    </button>
  )

  const listEmpty = t("appointments.noMatches")

  return (
    <div className="space-y-6 page-enter">
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
        <div>
          <h1 className="text-2xl font-bold text-foreground">{t("appointments.title")}</h1>
          <p className="text-sm text-muted-foreground mt-1">{t("appointments.subtitle")}</p>
        </div>
        <div className="flex items-center gap-2">
          <div className="flex bg-card rounded-xl p-1 border border-border">
            <button onClick={() => setView("list")} className={tab(view === "list")}>
              <List className="w-4 h-4" />
              {t("appointments.viewList")}
            </button>
            <button onClick={() => setView("calendar")} className={tab(view === "calendar")}>
              <Calendar className="w-4 h-4" />
              {t("appointments.viewCalendar")}
            </button>
          </div>
          <span className="hidden sm:inline text-xs text-muted-foreground whitespace-nowrap">
            <kbd className="px-1.5 py-0.5 rounded bg-muted text-[10px] font-mono">Shift+?</kbd> {t("appointments.shortcuts")}
          </span>
        </div>
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <KpiCard label={t("appointments.kpi.total")} value={String(kpis.total)} tone="blue" animated />
        <KpiCard label={t("appointments.kpi.past")} value={String(kpis.past)} tone="green" animated />
        <KpiCard label={t("appointments.kpi.upcoming")} value={String(kpis.upcoming)} tone="ocre" animated />
        <KpiCard label={t("appointments.kpi.revenueToday")} value={formatCurrency(kpis.revenueToday)} tone="cyan" animated />
      </div>

      <div className="flex flex-col md:flex-row gap-3 md:items-center md:justify-between">
        <div className="relative flex-1 max-w-sm">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
          <Input
            ref={searchRef}
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder={t("appointments.searchPlaceholder")}
            className="pl-9"
          />
        </div>
        <div className="flex flex-wrap items-center gap-2">
          <select
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value as AppointmentStatus | "all")}
            className={select}
            aria-label={t("appointments.statusFilter")}
          >
            <option value="all">{t("appointments.allStatuses")}</option>
            <option value="past">{t("status.past")}</option>
            <option value="upcoming">{t("status.upcoming")}</option>
            <option value="today">{t("status.today")}</option>
          </select>
          <select
            value={serviceFilter}
            onChange={(e) => setServiceFilter(e.target.value === "all" ? "all" : Number(e.target.value))}
            className={select}
            aria-label={t("appointments.serviceFilter")}
          >
            <option value="all">{t("appointments.allServices")}</option>
            {services?.map((s) => (
              <option key={s.id} value={s.id}>{s.name}</option>
            ))}
          </select>
          <DateFilter value={dateFrom} onChange={setDateFrom} label={t("appointments.from")} />
          <DateFilter value={dateTo} onChange={setDateTo} label={t("appointments.to")} />
          <Button variant="secondary" size="sm" onClick={clearFilters}>
            <RotateCcw className="w-4 h-4 mr-1" />
            {t("appointments.clear")}
          </Button>
        </div>
      </div>

      <h2 className="text-lg font-semibold text-foreground flex items-center gap-2">
        {t("appointments.history")}
        <StatusBadge variant="info">{filtered.length}</StatusBadge>
      </h2>

      {isLoading ? (
        <div className="bg-card rounded-2xl h-48 animate-pulse border border-border" />
      ) : view === "list" ? (
        <div key="view-list" className="page-enter bg-card rounded-2xl border border-border overflow-hidden">
          <div className="overflow-x-auto hidden sm:block">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-border text-left">
                  {(["date", "time", "customer", "service", "status", "price", "actions"] as const).map((col) => (
                    <th key={col} className="p-3 text-muted-foreground font-medium">{t(`appointments.columns.${col}`)}</th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {filtered.length === 0 ? (
                  <tr>
                    <td colSpan={7} className="p-6 text-center text-muted-foreground">{listEmpty}</td>
                  </tr>
                ) : (
                  filtered.map((a) => (
                    <tr key={a.id} className="border-b border-border last:border-0 hover:bg-accent/50 transition-colors">
                      <td className="p-3 text-muted-foreground">{formatDate(a.startsAt)}</td>
                      <td className="p-3 text-muted-foreground">{formatTime(a.startsAt)}</td>
                      <td className="p-3">{customerCell(a, true)}</td>
                      <td className="p-3 text-muted-foreground">{a.serviceName}</td>
                      <td className="p-3"><StatusLabel startsAt={a.startsAt} /></td>
                      <td className="p-3 text-foreground font-medium">{formatCurrency(a.servicePrice)}</td>
                      <td className="p-3">
                        <AppointmentActions appointment={a} onEdit={setEditing} onDelete={setDeletingId} />
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
          <div className="sm:hidden divide-y divide-border">
            {filtered.length === 0 ? (
              <p className="p-6 text-center text-muted-foreground text-sm">{listEmpty}</p>
            ) : (
              filtered.map((a) => (
                <div key={a.id} className="p-4 space-y-2">
                  <div className="flex items-center justify-between">
                    {customerCell(a, false)}
                    <StatusLabel startsAt={a.startsAt} />
                  </div>
                  <div className="flex items-center gap-3 text-xs text-muted-foreground">
                    <span>{formatDate(a.startsAt)}</span>
                    <span>{formatTime(a.startsAt)}</span>
                  </div>
                  <div className="flex items-center justify-between">
                    <div className="text-xs text-muted-foreground">
                      {a.serviceName} · {formatCurrency(a.servicePrice)}
                    </div>
                    <AppointmentActions appointment={a} onEdit={setEditing} onDelete={setDeletingId} />
                  </div>
                </div>
              ))
            )}
          </div>
        </div>
      ) : (
        <div key="view-calendar" className="page-enter space-y-4">
          <div className="bg-card rounded-2xl border border-border p-4">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-lg font-semibold text-foreground">{monthLabel}</h3>
              <div className="flex gap-1">
                <button
                  onClick={() => shiftMonth(-1)}
                  className="min-h-10 min-w-10 p-2 rounded-lg text-muted-foreground hover:text-foreground hover:bg-accent active:scale-95 transition-all"
                  aria-label={t("appointments.previousMonth")}
                >
                  <ChevronLeft className="w-5 h-5" />
                </button>
                <button
                  onClick={goToToday}
                  className="px-3 py-2 rounded-lg text-sm font-medium text-muted-foreground hover:text-foreground hover:bg-accent active:scale-95 transition-all"
                >
                  {t("appointments.today")}
                </button>
                <button
                  onClick={() => shiftMonth(1)}
                  className="min-h-10 min-w-10 p-2 rounded-lg text-muted-foreground hover:text-foreground hover:bg-accent active:scale-95 transition-all"
                  aria-label={t("appointments.nextMonth")}
                >
                  <ChevronRight className="w-5 h-5" />
                </button>
              </div>
            </div>
            <div className="grid grid-cols-7 gap-1 text-center mb-1">
              {weekdays.map((d, i) => (
                <span key={i} className="text-xs text-muted-foreground font-medium py-1">{d}</span>
              ))}
            </div>
            <div key={`${month.getFullYear()}-${month.getMonth()}`} className="page-enter grid grid-cols-7 gap-1">
              {monthCells.map((cell, i) => {
                if (!cell) return <div key={`empty-${i}`} />
                const items = byDay.get(cell.iso) ?? []
                const isSelected = selectedDay === cell.iso
                const isTodayCell = cell.iso === today
                const isDropTarget = dragOverDay === cell.iso && draggingId !== null
                return (
                  <button
                    key={cell.iso}
                    onClick={() => setSelectedDay(cell.iso)}
                    onDragOver={(e) => {
                      e.preventDefault()
                      e.dataTransfer.dropEffect = "move"
                      setDragOverDay(cell.iso)
                    }}
                    onDragLeave={() => setDragOverDay(null)}
                    onDrop={(e) => onDrop(e, cell.iso)}
                    className={`relative min-h-20 sm:min-h-24 rounded-lg p-1 flex flex-col items-stretch gap-0.5 overflow-hidden transition-all duration-150 ${
                      isDropTarget
                        ? "ring-2 ring-primary ring-offset-1 ring-offset-background scale-105"
                        : isSelected
                          ? "bg-primary/10 ring-2 ring-primary"
                          : isTodayCell
                            ? "border border-primary"
                            : "hover:bg-accent"
                    }`}
                  >
                    <span className={`text-xs text-left pl-1 ${isSelected || isTodayCell ? "text-primary font-bold" : "text-muted-foreground"}`}>
                      {cell.day}
                    </span>
                    {items.slice(0, 3).map((a) => (
                      <span
                        key={a.id}
                        className={`text-[9px] sm:text-[10px] leading-tight px-1 py-0.5 rounded truncate text-left ${
                          isSelected ? "bg-primary text-primary-foreground" : "bg-primary/20 text-foreground"
                        }`}
                      >
                        {formatTime(a.startsAt)} {a.serviceName}
                      </span>
                    ))}
                    {items.length > 3 && (
                      <span className="text-[9px] sm:text-[10px] pl-1 text-muted-foreground text-left leading-tight">
                        {t("appointments.more", { count: items.length - 3 })}
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
                {selectedDay ? formatDate(`${selectedDay}T00:00:00`) : t("appointments.dayTitle")}
              </h3>
              <StatusBadge variant="info">{dayAppointments.length}</StatusBadge>
            </div>
            {dayAppointments.length === 0 ? (
              <div className="p-4 text-sm text-muted-foreground text-center">
                {draggingId !== null ? t("appointments.dropHere") : t("appointments.pickDay")}
              </div>
            ) : (
              <>
                <div className="overflow-x-auto hidden sm:block">
                  <table className="w-full text-sm">
                    <thead>
                      <tr className="border-b border-border text-left">
                        <th className="p-3 w-8"><span className="sr-only">{t("appointments.drag")}</span></th>
                        {(["time", "customer", "service", "price", "actions"] as const).map((col) => (
                          <th key={col} className="p-3 text-muted-foreground font-medium">{t(`appointments.columns.${col}`)}</th>
                        ))}
                      </tr>
                    </thead>
                    <tbody>
                      {dayAppointments.map((a) => (
                        <tr
                          key={a.id}
                          draggable
                          onDragStart={(e) => {
                            e.dataTransfer.effectAllowed = "move"
                            e.dataTransfer.setData("text/plain", String(a.id))
                            setDraggingId(a.id)
                          }}
                          onDragEnd={() => {
                            setDraggingId(null)
                            setDragOverDay(null)
                          }}
                          className={`border-b border-border last:border-0 transition-colors ${
                            draggingId === a.id ? "opacity-50" : "hover:bg-accent/50"
                          }`}
                        >
                          <td className="p-3 text-muted-foreground cursor-grab active:cursor-grabbing">
                            <GripVertical className="w-4 h-4" />
                          </td>
                          <td className="p-3 text-muted-foreground">{formatTime(a.startsAt)}</td>
                          <td className="p-3">{customerCell(a, true)}</td>
                          <td className="p-3 text-muted-foreground">{a.serviceName}</td>
                          <td className="p-3 text-foreground font-medium">{formatCurrency(a.servicePrice)}</td>
                          <td className="p-3">
                            <AppointmentActions appointment={a} onEdit={setEditing} onMove={startMove} />
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
                <div className="sm:hidden divide-y divide-border">
                  {dayAppointments.map((a) => (
                    <div key={a.id} className="p-4 space-y-2">
                      <div className="flex items-center justify-between">
                        {customerCell(a, false)}
                        <span className="text-xs text-muted-foreground">{formatTime(a.startsAt)}</span>
                      </div>
                      <div className="flex items-center justify-between">
                        <span className="text-xs text-muted-foreground">
                          {a.serviceName} · {formatCurrency(a.servicePrice)}
                        </span>
                        <AppointmentActions appointment={a} onEdit={setEditing} onMove={startMove} />
                      </div>
                    </div>
                  ))}
                </div>
              </>
            )}
          </div>
        </div>
      )}

      {customer && (
        <CustomerDetail
          phone={customer.phone}
          name={customer.name}
          appointments={sorted}
          onBack={() => setCustomer(null)}
          onEdit={(a) => { setCustomer(null); setEditing(a) }}
          onDelete={(id) => { setCustomer(null); setDeletingId(id) }}
          onMove={(a) => { setCustomer(null); startMove(a) }}
        />
      )}

      {editing && (
        <EditDialog
          appointment={editing}
          saving={update.isPending}
          error={update.isError ? errorMessage(tc, update.error) : null}
          onCancel={() => {
            update.reset()
            setEditing(null)
          }}
          onSave={saveEdit}
        />
      )}

      {deletingId !== null && (
        <Dialog
          open
          onClose={() => setDeletingId(null)}
          title={t("appointments.deleteTitle")}
          closeLabel={tc("close")}
          footer={
            <>
              <Button variant="secondary" onClick={() => setDeletingId(null)}>{tc("cancel")}</Button>
              <Button variant="destructive" onClick={() => void confirmDelete(deletingId)} disabled={remove.isPending}>
                {remove.isPending ? t("appointments.deleting") : tc("delete")}
              </Button>
            </>
          }
        >
          <p className="text-sm text-muted-foreground">{t("appointments.deleteWarning")}</p>
          {remove.isError && (
            <p role="alert" className="text-sm text-destructive">{errorMessage(tc, remove.error)}</p>
          )}
        </Dialog>
      )}

      {movingId !== null && (
        <Dialog
          open
          onClose={() => setMovingId(null)}
          title={t("appointments.moveTitle")}
          closeLabel={tc("close")}
          footer={
            <>
              <Button variant="secondary" onClick={() => setMovingId(null)}>{tc("cancel")}</Button>
              <Button
                onClick={() => moveDate >= today && void moveToDay(movingId, moveDate)}
                disabled={update.isPending || moveDate < today}
              >
                {update.isPending ? t("appointments.moving") : t("appointments.move")}
              </Button>
            </>
          }
        >
          <div className="space-y-1.5">
            <label htmlFor="move-date" className="text-sm text-foreground">{t("appointments.newDate")}</label>
            <Input
              id="move-date"
              type="date"
              value={moveDate}
              min={today}
              onChange={(e) => setMoveDate(e.target.value)}
              className="w-full [color-scheme:dark]"
            />
            {moveDate < today && <p className="text-xs text-destructive">{t("appointments.noPastMove")}</p>}
          </div>
          {update.isError && (
            <p role="alert" className="text-sm text-destructive">{errorMessage(tc, update.error)}</p>
          )}
        </Dialog>
      )}
    </div>
  )
}

function EditDialog({
  appointment, saving, error, onCancel, onSave,
}: {
  appointment: Appointment
  saving: boolean
  error: string | null
  onCancel: () => void
  onSave: (startsAt: string) => void
}) {
  const { t } = useTranslation("admin")
  const { t: tc } = useTranslation()
  const slots = useMemo(() => timeSlots(), [])
  // "YYYY-MM-DDTHH:MM:SS" in business time: read from the string, not parsed
  // with the browser's clock/zone.
  const [initialDate = "", initialTime = ""] = appointment.startsAt.split("T")
  const [date, setDate] = useState(initialDate)
  const [time, setTime] = useState(slots.includes(initialTime.slice(0, 5)) ? initialTime.slice(0, 5) : slots[0])
  const [pastDate, setPastDate] = useState(false)

  function save() {
    const value = `${date}T${time}:00`
    if (!isFutureDateTime(value)) {
      setPastDate(true)
      return
    }
    onSave(value)
  }

  return (
    <Dialog
      open
      onClose={onCancel}
      title={t("appointments.editTitle")}
      closeLabel={tc("close")}
      footer={
        <>
          <Button variant="secondary" onClick={onCancel}>{tc("cancel")}</Button>
          <Button onClick={save} disabled={saving}>{saving ? t("appointments.saving") : tc("save")}</Button>
        </>
      }
    >
      <p className="text-xs text-muted-foreground">
        {customerName(appointment)} · {appointment.serviceName}
      </p>
      <div className="grid grid-cols-2 gap-3">
        <div className="space-y-1.5">
          <label htmlFor="edit-date" className="text-sm text-foreground">{t("appointments.date")}</label>
          <Input
            id="edit-date"
            type="date"
            value={date}
            min={todayInBusinessTz()}
            onChange={(e) => {
              setDate(e.target.value)
              setPastDate(false)
            }}
            className="[color-scheme:dark]"
          />
        </div>
        <div className="space-y-1.5">
          <label htmlFor="edit-time" className="text-sm text-foreground">{t("appointments.time")}</label>
          <select
            id="edit-time"
            value={time}
            onChange={(e) => {
              setTime(e.target.value)
              setPastDate(false)
            }}
            className="w-full bg-input border-2 border-primary text-foreground text-sm rounded-xl px-4 py-2.5 focus:outline-none"
          >
            {slots.map((h) => (
              <option key={h} value={h}>{h}</option>
            ))}
          </select>
        </div>
      </div>
      {pastDate && <p role="alert" className="text-sm text-destructive">{t("appointments.mustBeFuture")}</p>}
      {error && <p role="alert" className="text-sm text-destructive">{error}</p>}
    </Dialog>
  )
}
