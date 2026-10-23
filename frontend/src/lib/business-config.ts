/**
 * ╔══════════════════════════════════════════════════════════════════════════╗
 * ║ CONFIGURACIÓN DEL NEGOCIO — MÓDULO INTERINO                             ║
 * ║                                                                          ║
 * ║ Estos valores estaban codificados en Reserva.tsx y TurnosAdmin.tsx       ║
 * ║ (duplicados) y en turno-schema.ts. Se centralizan aquí mientras el       ║
 * ║ backend no expone la configuración pública del negocio:                  ║
 * ║   GET /publico/negocios/{slug}/configuracion                            ║
 * ║ Ver docs/frontend/04-contrato-configuracion-negocio.md.                  ║
 * ║                                                                          ║
 * ║ Al existir ese endpoint, este módulo se reemplaza por el hook que lo     ║
 * ║ consuma y desaparece esta configuración local. NO agregar más reglas     ║
 * ║ comerciales al cliente: el backend es la autoridad.                      ║
 * ╚══════════════════════════════════════════════════════════════════════════╝
 */

export interface BusinessConfig {
  /** IANA. Las horas de este módulo se interpretan en ESTA zona. */
  timeZone: string
  /** Apertura en "HH:MM" (hora del negocio). */
  openTime: string
  /** Cierre en "HH:MM" (hora del negocio), exclusivo. */
  closeTime: string
  /** Intervalo entre turnos en minutos. */
  intervalMinutes: number
  branding: {
    nombre: string
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
    nombre: "TH Barber Club",
    logoSrc: "/logo.svg",
    instagramUrl: "https://instagram.com",
  },
}

/** Grilla de horarios [apertura, cierre) generada desde la config. */
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

/** true si la "HH:MM" cae exactamente en un slot de la grilla. */
export function isValidSlotTime(time: string, config: BusinessConfig = businessConfig): boolean {
  return timeSlots(config).includes(time)
}
