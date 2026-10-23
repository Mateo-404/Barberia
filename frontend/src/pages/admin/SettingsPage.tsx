/**
 * Admin settings: business, account and security entry points.
 */

import { useNavigate } from "react-router-dom"
import { useTranslation } from "react-i18next"
import { Palette, Store, User, Shield } from "lucide-react"

export default function SettingsPage() {
  const navigate = useNavigate()
  const { t } = useTranslation("admin")

  const sections = [
    { key: "theme", icon: Palette, onClick: () => navigate("/admin/theme") },
    { key: "business", icon: Store, soon: true },
    { key: "account", icon: User, soon: true },
    { key: "security", icon: Shield, soon: true },
  ] as const

  return (
    <div className="mx-auto w-full max-w-3xl px-4 py-6 sm:px-6">
      <header className="mb-8 border-b border-border/60 pb-6">
        <h1 className="text-2xl font-semibold tracking-tight text-foreground sm:text-3xl">{t("settings.title")}</h1>
        <p className="mt-1 text-sm text-muted-foreground">{t("settings.subtitle")}</p>
      </header>

      <div className="space-y-3">
        {sections.map((s) => {
          const Icon = s.icon
          const soon = "soon" in s
          return (
            <button
              key={s.key}
              type="button"
              onClick={soon ? undefined : s.onClick}
              disabled={soon}
              className="group flex w-full items-center justify-between gap-4 rounded-lg border border-border/60 bg-card/40 p-4 text-left transition-colors hover:border-primary/40 hover:bg-card disabled:cursor-not-allowed disabled:opacity-50"
            >
              <div className="flex items-center gap-4">
                <span className="inline-flex h-10 w-10 items-center justify-center rounded-full bg-primary/10 text-primary">
                  <Icon className="h-5 w-5" />
                </span>
                <div>
                  <p className="text-sm font-medium text-foreground">{t(`settings.sections.${s.key}.title`)}</p>
                  <p className="text-xs text-muted-foreground">{t(`settings.sections.${s.key}.description`)}</p>
                </div>
              </div>
              <span className="text-xs text-muted-foreground">
                {soon ? t("settings.soon") : t(`settings.sections.${s.key}.action`)}
              </span>
            </button>
          )
        })}
      </div>
    </div>
  )
}
