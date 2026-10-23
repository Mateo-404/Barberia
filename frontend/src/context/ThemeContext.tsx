import { useLocation } from "react-router-dom"
import { createContext, useCallback, useContext, useEffect, useMemo, useState, type ReactNode } from "react"
import { isPreview } from "@/lib/site-preview"
import { THEME_APPLIED_EVENT, currentThemeMode, type ThemeMode } from "@/lib/tenant-theme"

type Scheme = "light" | "dark"

interface ThemeContextValue {
  /** Scheme in effect right now. */
  resolved: Scheme
  /** Mode the business chose: fixed light/dark, or "auto" (visitor decides). */
  mode: ThemeMode
  /** Whether a visitor may flip the scheme (always in the admin panel). */
  canToggle: boolean
  toggle: () => void
}

const STORAGE_KEY = "scheme"

const ThemeContext = createContext<ThemeContextValue | null>(null)

const systemScheme = (): Scheme =>
  window.matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light"

function savedScheme(): Scheme | null {
  try {
    const value = window.localStorage.getItem(STORAGE_KEY)
    return value === "light" || value === "dark" ? value : null
  } catch {
    return null
  }
}

/**
 * Owns the `data-scheme` attribute. The generated CSS (backend) defines both
 * schemes; here we only choose which one applies:
 *
 *   mode auto          visitor's saved choice, else the system preference
 *   mode light / dark  the business's choice; the admin panel may still flip it
 *
 * With nothing chosen the attribute is left unset and the CSS decides.
 */
export function ThemeProvider({ children }: { children: ReactNode }) {
  const adminArea = useLocation().pathname.startsWith("/admin")
  const [mode, setMode] = useState<ThemeMode>(currentThemeMode)
  const [choice, setChoice] = useState<Scheme | null>(savedScheme)
  const [system, setSystem] = useState<Scheme>(systemScheme)

  useEffect(() => {
    const onApplied = () => setMode(currentThemeMode())
    window.addEventListener(THEME_APPLIED_EVENT, onApplied)
    return () => window.removeEventListener(THEME_APPLIED_EVENT, onApplied)
  }, [])

  useEffect(() => {
    const query = window.matchMedia("(prefers-color-scheme: dark)")
    const onChange = () => setSystem(systemScheme())
    query.addEventListener("change", onChange)
    return () => query.removeEventListener("change", onChange)
  }, [])

  const canToggle = mode === "auto" || adminArea
  const fixed = mode === "light" || mode === "dark" ? mode : null
  const resolved: Scheme = (canToggle ? choice : null) ?? fixed ?? system

  useEffect(() => {
    // In the editor's preview iframe the editor owns the scheme.
    if (isPreview()) return
    const root = document.documentElement
    const explicit = canToggle ? choice : null
    if (explicit) root.dataset.scheme = explicit
    else delete root.dataset.scheme
  }, [canToggle, choice])

  const toggle = useCallback(() => {
    const next: Scheme = resolved === "dark" ? "light" : "dark"
    setChoice(next)
    try {
      window.localStorage.setItem(STORAGE_KEY, next)
    } catch {
      // Private mode: the choice lasts for this page only.
    }
  }, [resolved])

  const value = useMemo(() => ({ resolved, mode, canToggle, toggle }), [resolved, mode, canToggle, toggle])
  return <ThemeContext.Provider value={value}>{children}</ThemeContext.Provider>
}

export function useTheme(): ThemeContextValue {
  const ctx = useContext(ThemeContext)
  if (!ctx) throw new Error("useTheme must be used within ThemeProvider")
  return ctx
}
