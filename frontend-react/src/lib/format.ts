const currencyFormatter = new Intl.NumberFormat("es-AR", {
  style: "currency",
  currency: "ARS",
  maximumFractionDigits: 0,
})

export function formatCurrency(value: number): string {
  return currencyFormatter.format(value)
}

export function formatFecha(iso: string): string {
  const d = new Date(iso)
  return d.toLocaleDateString("es-AR", {
    day: "2-digit",
    month: "2-digit",
    year: "numeric",
  })
}

export function formatHora(iso: string): string {
  const d = new Date(iso)
  return d.toLocaleTimeString("es-AR", {
    hour: "2-digit",
    minute: "2-digit",
  })
}

/** Fecha de hoy en formato YYYY-MM-DD para inputs type="date". */
export function todayInputValue(): string {
  const d = new Date()
  const mm = String(d.getMonth() + 1).padStart(2, "0")
  const dd = String(d.getDate()).padStart(2, "0")
  return `${d.getFullYear()}-${mm}-${dd}`
}

/** True si el ISO cae en el día de hoy (misma fecha local). */
export function isHoy(iso: string): boolean {
  if (!iso) return false
  return new Date(iso).toDateString() === new Date().toDateString()
}
