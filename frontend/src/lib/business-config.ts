/**
 * Operating configuration of the business as the client needs it: time zone,
 * opening hours and slot length. The backend is the authority (it enforces
 * these rules); this module only lets the UI offer the right choices.
 *
 * It starts with the same defaults as the backend and is replaced by the
 * business's real data before the app mounts (`loadBusinessConfig`) and
 * whenever the owner saves new business data.
 */

import { API_BASE_URL } from "@/lib/api-client"

export interface BusinessConfig {
  /** IANA zone. Times in this module are read in THIS zone. */
  timeZone: string
  /** Opening time "HH:MM" (business time). */
  openTime: string
  /** Closing time "HH:MM" (business time); the last slot ends here. */
  closeTime: string
  /** Minutes between appointments. */
  intervalMinutes: number
}

export const businessConfig: BusinessConfig = {
  timeZone: "America/Argentina/Buenos_Aires",
  openTime: "13:00",
  closeTime: "20:00",
  intervalMinutes: 30,
}

interface BusinessData {
  tz: string
  open: string
  close: string
  slot: number
}

/** Replaces the configuration with the business's data. */
export function applyBusiness(business: BusinessData): void {
  businessConfig.timeZone = business.tz
  businessConfig.openTime = business.open
  businessConfig.closeTime = business.close
  businessConfig.intervalMinutes = business.slot
}

/** Fetches the published business data; keeps the defaults when it fails. */
export async function loadBusinessConfig(): Promise<void> {
  try {
    const res = await fetch(`${API_BASE_URL}/api/public/settings`, { signal: AbortSignal.timeout(3000) })
    if (res.ok) applyBusiness(((await res.json()) as { business: BusinessData }).business)
  } catch {
    // Keep the defaults: booking still works, the server enforces the rules.
  }
}

const toMinutes = (hhmm: string) => {
  const [h, m] = hhmm.split(":").map(Number)
  return h * 60 + m
}

/** Slot start times: from opening, every interval, while the slot still ends by closing. */
export function timeSlots(config: BusinessConfig = businessConfig): string[] {
  const slots: string[] = []
  const close = toMinutes(config.closeTime)
  for (let m = toMinutes(config.openTime); m + config.intervalMinutes <= close; m += config.intervalMinutes) {
    slots.push(`${String(Math.floor(m / 60)).padStart(2, "0")}:${String(m % 60).padStart(2, "0")}`)
  }
  return slots
}

/** True when "HH:MM" falls exactly on a slot of the grid. */
export function isValidSlotTime(time: string, config: BusinessConfig = businessConfig): boolean {
  return timeSlots(config).includes(time)
}
