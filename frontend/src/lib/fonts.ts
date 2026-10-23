/**
 * Self-hosted typefaces a business can pick (the CSP only allows our own
 * origin). Each font is a lazy chunk: it downloads only when a theme that
 * uses it is applied, or when the editor shows its specimen.
 *
 * Keep this list in sync with FontAllowlist in
 * backend-go/internal/settings/fonts.go (a test checks it).
 */

type Loader = () => Promise<unknown>

const both = (a: Loader, b: Loader): Loader => () => Promise.all([a(), b()])

const LOADERS: Record<string, Loader> = {
  "Playfair Display": both(() => import("@fontsource/playfair-display/400.css"), () => import("@fontsource/playfair-display/700.css")),
  Merriweather: both(() => import("@fontsource/merriweather/400.css"), () => import("@fontsource/merriweather/700.css")),
  Lora: both(() => import("@fontsource/lora/400.css"), () => import("@fontsource/lora/700.css")),
  Cinzel: both(() => import("@fontsource/cinzel/400.css"), () => import("@fontsource/cinzel/700.css")),
  "Cormorant Garamond": both(() => import("@fontsource/cormorant-garamond/400.css"), () => import("@fontsource/cormorant-garamond/700.css")),
  Inter: both(() => import("@fontsource/inter/400.css"), () => import("@fontsource/inter/700.css")),
  "DM Sans": both(() => import("@fontsource/dm-sans/400.css"), () => import("@fontsource/dm-sans/700.css")),
  Montserrat: both(() => import("@fontsource/montserrat/400.css"), () => import("@fontsource/montserrat/700.css")),
  Poppins: both(() => import("@fontsource/poppins/400.css"), () => import("@fontsource/poppins/700.css")),
  Raleway: both(() => import("@fontsource/raleway/400.css"), () => import("@fontsource/raleway/700.css")),
  "Work Sans": both(() => import("@fontsource/work-sans/400.css"), () => import("@fontsource/work-sans/700.css")),
  "Bebas Neue": () => import("@fontsource/bebas-neue/400.css"),
  Oswald: both(() => import("@fontsource/oswald/400.css"), () => import("@fontsource/oswald/700.css")),
  "Barlow Condensed": both(() => import("@fontsource/barlow-condensed/400.css"), () => import("@fontsource/barlow-condensed/700.css")),
}

/** Every font the app can load, in display order. */
export const FONT_NAMES = Object.keys(LOADERS)

/** Generic family used while a font loads (and when it is unknown). */
export const FONT_FALLBACK: Record<string, "serif" | "sans-serif"> = {
  "Playfair Display": "serif",
  Lora: "serif",
  Merriweather: "serif",
  Cinzel: "serif",
  "Cormorant Garamond": "serif",
}

export const fontStack = (name: string) => `"${name}",${FONT_FALLBACK[name] ?? "sans-serif"}`

const loaded = new Map<string, Promise<unknown>>()

/** Starts loading a font once; unknown names are ignored. */
export function loadFont(name: string): Promise<unknown> {
  const loader = LOADERS[name]
  if (!loader) return Promise.resolve()
  let pending = loaded.get(name)
  if (!pending) {
    pending = loader().catch(() => loaded.delete(name))
    loaded.set(name, pending)
  }
  return pending
}

/** The first family of a CSS font stack: `"Lora",serif` → `Lora`. */
export function firstFamily(stack: string): string {
  return stack.split(",")[0].trim().replace(/^["']|["']$/g, "")
}
