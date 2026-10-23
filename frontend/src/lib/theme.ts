/**
 * Aplica branding del negocio a las CSS variables en runtime.
 * Interino: usa businessConfig.branding. Cuando exista el endpoint
 * GET /publico/negocios/{slug}/configuracion, se llama con los valores
 * del backend.
 */

export interface BrandingTokens {
  primary?: string
  background?: string
  card?: string
  foreground?: string
  accent?: string
  destructive?: string
  muted?: string
  border?: string
  fontSans?: string
  fontHeading?: string
}

const DEFAULT_TOKENS: Required<BrandingTokens> = {
  primary: "#d4a574",
  background: "#1c1c1c",
  card: "#2a2a2a",
  foreground: "#f5f5f5",
  accent: "#d4a574",
  destructive: "#ef4444",
  muted: "#6b6b6b",
  border: "#3a3a3a",
  fontSans: "Poppins",
  fontHeading: "Poppins",
}

export function applyBranding(tokens: Partial<BrandingTokens> = {}) {
  const t = { ...DEFAULT_TOKENS, ...tokens }
  const root = document.documentElement
  root.style.setProperty("--primary", t.primary)
  root.style.setProperty("--background", t.background)
  root.style.setProperty("--card", t.card)
  root.style.setProperty("--foreground", t.foreground)
  root.style.setProperty("--accent", t.accent)
  root.style.setProperty("--destructive", t.destructive)
  root.style.setProperty("--muted", t.muted)
  root.style.setProperty("--border", t.border)
  root.style.setProperty("--font-sans", `"${t.fontSans}", sans-serif`)
  root.style.setProperty("--font-heading", `"${t.fontHeading}", sans-serif`)
  document.body.style.fontFamily = `"${t.fontSans}", sans-serif`
}

export function applyDefaultBranding() {
  applyBranding({})
}
