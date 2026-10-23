import { isFutureDateTime } from "@/lib/datetime"
import { isToday } from "@/lib/format"

export type AppointmentStatus = "today" | "upcoming" | "past"

/** Status of an appointment relative to now, in the business time zone. */
export function appointmentStatus(startsAt: string): AppointmentStatus {
  if (isToday(startsAt)) return "today"
  return isFutureDateTime(startsAt) ? "upcoming" : "past"
}

export const STATUS_BADGE_VARIANT = {
  today: "info",
  upcoming: "warning",
  past: "success",
} as const
