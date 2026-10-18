import { useMemo } from "react"
import { NavLink, Outlet, useNavigate } from "react-router-dom"
import { useAuth } from "@/context/AuthContext"
import { Button } from "@/components/ui/button"
import { cn } from "@/lib/utils"
import { useHotkeys } from "@/hooks/useHotkeys"
import { useRegisterShortcuts } from "@/hooks/useShortcutsRegistry"

const NAV_ITEMS = [
  { to: "/admin", label: "Panel" },
  { to: "/admin/estadisticas", label: "Estadísticas" },
  { to: "/admin/turnos", label: "Turnos" },
]

export default function AdminLayout() {
  const { admin, logout } = useAuth()
  const navigate = useNavigate()

  const hotkeyMap = useMemo(
    () => ({
      "Shift+g": () => navigate("/admin"),
      "Shift+e": () => navigate("/admin/estadisticas"),
      "Shift+t": () => navigate("/admin/turnos"),
      "Shift+h": () => navigate("/"),
      "Ctrl+Shift+q": () => {
        logout()
        navigate("/login")
      },
    }),
    [navigate, logout],
  )

  useHotkeys(hotkeyMap, true, { ignoreWhenDialogOpen: true })

  const navDefs = useMemo(
    () => [
      { key: "Shift+g", description: "Ir al Panel" },
      { key: "Shift+e", description: "Ir a Estadísticas" },
      { key: "Shift+t", description: "Ir a Turnos" },
      { key: "Shift+h", description: "Ir al sitio público" },
      { key: "Ctrl+Shift+q", description: "Cerrar sesión" },
    ],
    [],
  )
  useRegisterShortcuts("admin", navDefs)

  function handleLogout() {
    logout()
    navigate("/login")
  }

  return (
    <div className="min-h-screen bg-background flex flex-col">
      <nav
        className="flex items-center justify-between px-4"
        style={{ height: "var(--nav-height)" }}
      >
        <button onClick={() => navigate("/")} className="bg-transparent border-0 p-0 cursor-pointer rounded focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-focus-ring">
          <img src="/img/flecha.svg" alt="Volver" className="h-8" />
        </button>

        <a href="/" className="inline-flex">
          <img src="/logo.svg" alt="TH Barber Club" className="h-12" />
        </a>

        <a href="https://instagram.com" target="_blank" rel="noopener noreferrer" className="inline-flex">
          <img src="/img/icono-instagram.svg" alt="Instagram" className="h-8" />
        </a>
      </nav>

      <div className="flex flex-col gap-2 px-4 mb-4">
        <div className="flex items-center justify-between gap-4 flex-wrap">
          <span className="text-sm text-muted-foreground">
            {admin?.nombre} {admin?.apellido}
          </span>
          <Button variant="ghost" size="sm" onClick={handleLogout}>
            Cerrar sesión
          </Button>
        </div>
        <nav className="flex gap-1 bg-card rounded-xl p-1 border border-border overflow-x-auto">
          {NAV_ITEMS.map((item) => (
            <NavLink
              key={item.to}
              to={item.to}
              end={item.to === "/admin"}
              className={({ isActive }) =>
                cn(
                  "flex-1 text-center px-4 py-2 rounded-lg text-sm font-medium whitespace-nowrap transition-colors",
                  isActive
                    ? "bg-primary text-primary-foreground"
                    : "text-muted-foreground hover:text-foreground",
                )
              }
            >
              {item.label}
            </NavLink>
          ))}
        </nav>
        <div className="flex items-center justify-center gap-3 text-[11px] text-muted-foreground mt-1">
          <span><kbd className="px-1.5 py-0.5 rounded bg-muted text-[10px] font-mono">Shift+g</kbd> Panel</span>
          <span><kbd className="px-1.5 py-0.5 rounded bg-muted text-[10px] font-mono">Shift+e</kbd> Estadísticas</span>
          <span><kbd className="px-1.5 py-0.5 rounded bg-muted text-[10px] font-mono">Shift+t</kbd> Turnos</span>
        </div>
      </div>

      <main className="flex-1 px-4 pb-8">
        <Outlet />
      </main>
    </div>
  )
}
