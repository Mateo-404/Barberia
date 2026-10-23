/**
 * Campos del formulario de marca (settings v1).
 *
 * Los mismos componentes usan el wizard de primer login (`/onboarding`) y el
 * editor de marca del panel (`/admin/marca`), así la validación y el texto
 * son idénticos en los dos lugares.
 */

import { useState, type ChangeEvent } from "react"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import {
  CURRENCIES,
  FONT_ALLOWLIST,
  TIMEZONES,
  contrastRatio,
  patchBiz,
  patchColors,
  patchSettings,
} from "@/lib/settings"
import type { Settings, SettingsFieldErrors } from "@/types/settings"

export interface SettingsFieldsProps {
  value: Settings
  onChange: (next: Settings) => void
  /** Errores por campo, indexados por path (`name`, `c.primary`, `biz.slot`). */
  errors?: SettingsFieldErrors
}

const selectClass =
  "flex w-full rounded-xl border-2 border-primary bg-input px-4 py-2.5 text-sm text-foreground focus-visible:border-primary/80 focus-visible:outline-none focus-visible:shadow-[0_0_0_3px_var(--focus-ring)]"

const MAX_LOGO_BYTES = 300 * 1024
const LOGO_TYPES = ["image/png", "image/jpeg", "image/webp"]

function FieldError({ message }: { message?: string }) {
  if (!message) return null
  return (
    <p role="alert" className="text-xs text-destructive">
      {message}
    </p>
  )
}

function readAsDataUri(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => resolve(String(reader.result ?? ""))
    reader.onerror = () => reject(reader.error ?? new Error("No se pudo leer el archivo"))
    reader.readAsDataURL(file)
  })
}

/* -------------------------------------------------------------------------- */
/* Paso 1 — Marca                                                             */
/* -------------------------------------------------------------------------- */

export function BrandFields({ value, onChange, errors = {} }: SettingsFieldsProps) {
  const [logoError, setLogoError] = useState("")

  async function onLogoChange(event: ChangeEvent<HTMLInputElement>) {
    const file = event.target.files?.[0]
    if (!file) return
    if (!LOGO_TYPES.includes(file.type)) {
      setLogoError("El logo tiene que ser un PNG, JPG o WebP.")
      return
    }
    if (file.size > MAX_LOGO_BYTES) {
      setLogoError("El logo no puede pesar más de 300 KB.")
      return
    }
    setLogoError("")
    try {
      const dataUri = await readAsDataUri(file)
      onChange(patchSettings(value, { logo: dataUri }))
    } catch {
      setLogoError("No se pudo leer el archivo. Probá de nuevo.")
    }
  }

  return (
    <div className="space-y-4">
      <div className="space-y-2">
        <Label htmlFor="settings-name">Nombre de la barbería</Label>
        <Input
          id="settings-name"
          value={value.name}
          maxLength={60}
          placeholder="Barbería Don Mateo"
          onChange={(e) => onChange(patchSettings(value, { name: e.target.value }))}
          aria-invalid={errors.name ? true : undefined}
        />
        <FieldError message={errors.name} />
      </div>

      <div className="space-y-2">
        <Label htmlFor="settings-tagline">Frase corta</Label>
        <Input
          id="settings-tagline"
          value={value.tagline}
          maxLength={120}
          placeholder="Corte clásico, estilo propio"
          onChange={(e) => onChange(patchSettings(value, { tagline: e.target.value }))}
          aria-invalid={errors.tagline ? true : undefined}
        />
        <FieldError message={errors.tagline} />
      </div>

      <div className="space-y-2">
        <Label htmlFor="settings-logo">Logo (PNG, JPG o WebP, hasta 300 KB)</Label>
        <input
          id="settings-logo"
          type="file"
          accept="image/png,image/jpeg,image/webp"
          onChange={onLogoChange}
          className="block w-full text-sm text-muted-foreground file:mr-3 file:rounded-lg file:border-0 file:bg-primary file:px-3 file:py-2 file:text-sm file:font-bold file:text-primary-foreground"
        />
        {value.logo ? (
          <div className="flex items-center gap-3">
            <img src={value.logo} alt="Vista previa del logo" className="h-10 w-10 rounded object-contain" />
            <button
              type="button"
              onClick={() => onChange(patchSettings(value, { logo: "" }))}
              className="text-xs text-primary underline-offset-4 hover:underline"
            >
              Quitar logo
            </button>
          </div>
        ) : (
          <p className="text-xs text-muted-foreground">Sin logo: se muestra el nombre.</p>
        )}
        <FieldError message={logoError || errors.logo} />
      </div>
    </div>
  )
}

/* -------------------------------------------------------------------------- */
/* Paso 2 — Colores                                                           */
/* -------------------------------------------------------------------------- */

const COLOR_FIELDS: Array<{ key: keyof Settings["c"]; label: string }> = [
  { key: "primary", label: "Principal" },
  { key: "accent", label: "Acento" },
  { key: "bg", label: "Fondo" },
  { key: "surface", label: "Tarjetas" },
  { key: "text", label: "Texto" },
]

export function ColorFields({ value, onChange, errors = {} }: SettingsFieldsProps) {
  const textBg = contrastRatio(value.c.text, value.c.bg)
  const primaryBg = contrastRatio(value.c.primary, value.c.bg)

  return (
    <div className="space-y-4">
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
        {COLOR_FIELDS.map(({ key, label }) => (
          <div key={key} className="space-y-2">
            <Label htmlFor={`settings-color-${key}`}>{label}</Label>
            <div className="flex items-center gap-2">
              <input
                id={`settings-color-${key}`}
                type="color"
                value={value.c[key]}
                onChange={(e) => onChange(patchColors(value, { [key]: e.target.value }))}
                className="h-11 w-14 shrink-0 cursor-pointer rounded-lg border-2 border-primary bg-input p-1"
                aria-label={label}
              />
              <Input
                value={value.c[key]}
                maxLength={7}
                onChange={(e) => onChange(patchColors(value, { [key]: e.target.value }))}
                className="font-mono text-sm"
                aria-label={`${label} en hexadecimal`}
              />
            </div>
            <FieldError message={errors[`c.${key}`]} />
          </div>
        ))}
      </div>

      <div
        className={
          textBg < 3
            ? "rounded-lg bg-destructive/10 p-3 text-xs text-destructive"
            : textBg < 4.5
              ? "rounded-lg bg-warning/10 p-3 text-xs text-warning"
              : "rounded-lg bg-success/10 p-3 text-xs text-success"
        }
      >
        <p>
          Contraste texto sobre fondo: <strong>{textBg.toFixed(1)}:1</strong>
        </p>
        {textBg < 3 ? (
          <p>Conviene al menos 3:1. Ajustá el texto o el fondo.</p>
        ) : textBg < 4.5 ? (
          <p>Conviene al menos 4,5:1 para leer cómodo.</p>
        ) : (
          <p>Contraste correcto.</p>
        )}
        {primaryBg < 3 && (
          <p className="mt-1">El botón principal tiene poco contraste sobre el fondo ({primaryBg.toFixed(1)}:1).</p>
        )}
      </div>
    </div>
  )
}

/* -------------------------------------------------------------------------- */
/* Paso 3 — Tipografía                                                        */
/* -------------------------------------------------------------------------- */

const FONT_GROUPS = ["Serif", "Sans", "Display"] as const

export function TypographyFields({ value, onChange, errors = {} }: SettingsFieldsProps) {
  return (
    <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
      <div className="space-y-2">
        <Label htmlFor="settings-hf">Títulos</Label>
        <select
          id="settings-hf"
          value={value.hf}
          onChange={(e) => onChange(patchSettings(value, { hf: e.target.value }))}
          className={selectClass}
        >
          {FONT_GROUPS.map((group) => (
            <optgroup key={group} label={group}>
              {FONT_ALLOWLIST.filter((f) => f.group === group).map((font) => (
                <option key={font.name} value={font.name} style={{ fontFamily: font.name }}>
                  {font.name}
                </option>
              ))}
            </optgroup>
          ))}
        </select>
        <FieldError message={errors.hf} />
      </div>

      <div className="space-y-2">
        <Label htmlFor="settings-bf">Textos</Label>
        <select
          id="settings-bf"
          value={value.bf}
          onChange={(e) => onChange(patchSettings(value, { bf: e.target.value }))}
          className={selectClass}
        >
          {FONT_GROUPS.map((group) => (
            <optgroup key={group} label={group}>
              {FONT_ALLOWLIST.filter((f) => f.group === group).map((font) => (
                <option key={font.name} value={font.name} style={{ fontFamily: font.name }}>
                  {font.name}
                </option>
              ))}
            </optgroup>
          ))}
        </select>
        <FieldError message={errors.bf} />
      </div>
    </div>
  )
}

/* -------------------------------------------------------------------------- */
/* Paso 4 — Estilo                                                            */
/* -------------------------------------------------------------------------- */

export function StyleFields({ value, onChange, errors = {} }: SettingsFieldsProps) {
  return (
    <div className="space-y-5">
      <div className="space-y-2">
        <Label htmlFor="settings-radius">Redondeo de bordes: {value.r} px</Label>
        <input
          id="settings-radius"
          type="range"
          min={0}
          max={24}
          step={1}
          value={value.r}
          onChange={(e) => onChange(patchSettings(value, { r: Number(e.target.value) }))}
          className="w-full accent-primary"
        />
        <div className="flex justify-between text-xs text-muted-foreground">
          <span>Recto</span>
          <span>Redondo</span>
        </div>
        <FieldError message={errors.r} />
      </div>

      <fieldset className="space-y-2">
        <legend className="text-sm font-medium">Estilo de botones</legend>
        <div className="grid grid-cols-1 gap-2 sm:grid-cols-2">
          {(
            [
              { value: "solid", label: "Rellenos" },
              { value: "outline", label: "Con borde" },
            ] as const
          ).map((option) => (
            <label
              key={option.value}
              className={`flex cursor-pointer items-center gap-2 rounded-xl border-2 p-3 text-sm ${
                value.btn === option.value ? "border-primary bg-accent" : "border-border"
              }`}
            >
              <input
                type="radio"
                name="settings-btn"
                value={option.value}
                checked={value.btn === option.value}
                onChange={() => onChange(patchSettings(value, { btn: option.value }))}
                className="accent-primary"
              />
              {option.label}
            </label>
          ))}
        </div>
        <FieldError message={errors.btn} />
      </fieldset>
    </div>
  )
}

/* -------------------------------------------------------------------------- */
/* Paso 5 — Tu negocio                                                        */
/* -------------------------------------------------------------------------- */

export function BusinessFields({ value, onChange, errors = {} }: SettingsFieldsProps) {
  const biz = value.biz

  return (
    <div className="space-y-4">
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
        <div className="space-y-2">
          <Label htmlFor="settings-whatsapp">WhatsApp</Label>
          <Input
            id="settings-whatsapp"
            type="text"
            inputMode="tel"
            value={biz.whatsapp}
            placeholder="+54 9 3492 123456"
            onChange={(e) => onChange(patchBiz(value, { whatsapp: e.target.value }))}
            aria-invalid={errors["biz.whatsapp"] ? true : undefined}
          />
          <FieldError message={errors["biz.whatsapp"]} />
        </div>

        <div className="space-y-2">
          <Label htmlFor="settings-instagram">Instagram</Label>
          <Input
            id="settings-instagram"
            type="text"
            value={biz.instagram}
            placeholder="@mibarberia"
            onChange={(e) => onChange(patchBiz(value, { instagram: e.target.value }))}
            aria-invalid={errors["biz.instagram"] ? true : undefined}
          />
          <FieldError message={errors["biz.instagram"]} />
        </div>
      </div>

      <div className="space-y-2">
        <Label htmlFor="settings-address">Dirección</Label>
        <Input
          id="settings-address"
          type="text"
          value={biz.address}
          placeholder="Av. Corrientes 1234"
          onChange={(e) => onChange(patchBiz(value, { address: e.target.value }))}
          aria-invalid={errors["biz.address"] ? true : undefined}
        />
        <FieldError message={errors["biz.address"]} />
      </div>

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
        <div className="space-y-2">
          <Label htmlFor="settings-open">Apertura</Label>
          <Input
            id="settings-open"
            type="time"
            value={biz.open}
            onChange={(e) => onChange(patchBiz(value, { open: e.target.value }))}
            aria-invalid={errors["biz.open"] ? true : undefined}
          />
          <FieldError message={errors["biz.open"]} />
        </div>

        <div className="space-y-2">
          <Label htmlFor="settings-close">Cierre</Label>
          <Input
            id="settings-close"
            type="time"
            value={biz.close}
            onChange={(e) => onChange(patchBiz(value, { close: e.target.value }))}
            aria-invalid={errors["biz.close"] ? true : undefined}
          />
          <FieldError message={errors["biz.close"]} />
        </div>
      </div>

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
        <div className="space-y-2">
          <Label htmlFor="settings-slot">Minutos por turno</Label>
          <Input
            id="settings-slot"
            type="number"
            min={10}
            max={240}
            step={5}
            value={biz.slot}
            onChange={(e) => onChange(patchBiz(value, { slot: Number(e.target.value) }))}
            aria-invalid={errors["biz.slot"] ? true : undefined}
          />
          <FieldError message={errors["biz.slot"]} />
        </div>

        <div className="space-y-2">
          <Label htmlFor="settings-notice">Anticipación mínima (hs)</Label>
          <Input
            id="settings-notice"
            type="number"
            min={0}
            max={168}
            value={biz.notice}
            onChange={(e) => onChange(patchBiz(value, { notice: Number(e.target.value) }))}
            aria-invalid={errors["biz.notice"] ? true : undefined}
          />
          <FieldError message={errors["biz.notice"]} />
        </div>
      </div>

      <div className="space-y-2">
        <Label htmlFor="settings-currency">Moneda</Label>
        <select
          id="settings-currency"
          value={biz.currency}
          onChange={(e) =>
            onChange(patchBiz(value, { currency: e.target.value as Settings["biz"]["currency"] }))
          }
          className={selectClass}
          aria-invalid={errors["biz.currency"] ? true : undefined}
        >
          {CURRENCIES.map((currency) => (
            <option key={currency} value={currency}>
              {currency}
            </option>
          ))}
        </select>
        <FieldError message={errors["biz.currency"]} />
      </div>

      <div className="space-y-2">
        <Label htmlFor="settings-tz">Zona horaria</Label>
        <select
          id="settings-tz"
          value={biz.tz}
          onChange={(e) => onChange(patchBiz(value, { tz: e.target.value }))}
          className={selectClass}
          aria-invalid={errors["biz.tz"] ? true : undefined}
        >
          {TIMEZONES.map((tz) => (
            <option key={tz} value={tz}>
              {tz}
            </option>
          ))}
        </select>
        <FieldError message={errors["biz.tz"]} />
      </div>
    </div>
  )
}
