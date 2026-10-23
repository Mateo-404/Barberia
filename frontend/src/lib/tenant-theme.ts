/**
 * Bootstrap del tema público (F-02).
 *
 * El CSS publicado por el backend (`GET /api/public/theme.css`) se inyecta en
 * el `<head>` del documento ANTES de montar React, para que quede en el HTML
 * inicial que sirve la app. Además se replican las variables del tema sobre
 * las que usa la UI (Tailwind) para que la página pública use la marca.
 *
 * Si el backend no responde, la app arranca igual con los valores por
 * defecto de `index.css`: el tema es una mejora, nunca un bloqueo.
 */

import { API_BASE_URL } from "@/lib/api-client"
import { hexToRgba } from "@/lib/settings"

const THEME_STYLE_ID = "tenant-theme"
const THEME_TIMEOUT_MS = 3000

/** Devuelve (o crea) el `<style id="tenant-theme">` del head. */
function ensureThemeStyle(): HTMLStyleElement {
  const existing = document.getElementById(THEME_STYLE_ID)
  if (existing instanceof HTMLStyleElement) return existing
  const style = document.createElement("style")
  style.id = THEME_STYLE_ID
  document.head.appendChild(style)
  return style
}

/** Extrae las custom properties de un bloque CSS. */
export function extractCssVars(css: string): Record<string, string> {
  const vars: Record<string, string> = {}
  const pattern = /(--[a-z0-9-]+)\s*:\s*([^;}]+)/gi
  let match: RegExpExecArray | null
  while ((match = pattern.exec(css)) !== null) {
    vars[match[1]] = match[2].trim()
  }
  return vars
}

/**
 * Replica las variables del tema sobre las que consume la UI.
 * Las que no tienen equivalente directo (muted-foreground, border, accent)
 * se derivan para no perder la jerarquía visual de la app.
 */
function applyAppVars(vars: Record<string, string>): void {
  const root = document.documentElement
  const primary = vars["--color-primary"]
  const accent = vars["--color-accent"]
  const text = vars["--color-text"]

  const set = (name: string, value: string | undefined) => {
    if (value) root.style.setProperty(name, value)
  }

  set("--primary", primary)
  set("--primary-foreground", vars["--color-on-primary"])
  set("--background", vars["--color-bg"])
  set("--foreground", text)
  set("--card", vars["--color-surface"])
  set("--card-foreground", text)
  set("--popover", vars["--color-surface"])
  set("--popover-foreground", text)
  set("--input", vars["--color-surface"])
  set("--muted", vars["--color-surface"])
  set("--radius", vars["--radius"])

  if (text) set("--muted-foreground", hexToRgba(text, 0.72))
  if (accent) set("--accent", hexToRgba(accent, 0.15))
  if (text) set("--accent-foreground", text)
  if (primary) {
    set("--border", hexToRgba(primary, 0.25))
    set("--ring", primary)
    set("--focus-ring", hexToRgba(primary, 0.4))
  }

  // La fuente del tema se aplica al body: `@theme inline` resuelve
  // `font-sans` en tiempo de build, así que la variable no basta.
  const bodyFont = vars["--font-body"]
  if (bodyFont) document.body.style.fontFamily = bodyFont
}

/** Escribe el CSS en el head y aplica las variables a la app. */
export function applyThemeCss(css: string): void {
  ensureThemeStyle().textContent = css
  applyAppVars(extractCssVars(css))
}

/** Descarga el CSS publicado y lo deja en el head. */
export async function loadPublicTheme(): Promise<void> {
  try {
    const res = await fetch(`${API_BASE_URL}/api/public/theme.css`, {
      signal: AbortSignal.timeout(THEME_TIMEOUT_MS),
    })
    if (!res.ok) return
    applyThemeCss(await res.text())
  } catch {
    // Sin tema publicado (o backend caído): se sigue con los defaults.
  }
}
