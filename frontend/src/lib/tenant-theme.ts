/**
 * Runtime of the business theme.
 *
 * The backend renders the whole theme as CSS custom properties
 * (`/api/public/theme.css`). Here we only (1) put that CSS in the document,
 * (2) load the fonts it names, and (3) expose the few values the app needs
 * from it (mode, button style). Mapping tokens to Tailwind lives in
 * index.css, so there is no color logic in JavaScript.
 *
 * If the backend does not answer, the app starts with the defaults of
 * index.css: the theme is an enhancement, never a blocker.
 */

import { API_BASE_URL } from "@/lib/api-client"
import { firstFamily, loadFont } from "@/lib/fonts"

const THEME_STYLE_ID = "tenant-theme"
const THEME_TIMEOUT_MS = 3000

/** Fired on `window` after a theme is applied. */
export const THEME_APPLIED_EVENT = "theme:applied"

export type ThemeMode = "light" | "dark" | "auto"
export type ButtonStyle = "solid" | "outline"

/** Returns (or creates) the `<style id="tenant-theme">` of the head. */
function themeStyle(): HTMLStyleElement {
  const existing = document.getElementById(THEME_STYLE_ID)
  if (existing instanceof HTMLStyleElement) return existing
  const style = document.createElement("style")
  style.id = THEME_STYLE_ID
  document.head.appendChild(style)
  return style
}

/** Reads one custom property declared at :root in a block of theme CSS. */
export function themeVar(css: string, name: string): string | undefined {
  return new RegExp(`${name}\\s*:\\s*([^;}]+)`).exec(css)?.[1].trim()
}

/** Puts the CSS in the document, loads its fonts and announces the change. */
export function applyThemeCss(css: string): void {
  themeStyle().textContent = css
  for (const name of ["--font-heading", "--font-body"]) {
    const stack = themeVar(css, name)
    if (stack) void loadFont(firstFamily(stack))
  }
  document.documentElement.dataset.button = themeVar(css, "--button-style") === "outline" ? "outline" : "solid"
  window.dispatchEvent(new CustomEvent(THEME_APPLIED_EVENT))
}

/** The color mode of the applied theme; "auto" until a theme says otherwise. */
export function currentThemeMode(): ThemeMode {
  const mode = getComputedStyle(document.documentElement).getPropertyValue("--theme-mode").trim()
  return mode === "light" || mode === "dark" ? mode : "auto"
}

/** Downloads the published theme and applies it. */
export async function loadPublicTheme(): Promise<void> {
  try {
    const res = await fetch(`${API_BASE_URL}/api/public/theme.css`, {
      signal: AbortSignal.timeout(THEME_TIMEOUT_MS),
    })
    if (res.ok) applyThemeCss(await res.text())
  } catch {
    // No published theme or backend down: keep the defaults.
  }
}
