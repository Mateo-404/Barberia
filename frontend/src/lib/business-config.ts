/**
 * Interim business configuration: opening hours and the slot grid used by the
 * booking wizard, until every screen reads them from the backend's public
 * settings (`GET /api/public/settings`, `biz.*`). Do NOT add more commercial
 * rules here: the backend is the authority.
 */

export interface BusinessConfig {
  /** IANA. Times in this module are read in THIS zone. */
  timeZone: string
  /** Opening time "HH:MM" (business time). */
  openTime: string
  /** Closing time "HH:MM" (business time), exclusive. */
  closeTime: string
  /** Minutes between appointments. */
  intervalMinutes: number
  branding: {
    name: string
    logoSrc: string
    instagramUrl: string
  }
}

export const businessConfig: BusinessConfig = {
  timeZone: "America/Argentina/Buenos_Aires",
  openTime: "13:00",
  closeTime: "20:00",
  intervalMinutes: 30,
  branding: {
    name: "TH Barber Club",
    logoSrc: "/logo.svg",
    instagramUrl: "https://instagram.com",
  },
}

/** Slot grid [opening, closing) built from the config. */
export function timeSlots(config: BusinessConfig = businessConfig): string[] {
  const [oh, om] = config.openTime.split(":").map(Number)
  const [ch, cm] = config.closeTime.split(":").map(Number)
  const open = oh * 60 + om
  const close = ch * 60 + cm
  const slots: string[] = []
  for (let m = open; m < close; m += config.intervalMinutes) {
    const h = String(Math.floor(m / 60)).padStart(2, "0")
    const min = String(m % 60).padStart(2, "0")
    slots.push(`${h}:${min}`)
  }
  return slots
}

/** True when "HH:MM" falls exactly on a grid slot. */
export function isValidSlotTime(time: string, config: BusinessConfig = businessConfig): boolean {
  return timeSlots(config).includes(time)
}
