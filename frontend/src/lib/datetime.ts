import { businessConfig } from "@/lib/business-config"

/**
 * Fechas y horas con zona horaria del NEGOCIO (no la del navegador).
 * `fechaHora` viaja al backend como "YYYY-MM-DDTHH:MM:00" sin offset (contrato
 * del API actual: el backend lo interpreta en la zona del negocio). Todas las
 * comparaciones de "futuro/pasado" se hacen en esa zona.
 */

interface BusinessNow {
  /** "YYYY-MM-DD" en la zona del negocio. */
  date: string
  /** Minutos desde 00:00 en la zona del negocio. */
  minutes: number
}

export function nowInBusinessTz(timeZone: string = businessConfig.timeZone, at: Date = new Date()): BusinessNow {
  const fmt = new Intl.DateTimeFormat("en-CA", {
    timeZone,
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
    hourCycle: "h23",
  })
  const parts = Object.fromEntries(
    fmt.formatToParts(at).filter((p) => p.type !== "literal").map((p) => [p.type, p.value]),
  )
  return {
    date: `${parts.year}-${parts.month}-${parts.day}`,
    minutes: Number(parts.hour) * 60 + Number(parts.minute),
  }
}

/** "YYYY-MM-DD" de hoy en la zona del negocio (para `min` de inputs date). */
export function todayInBusinessTz(timeZone: string = businessConfig.timeZone): string {
  return nowInBusinessTz(timeZone).date
}

function minutesOf(time: string): number {
  const [h, m] = time.split(":").map(Number)
  return h * 60 + m
}

/**
 * ¿La fecha/hora (hora del negocio) todavía no pasó?
 * Compara contra el reloj actual **en la zona del negocio**.
 */
export function isFutureInBusinessTz(
  date: string,
  time: string,
  timeZone: string = businessConfig.timeZone,
  at: Date = new Date(),
): boolean {
  const now = nowInBusinessTz(timeZone, at)
  if (date > now.date) return true
  if (date < now.date) return false
  return minutesOf(time) > now.minutes
}

/** Validación de un `fechaHora` ISO local "YYYY-MM-DDTHH:MM(:SS)". */
export function isFutureFechaHora(fechaHora: string, timeZone: string = businessConfig.timeZone): boolean {
  const [date, time] = fechaHora.split("T")
  if (!date || !time) return false
  return isFutureInBusinessTz(date, time.slice(0, 5), timeZone)
}
