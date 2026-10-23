import { todayInBusinessTz } from "@/lib/datetime"

/** Formats an amount of money with the app locale and currency. */
export function formatCurrency(value: number, currency = "ARS", locale = "es-AR"): string {
  return new Intl.NumberFormat(locale, { style: "currency", currency, maximumFractionDigits: 0 }).format(value)
}

/**
 * Business date/time: backend strings are "YYYY-MM-DDTHH:MM:SS" in the
 * business time zone, so they are read from the string and NEVER parsed with
 * `new Date()` (which would shift them to the browser's zone).
 */
export function formatDate(iso: string): string {
  const [y, m, d] = iso.slice(0, 10).split("-")
  if (!y || !m || !d) return iso
  return `${d}/${m}/${y}`
}

export function formatTime(iso: string): string {
  const time = iso.split("T")[1]
  if (!time) return ""
  return time.slice(0, 5)
}

/** True when the ISO string falls on today in the business time zone. */
export function isToday(iso: string): boolean {
  if (!iso) return false
  return iso.slice(0, 10) === todayInBusinessTz()
}
