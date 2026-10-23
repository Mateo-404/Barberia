/**
 * Ajustes del panel (cuenta, comercio, seguridad).
 */

import { useNavigate } from "react-router-dom"
import { Store, User, Shield } from "lucide-react"

export default function AjustesAdmin() {
  const navigate = useNavigate()

  const sections = [
    {
      icon: Store,
      title: "Comercio",
      description: "Marca, colores, tipografía, logo y datos del negocio.",
      action: "Administrar marca",
      onClick: () => navigate("/admin/marca"),
    },
    {
      icon: User,
      title: "Cuenta",
      description: "Nombre, apellido y contraseña del administrador.",
      action: "Próximamente",
      onClick: () => {},
      disabled: true,
    },
    {
      icon: Shield,
      title: "Seguridad",
      description: "Sesiones activas y rotación de credenciales.",
      action: "Próximamente",
      onClick: () => {},
      disabled: true,
    },
  ]

  return (
    <div className="mx-auto w-full max-w-3xl px-4 py-6 sm:px-6">
      <header className="mb-8 border-b border-border/60 pb-6">
        <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground/70">Configuración</p>
        <h1 className="mt-1 text-2xl font-semibold tracking-tight text-foreground sm:text-3xl">
          Ajustes
        </h1>
        <p className="mt-1 text-sm text-muted-foreground">
          Administrá tu cuenta y el comercio.
        </p>
      </header>

      <div className="space-y-3">
        {sections.map((s) => {
          const Icon = s.icon
          return (
            <button
              key={s.title}
              type="button"
              onClick={s.onClick}
              disabled={s.disabled}
              className="group flex w-full items-center justify-between gap-4 rounded-sm border border-border/60 bg-card/40 p-4 text-left transition-colors hover:border-primary/40 hover:bg-card disabled:cursor-not-allowed disabled:opacity-50"
            >
              <div className="flex items-center gap-4">
                <span className="inline-flex h-10 w-10 items-center justify-center rounded-full bg-primary/10 text-primary">
                  <Icon className="h-5 w-5" />
                </span>
                <div>
                  <p className="text-sm font-medium text-foreground">{s.title}</p>
                  <p className="text-xs text-muted-foreground">{s.description}</p>
                </div>
              </div>
              <span className="text-xs text-muted-foreground transition-transform group-hover:translate-x-0.5">
                {s.action} →
              </span>
            </button>
          )
        })}
      </div>
    </div>
  )
}
