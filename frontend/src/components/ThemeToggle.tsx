import { Moon, Sun } from "lucide-react"
import { useTheme } from "@/context/ThemeContext"

export function ThemeToggle({ className = "" }: { className?: string }) {
  const { resolved, toggle } = useTheme()
  return (
    <button
      type="button"
      onClick={toggle}
      aria-label={resolved === "dark" ? "Cambiar a modo claro" : "Cambiar a modo oscuro"}
      className={`inline-flex h-9 w-9 items-center justify-center rounded-full border border-border/60 bg-card/40 text-foreground transition-colors hover:bg-muted ${className}`}
    >
      {resolved === "dark" ? <Sun className="h-4 w-4" /> : <Moon className="h-4 w-4" />}
    </button>
  )
}
