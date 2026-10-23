import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"
import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { useLocation, useNavigate } from "react-router-dom"
import { useAuth } from "@/context/AuthContext"
import { safeInternalPath } from "@/lib/paths"
import { useHotkeys } from "@/hooks/useHotkeys"
import { useRegisterShortcuts } from "@/hooks/useShortcutsRegistry"
import { loginSchema, type LoginFormData } from "@/lib/schemas/login-schema"
import { SiteMark } from "@/components/SiteMark"
import { useSiteIdentity } from "@/api/site"
import { errorMessage, fieldErrorMessage } from "@/lib/error-message"
import { useAdminSettings } from "@/api/settings"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"

export default function Login() {
  const { t } = useTranslation("admin")
  const { t: tc } = useTranslation()
  const { t: ts } = useTranslation("shortcuts")
  const site = useSiteIdentity()
  const { login, isAuthenticated } = useAuth()
  // Brand state: only queried once signed in.
  const settingsQuery = useAdminSettings({ enabled: isAuthenticated })
  const navigate = useNavigate()

  const location = useLocation()
  const from = safeInternalPath((location.state as { from?: unknown } | null)?.from) ?? "/admin"
  const [showPassword, setShowPassword] = useState(false)

  useEffect(() => {
    if (!isAuthenticated) return
    if (settingsQuery.isLoading) return
    const prefs = settingsQuery.data
    const target = prefs && !prefs.onboarded && prefs.preferences.onboardingEnabled ? "/welcome" : from
    navigate(target, { replace: true })
  }, [isAuthenticated, settingsQuery.isLoading, settingsQuery.data, from, navigate])

  const {
    register,
    handleSubmit,
    getValues,
    setError,
    formState: { errors, isSubmitting },
  } = useForm<LoginFormData>({
    resolver: zodResolver(loginSchema),
    defaultValues: { email: "", password: "" },
  })

  useHotkeys({
    Escape: () => {
      if (errors.root) {
        setError("root", { message: "" })
      } else if (getValues().email || getValues().password) {
        document.getElementById("email")?.focus()
      }
    },
    "Alt+d": () => setShowPassword((v) => !v),
    "Alt+Shift+d": () => setShowPassword((v) => !v),
    "Ctrl+Enter": () => handleSubmit(onSubmit)(),
  })

  useRegisterShortcuts("admin", [
    { key: "Esc", description: ts("login.escape") },
    { key: "Alt+d", description: ts("login.togglePassword") },
    { key: "Ctrl+Enter", description: ts("login.submit") },
  ])

  async function onSubmit(data: LoginFormData) {
    try {
      await login(data.email, data.password)
      // The redirect is handled by the effect above, depending on `onboarded`.
    } catch (err) {
      setError("root", { message: errorMessage(tc, err) })
    }
  }

  return (
    <main className="min-h-screen flex items-center justify-center p-6 bg-background">
      <div className="w-full max-w-sm space-y-6 page-enter">
        <div className="text-center space-y-2">
          <SiteMark className="mx-auto mb-4 text-3xl" imageClassName="h-16 sm:h-20" />
          <h1 className="text-xl font-bold text-foreground">{t("login.title")}</h1>
          <p className="text-sm text-muted-foreground">{t("login.subtitle")}</p>
        </div>

        {errors.root && (
          <div role="alert" className="bg-destructive/10 text-destructive text-sm p-3 rounded-lg text-center">
            {errors.root.message}
          </div>
        )}

        <div className="bg-card rounded-2xl p-6 animate-in fade-in-0 zoom-in-95 duration-200">
          <form onSubmit={handleSubmit(onSubmit)} className="space-y-5">
            <div className="space-y-2">
              <Label htmlFor="email">{t("login.email")}</Label>
              <Input
                id="email"
                type="email"
                placeholder={t("login.emailPlaceholder")}
                autoComplete="email"
                autoFocus
                {...register("email")}
                className="text-base"
              />
              {errors.email && (
                <p className="text-xs text-destructive">{fieldErrorMessage(tc, errors.email.message ?? "")}</p>
              )}
            </div>

            <div className="space-y-2">
              <Label htmlFor="password">{t("login.password")}</Label>
              <div className="relative">
                <Input
                  id="password"
                  type={showPassword ? "text" : "password"}
                  placeholder={t("login.passwordPlaceholder")}
                  autoComplete="current-password"
                  className="pr-12 text-base"
                  {...register("password")}
                />
                <button
                  type="button"
                  onClick={() => setShowPassword((v) => !v)}
                  className="absolute right-2 top-1/2 -translate-y-1/2 text-sm text-muted-foreground hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-focus-ring rounded px-1 transition-colors duration-200"
                  aria-label={showPassword ? t("login.hidePassword") : t("login.showPassword")}
                >
                  {showPassword ? t("login.hide") : t("login.show")}
                </button>
              </div>
              {errors.password && (
                <p className="text-xs text-destructive">{fieldErrorMessage(tc, errors.password.message ?? "")}</p>
              )}
            </div>

            <Button type="submit" disabled={isSubmitting} className="w-full py-6 text-base active:scale-[0.98] transition-transform duration-100">
              {isSubmitting ? t("login.submitting") : t("login.submitLabel")}
            </Button>
          </form>
        </div>

        <div className="text-center space-y-3">
          <a href="/" className="text-xs text-foreground/70 no-underline hover:underline focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-focus-ring rounded py-1">
            {t("login.home")}
          </a>
          <p className="text-xs text-muted-foreground">{t("login.rights", { year: new Date().getFullYear(), name: site.name })}</p>
        </div>
      </div>
    </main>
  )
}
