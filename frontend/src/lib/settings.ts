/**
 * Utilidades del dominio "settings v1".
 *
 * - defaults y allowlist de fuentes (SPEC §5 y §6);
 * - contraste WCAG (misma fórmula que `backend-go/internal/settings/contrast.go`);
 * - escape de HTML para la vista previa (F-04): TODO texto que venga del
 *   usuario se escapa antes de meterse en el `srcDoc` del iframe.
 */

import type {
  Settings,
  SettingsBiz,
  SettingsColors,
  SettingsCurrency,
} from "@/types/settings"

export const DEFAULT_SETTINGS: Settings = {
  name: "Barbería Don Mateo",
  tagline: "Corte clásico, estilo propio",
  logo: "",
  c: {
    primary: "#c9a24b",
    accent: "#e8d9a8",
    bg: "#161616",
    surface: "#222222",
    text: "#f4f0e8",
  },
  hf: "Playfair Display",
  bf: "Inter",
  r: 8,
  btn: "solid",
  biz: {
    whatsapp: "",
    instagram: "",
    address: "",
    currency: "ARS",
    slot: 30,
    notice: 1,
    open: "09:00",
    close: "20:00",
    tz: "America/Argentina/Buenos_Aires",
  },
}

export interface FontOption {
  name: string
  generic: "serif" | "sans-serif"
  group: "Serif" | "Sans" | "Display"
}

/** Allowlist de fuentes (SPEC §6). Mismo contenido que el backend. */
export const FONT_ALLOWLIST: FontOption[] = [
  { name: "Playfair Display", generic: "serif", group: "Serif" },
  { name: "Lora", generic: "serif", group: "Serif" },
  { name: "Merriweather", generic: "serif", group: "Serif" },
  { name: "Cinzel", generic: "serif", group: "Serif" },
  { name: "Cormorant Garamond", generic: "serif", group: "Serif" },
  { name: "Libre Baskerville", generic: "serif", group: "Serif" },
  { name: "Inter", generic: "sans-serif", group: "Sans" },
  { name: "DM Sans", generic: "sans-serif", group: "Sans" },
  { name: "Montserrat", generic: "sans-serif", group: "Sans" },
  { name: "Poppins", generic: "sans-serif", group: "Sans" },
  { name: "Raleway", generic: "sans-serif", group: "Sans" },
  { name: "Roboto", generic: "sans-serif", group: "Sans" },
  { name: "Rubik", generic: "sans-serif", group: "Sans" },
  { name: "Work Sans", generic: "sans-serif", group: "Sans" },
  { name: "Nunito", generic: "sans-serif", group: "Sans" },
  { name: "Josefin Sans", generic: "sans-serif", group: "Sans" },
  { name: "Space Grotesk", generic: "sans-serif", group: "Sans" },
  { name: "Bebas Neue", generic: "sans-serif", group: "Display" },
  { name: "Oswald", generic: "sans-serif", group: "Display" },
  { name: "Barlow Condensed", generic: "sans-serif", group: "Display" },
  { name: "Archivo Black", generic: "sans-serif", group: "Display" },
  { name: "Anton", generic: "sans-serif", group: "Display" },
]

export const FONT_NAMES: string[] = FONT_ALLOWLIST.map((f) => f.name)

export const CURRENCIES: SettingsCurrency[] = ["ARS", "USD", "EUR", "UYU", "CLP"]

/** Zonas horarias frecuentes. El backend acepta cualquier IANA válida. */
export const TIMEZONES: string[] = [
  "America/Argentina/Buenos_Aires",
  "America/Argentina/Cordoba",
  "America/Santiago",
  "America/Bogota",
  "America/Lima",
  "America/Mexico_City",
  "America/New_York",
  "Europe/Madrid",
  "UTC",
]

/* -------------------------------------------------------------------------- */
/* Patches tipados                                                            */
/* -------------------------------------------------------------------------- */

export function patchSettings(current: Settings, patch: Partial<Settings>): Settings {
  return { ...current, ...patch }
}

export function patchColors(current: Settings, patch: Partial<SettingsColors>): Settings {
  return { ...current, c: { ...current.c, ...patch } }
}

export function patchBiz(current: Settings, patch: Partial<SettingsBiz>): Settings {
  return { ...current, biz: { ...current.biz, ...patch } }
}

/**
 * Normaliza un settings que viene del backend: completa los campos que
 * pueden faltar y fuerza los tipos esperados por el formulario.
 */
export function normalizeSettings(raw: unknown): Settings {
  const value = (raw ?? {}) as Partial<Settings>
  const c = (value.c ?? {}) as Partial<SettingsColors>
  const biz = (value.biz ?? {}) as Partial<SettingsBiz>
  return {
    name: typeof value.name === "string" ? value.name : DEFAULT_SETTINGS.name,
    tagline: typeof value.tagline === "string" ? value.tagline : DEFAULT_SETTINGS.tagline,
    logo: typeof value.logo === "string" ? value.logo : "",
    c: {
      primary: c.primary ?? DEFAULT_SETTINGS.c.primary,
      accent: c.accent ?? DEFAULT_SETTINGS.c.accent,
      bg: c.bg ?? DEFAULT_SETTINGS.c.bg,
      surface: c.surface ?? DEFAULT_SETTINGS.c.surface,
      text: c.text ?? DEFAULT_SETTINGS.c.text,
    },
    hf: typeof value.hf === "string" ? value.hf : DEFAULT_SETTINGS.hf,
    bf: typeof value.bf === "string" ? value.bf : DEFAULT_SETTINGS.bf,
    r: typeof value.r === "number" ? value.r : DEFAULT_SETTINGS.r,
    btn: value.btn === "outline" ? "outline" : "solid",
    biz: {
      whatsapp: biz.whatsapp ?? "",
      instagram: biz.instagram ?? "",
      address: biz.address ?? "",
      currency: (CURRENCIES as string[]).includes(biz.currency ?? "")
        ? (biz.currency as SettingsCurrency)
        : DEFAULT_SETTINGS.biz.currency,
      slot: typeof biz.slot === "number" ? biz.slot : DEFAULT_SETTINGS.biz.slot,
      notice: typeof biz.notice === "number" ? biz.notice : DEFAULT_SETTINGS.biz.notice,
      open: biz.open ?? DEFAULT_SETTINGS.biz.open,
      close: biz.close ?? DEFAULT_SETTINGS.biz.close,
      tz: biz.tz ?? DEFAULT_SETTINGS.biz.tz,
    },
  }
}

/* -------------------------------------------------------------------------- */
/* Escape (F-04)                                                              */
/* -------------------------------------------------------------------------- */

/** Escapa los 5 caracteres que pueden romper el HTML de la vista previa. */
export function escapeHtml(value: unknown): string {
  return String(value ?? "").replace(/[&<>"']/g, (ch) => {
    switch (ch) {
      case "&":
        return "&amp;"
      case "<":
        return "&lt;"
      case ">":
        return "&gt;"
      case '"':
        return "&quot;"
      case "'":
        return "&#39;"
      default:
        return ch
    }
  })
}

/**
 * Sólo deja pasar logos que el backend aceptaría (`https://…` o data URI
 * png/jpeg/webp). Evita que un `javascript:` llegue al `src` del preview.
 */
const LOGO_SAFE = /^(https:\/\/[^\s"'<>]+|data:image\/(?:png|jpeg|webp);base64,[A-Za-z0-9+/=]+)$/

export function safeLogoSrc(logo: string): string | null {
  return LOGO_SAFE.test(logo) ? logo : null
}

/* -------------------------------------------------------------------------- */
/* Contraste (WCAG 2.x)                                                       */
/* -------------------------------------------------------------------------- */

export function hexToRgb(hex: string): [number, number, number] | null {
  if (!/^#[0-9a-fA-F]{6}$/.test(hex)) return null
  return [
    parseInt(hex.slice(1, 3), 16),
    parseInt(hex.slice(3, 5), 16),
    parseInt(hex.slice(5, 7), 16),
  ]
}

function channel(value: number): number {
  const c = value / 255
  return c <= 0.03928 ? c / 12.92 : Math.pow((c + 0.055) / 1.055, 2.4)
}

/** Luminancia relativa WCAG. 0 si el color no es un hex válido. */
export function relativeLuminance(hex: string): number {
  const rgb = hexToRgb(hex)
  if (!rgb) return 0
  const [r, g, b] = rgb.map(channel)
  return 0.2126 * r + 0.7152 * g + 0.0722 * b
}

/** Ratio de contraste entre dos colores (1 → 21). */
export function contrastRatio(fg: string, bg: string): number {
  const l1 = relativeLuminance(fg)
  const l2 = relativeLuminance(bg)
  const [hi, lo] = l1 >= l2 ? [l1, l2] : [l2, l1]
  return (hi + 0.05) / (lo + 0.05)
}

/**
 * Color del texto sobre el botón principal: el de mayor contraste entre
 * `#111111` y `#ffffff` (empate → `#111111`). Mismo criterio que el backend.
 */
export function onPrimaryColor(bg: string): string {
  const dark = contrastRatio("#111111", bg)
  const light = contrastRatio("#ffffff", bg)
  if (dark > light) return "#111111"
  if (light > dark) return "#ffffff"
  return "#111111"
}

/** `rgba()` a partir de un hex (para tintes de hover). */
export function hexToRgba(hex: string, alpha: number): string {
  const rgb = hexToRgb(hex)
  if (!rgb) return `rgba(0, 0, 0, ${alpha})`
  return `rgba(${rgb[0]}, ${rgb[1]}, ${rgb[2]}, ${alpha})`
}

/* -------------------------------------------------------------------------- */
/* Utilidades de presentación                                                 */
/* -------------------------------------------------------------------------- */

/** "HH:MM" → minutos desde la medianoche. */
export function timeToMinutes(time: string): number {
  const [h, m] = time.split(":").map(Number)
  if (Number.isNaN(h) || Number.isNaN(m)) return 0
  return h * 60 + m
}

/** Primeras franjas horarias del negocio (para la vista previa). */
export function previewSlots(settings: Settings, limit = 5): string[] {
  const { open, close, slot } = settings.biz
  const start = timeToMinutes(open)
  const end = timeToMinutes(close)
  const step = Math.max(slot, 10)
  const slots: string[] = []
  for (let t = start; t + slot <= end && slots.length < limit; t += step) {
    const h = String(Math.floor(t / 60)).padStart(2, "0")
    const m = String(t % 60).padStart(2, "0")
    slots.push(`${h}:${m}`)
  }
  return slots
}

/** Precio de referencia formateado en la moneda del negocio. */
export function formatPrice(amount: number, currency: string): string {
  const formatted = new Intl.NumberFormat("es-AR", { maximumFractionDigits: 0 }).format(amount)
  return `${currency} ${formatted}`
}
