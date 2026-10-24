import { useEffect } from "react"
import { Compass, RotateCcw } from "lucide-react"
import { resetDb } from "@/mocks/db"
import { DEMO_ADMIN } from "@/mocks/handlers"
import { shouldAutoStart, startTour } from "./tour"

export default function DemoControls() {
  useEffect(() => {
    if (!shouldAutoStart()) return
    const t = setTimeout(() => startTour(), 600)
    return () => clearTimeout(t)
  }, [])

  // El banner y los botones son fixed: reservamos espacio para no tapar la app.
  useEffect(() => {
    document.body.style.paddingTop = "28px"
    document.body.style.paddingBottom = "64px"
    return () => {
      document.body.style.paddingTop = ""
      document.body.style.paddingBottom = ""
    }
  }, [])

  const reiniciar = () => {
    resetDb()
    sessionStorage.clear()
    window.location.hash = "#/"
    window.location.reload()
  }

  const btn =
    "inline-flex items-center gap-1.5 rounded-full border border-primary bg-card px-4 py-2 text-xs font-bold text-primary shadow-lg transition-colors hover:bg-accent focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-focus-ring"

  return (
    <>
      <div className="fixed inset-x-0 top-0 z-40 truncate bg-primary px-4 py-1.5 text-center text-xs font-semibold text-primary-foreground">
        Modo demo · datos simulados en tu navegador · admin: {DEMO_ADMIN.email} / {DEMO_ADMIN.password}
      </div>
      <div className="fixed bottom-4 left-4 z-50 flex flex-wrap gap-2">
        <button type="button" className={btn} onClick={() => startTour()}>
          <Compass className="h-4 w-4" aria-hidden /> Visita guiada
        </button>
        <button type="button" className={btn} onClick={reiniciar}>
          <RotateCcw className="h-4 w-4" aria-hidden /> Reiniciar datos
        </button>
      </div>
    </>
  )
}
