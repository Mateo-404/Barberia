import { createContext, useCallback, useContext, useEffect, useState, type ReactNode } from "react"

type Theme = "light" | "dark" | "system"
type Resolved = "light" | "dark"

interface ThemeContextValue {
  theme: Theme
  resolved: Resolved
  setTheme: (t: Theme) => void
  toggle: () => void
}

const STORAGE_KEY = "theme"

const ThemeContext = createContext<ThemeContextValue | null>(null)

function systemPreference(): Resolved {
  if (typeof window === "undefined") return "dark"
  return window.matchMedia("(prefers-color-scheme: light)").matches ? "light" : "dark"
}

function apply(resolved: Resolved) {
  const root = document.documentElement
  if (resolved === "light") root.classList.add("light")
  else root.classList.remove("light")
}

export function ThemeProvider({ children }: { children: ReactNode }) {
  const [theme, setThemeState] = useState<Theme>(() => {
    if (typeof window === "undefined") return "system"
    const saved = window.localStorage.getItem(STORAGE_KEY) as Theme | null
    return saved === "light" || saved === "dark" || saved === "system" ? saved : "system"
  })
  const [resolved, setResolved] = useState<Resolved>(() =>
    theme === "system" ? systemPreference() : (theme as Resolved),
  )

  useEffect(() => {
    const r = theme === "system" ? systemPreference() : (theme as Resolved)
    setResolved(r)
    apply(r)
  }, [theme])

  useEffect(() => {
    if (theme !== "system") return
    const mq = window.matchMedia("(prefers-color-scheme: light)")
    const handler = () => {
      const r = systemPreference()
      setResolved(r)
      apply(r)
    }
    mq.addEventListener("change", handler)
    return () => mq.removeEventListener("change", handler)
  }, [theme])

  const setTheme = useCallback((t: Theme) => {
    setThemeState(t)
    if (t === "system") window.localStorage.removeItem(STORAGE_KEY)
    else window.localStorage.setItem(STORAGE_KEY, t)
  }, [])

  const toggle = useCallback(() => {
    setTheme(resolved === "dark" ? "light" : "dark")
  }, [resolved, setTheme])

  return (
    <ThemeContext.Provider value={{ theme, resolved, setTheme, toggle }}>
      {children}
    </ThemeContext.Provider>
  )
}

export function useTheme() {
  const ctx = useContext(ThemeContext)
  if (!ctx) throw new Error("useTheme must be used within ThemeProvider")
  return ctx
}
