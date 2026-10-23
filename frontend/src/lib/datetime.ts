import { businessConfig } from "@/lib/business-config"

/**
 * Dates and times in the BUSINESS time zone (not the browser's).
 * `startsAt` travels to the backend as "YYYY-MM-DDTHH:MM:00" with no offset
 * (the backend reads it in the business zone). Every "future/past"
 * comparison happens in that zone.
 */

interface BusinessNow {
  /** "YYYY-MM-DD" in the business zone. */
  date: string
  /** Minutes since 00:00 in the business zone. */
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

/** Today as "YYYY-MM-DD" in the business zone (for date inputs' `min`). */
export function todayInBusinessTz(timeZone: string = businessConfig.timeZone): string {
  return nowInBusinessTz(timeZone).date
}

function minutesOf(time: string): number {
  const [h, m] = time.split(":").map(Number)
  return h * 60 + m
}

/** Is the date/time (business time) still in the future? Compared against
 *  the current clock **in the business zone**. */
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

/** Validates a local ISO "YYYY-MM-DDTHH:MM(:SS)" `startsAt`. */
export function isFutureDateTime(startsAt: string, timeZone: string = businessConfig.timeZone): boolean {
  const [date, time] = startsAt.split("T")
  if (!date || !time) return false
  return isFutureInBusinessTz(date, time.slice(0, 5), timeZone)
}
