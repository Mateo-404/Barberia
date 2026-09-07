import { Outlet, useNavigate } from "react-router-dom"
import { useAuth } from "@/context/AuthContext"
import { Button } from "@/components/ui/button"

export default function AdminLayout() {
  const { admin, logout } = useAuth()
  const navigate = useNavigate()

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

      <div className="flex items-center justify-between px-4 mb-4">
        <span className="text-sm text-muted-foreground">
          {admin?.nombre} {admin?.apellido}
        </span>
        <Button variant="ghost" size="sm" onClick={handleLogout}>
          Cerrar sesión
        </Button>
      </div>

      <main className="flex-1 px-4 pb-8">
        <Outlet />
      </main>
    </div>
  )
}
