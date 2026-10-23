import { todayInBusinessTz } from "@/lib/datetime"

const currencyFormatter = new Intl.NumberFormat("es-AR", {
  style: "currency",
  currency: "ARS",
  maximumFractionDigits: 0,
})

export function formatCurrency(value: number): string {
  return currencyFormatter.format(value)
}

/**
 * Fecha/hora del negocio: los strings del backend son "YYYY-MM-DDTHH:MM:SS"
 * en la zona horaria del negocio, así que se leen del string y NO se
 * interpretan con `new Date()` (que los convertiría a la zona del navegador).
 */
export function formatFecha(iso: string): string {
  const [y, m, d] = iso.slice(0, 10).split("-")
  if (!y || !m || !d) return iso
  return `${d}/${m}/${y}`
}

export function formatHora(iso: string): string {
  const time = iso.split("T")[1]
  if (!time) return ""
  return time.slice(0, 5)
}

/** True si el ISO cae en el día de hoy en la zona del negocio. */
export function isHoy(iso: string): boolean {
  if (!iso) return false
  return iso.slice(0, 10) === todayInBusinessTz()
}
