/**
 * Tipos del contrato "settings v1" (SPEC §5).
 *
 * El backend Go es la autoridad: estos tipos describen el JSON que se envía
 * en `PUT /api/admin/settings/draft` y que se devuelve en el `draft` de
 * `GET /api/admin/settings`.
 */

/** Paleta de la marca. Los colores son hex de 6 dígitos (`#rrggbb`). */
export interface SettingsColors {
  primary: string
  accent: string
  bg: string
  surface: string
  text: string
}

/** Estilo del botón principal. */
export type SettingsButtonStyle = "solid" | "outline"

/** Moneda admitida por el backend. */
export type SettingsCurrency = "ARS" | "USD" | "EUR" | "UYU" | "CLP"

/** Datos comerciales del negocio. */
export interface SettingsBiz {
  whatsapp: string
  instagram: string
  address: string
  currency: SettingsCurrency
  /** Duración de un turno, en minutos. */
  slot: number
  /** Anticipación mínima para reservar, en horas. */
  notice: number
  /** Apertura en `HH:MM` (24 h). */
  open: string
  /** Cierre en `HH:MM` (24 h). */
  close: string
  /** Zona horaria IANA. */
  tz: string
}

/** Documento completo de settings v1. */
export interface Settings {
  name: string
  tagline: string
  /** Vacío, `https://…` o `data:image/(png|jpeg|webp);base64,…`. */
  logo: string
  c: SettingsColors
  /** Fuente de títulos (allowlist, SPEC §6). */
  hf: string
  /** Fuente de textos (allowlist, SPEC §6). */
  bf: string
  /** Redondeo de bordes, 0–24 px. */
  r: number
  btn: SettingsButtonStyle
  biz: SettingsBiz
}

/** Un error de validación por campo (`details[]` del 422). */
export interface SettingsValidationDetail {
  field: string
  message: string
}

/** Estado de la configuración del tenant: `GET /api/admin/settings`. */
export interface SettingsAdminState {
  onboarded: boolean
  draft: Settings | null
  published_version: number
  hash: string
}

/** Una versión publicada: `GET /api/admin/settings/versions`. */
export interface SettingsVersion {
  version: number
  hash: string
  created_at: string
}

/** Respuesta de `PUT /api/admin/settings/draft`. */
export interface SettingsDraftResponse {
  draft: Settings
  warnings: SettingsValidationDetail[]
}

/** Respuesta de `POST /api/admin/settings/publish` y del rollback. */
export interface SettingsPublishResponse {
  version: number
  hash: string
}

/** Errores por campo, indexados por el path del campo (`c.primary`, …). */
export type SettingsFieldErrors = Record<string, string>
