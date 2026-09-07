import { useState } from "react"
import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { Navigate, useNavigate } from "react-router-dom"
import { useAuth } from "@/context/AuthContext"
import { loginSchema, type LoginFormData } from "@/lib/schemas/login-schema"
import { ApiError } from "@/lib/api-client"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"

export default function Login() {
  const { login, isAuthenticated } = useAuth()
  const navigate = useNavigate()
  const [showPassword, setShowPassword] = useState(false)

  const {
    register,
    handleSubmit,
    setError,
    formState: { errors, isSubmitting },
  } = useForm<LoginFormData>({
    resolver: zodResolver(loginSchema),
    defaultValues: { email: "", contrasenia: "" },
  })

  if (isAuthenticated) return <Navigate to="/admin" replace />

  async function onSubmit(data: LoginFormData) {
    try {
      await login(data.email, data.contrasenia)
      navigate("/admin", { replace: true })
    } catch (err) {
      let message = "Error de conexión. Intentá de nuevo."
      if (err instanceof ApiError) {
        if (err.status === 401) {
          message = "Email o contraseña incorrectos."
        } else if (err.status === 0) {
          message = err.detail
        } else {
          message = err.detail
        }
      }
      setError("root", { message })
    }
  }

  return (
    <main className="min-h-screen flex items-center justify-center p-6 bg-background">
      <div className="w-full max-w-sm space-y-6 page-enter">
        <div className="text-center space-y-2">
          <img src="/logo.svg" alt="TH Barber Club" className="mx-auto h-16 sm:h-20 mb-4" />
          <h1 className="text-xl font-bold text-foreground">¡Bienvenido de vuelta!</h1>
          <p className="text-sm text-muted-foreground">Inicia sesión para acceder a tu cuenta</p>
        </div>

        {errors.root && (
          <div className="bg-destructive/10 text-destructive text-sm p-3 rounded-lg text-center">
            {errors.root.message}
          </div>
        )}

        <div className="bg-card rounded-2xl p-6">
          <form onSubmit={handleSubmit(onSubmit)} className="space-y-5">
            <div className="space-y-2">
              <Label htmlFor="email">Correo electrónico</Label>
              <Input
                id="email"
                type="email"
                placeholder="admin@barberia.com"
                autoComplete="email"
                autoFocus
                {...register("email")}
              />
              {errors.email && (
                <p className="text-xs text-destructive">{errors.email.message}</p>
              )}
            </div>

            <div className="space-y-2">
              <Label htmlFor="contrasenia">Contraseña</Label>
              <div className="relative">
                <Input
                  id="contrasenia"
                  type={showPassword ? "text" : "password"}
                  placeholder="Tu contraseña"
                  autoComplete="current-password"
                  className="pr-12"
                  {...register("contrasenia")}
                />
                <button
                  type="button"
                  onClick={() => setShowPassword((v) => !v)}
                  className="absolute right-2 top-1/2 -translate-y-1/2 text-sm text-muted-foreground hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-focus-ring rounded px-1"
                  aria-label={showPassword ? "Ocultar contraseña" : "Mostrar contraseña"}
                >
                  {showPassword ? "Ocultar" : "Mostrar"}
                </button>
              </div>
              {errors.contrasenia && (
                <p className="text-xs text-destructive">{errors.contrasenia.message}</p>
              )}
            </div>

            <div className="flex items-center justify-between">
              <label className="flex items-center gap-2 text-sm text-muted-foreground cursor-pointer">
                <input type="checkbox" className="accent-primary rounded" />
                Recordarme
              </label>
              <a href="#" className="text-sm text-primary no-underline hover:underline focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-focus-ring rounded">
                ¿Olvidaste tu contraseña?
              </a>
            </div>

            <Button type="submit" disabled={isSubmitting} className="w-full py-6 text-base">
              {isSubmitting ? "Ingresando..." : "Iniciar Sesión"}
            </Button>
          </form>
        </div>

        <div className="text-center space-y-3">
          <div className="flex justify-center gap-6">
            <a href="#" className="text-xs text-foreground/70 no-underline hover:underline focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-focus-ring rounded">Términos</a>
            <a href="#" className="text-xs text-foreground/70 no-underline hover:underline focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-focus-ring rounded">Privacidad</a>
            <a href="/" className="text-xs text-foreground/70 no-underline hover:underline focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-focus-ring rounded">Inicio</a>
          </div>
          <p className="text-xs text-muted-foreground">© 2026 TH Barber Club. Todos los derechos reservados.</p>
        </div>
      </div>
    </main>
  )
}
