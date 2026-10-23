/**
 * Admin settings: where the business and the panel are configured.
 */

import { useId } from "react"
import { useNavigate } from "react-router-dom"
import { useTranslation } from "react-i18next"
import { Palette, Store, User, Shield } from "lucide-react"
import { useAdminSettings, useUpdatePreferences, type Preferences } from "@/api/settings"
import { Switch } from "@/components/ui/switch"
import { useAuth } from "@/context/AuthContext"
import { errorMessage } from "@/lib/error-message"

type Section = {
  key: "theme" | "business" | "account" | "security"
  icon: typeof Palette
  onClick?: () => void
  soon?: boolean
  hidden?: boolean
}

function PreferenceRow({
  title,
  description,
  checked,
  disabled,
  onChange,
}: {
  title: string
  description: string
  checked: boolean
  disabled: boolean
  onChange: (checked: boolean) => void
}) {
  const titleId = useId()
  const descriptionId = useId()
  return (
    <div className="flex items-start justify-between gap-4 rounded-lg border border-border/60 bg-card/40 p-4">
      <div>
        <p id={titleId} className="text-sm font-medium text-foreground">{title}</p>
        <p id={descriptionId} className="mt-0.5 text-xs text-muted-foreground">{description}</p>
      </div>
      <Switch checked={checked} onChange={onChange} disabled={disabled} aria-labelledby={titleId} aria-describedby={descriptionId} />
    </div>
  )
}

export default function SettingsPage() {
  const navigate = useNavigate()
  const { t } = useTranslation("admin")
  const { t: tc } = useTranslation()
  const { admin } = useAuth()
  // Site settings are for owners and admins; staff only sees their own account.
  const canConfigure = admin?.role === "owner" || admin?.role === "admin"
  const settings = useAdminSettings({ enabled: canConfigure })
  const update = useUpdatePreferences()
  const preferences = settings.data?.preferences

  const configurable: Section[] = [
    { key: "theme", icon: Palette, onClick: () => navigate("/admin/theme"), hidden: preferences?.themeEditorEnabled === false },
    { key: "business", icon: Store, onClick: () => navigate("/admin/business") },
  ]
  const sections: Section[] = [
    ...(canConfigure ? configurable : []),
    { key: "account", icon: User, soon: true },
    { key: "security", icon: Shield, soon: true },
  ]

  const setPreference = (patch: Partial<Preferences>) => update.mutate(patch)

  return (
    <div className="mx-auto w-full max-w-3xl px-4 py-6 sm:px-6">
      <header className="mb-8 border-b border-border/60 pb-6">
        <h1 className="text-2xl font-semibold tracking-tight text-foreground sm:text-3xl">{t("settings.title")}</h1>
        <p className="mt-1 text-sm text-muted-foreground">{t("settings.subtitle")}</p>
      </header>

      <div className="space-y-3">
        {sections.map((s) => {
          if (s.hidden) return null
          const Icon = s.icon
          const soon = Boolean(s.soon)
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

      {canConfigure && preferences && (
        <section className="mt-10 space-y-3" aria-labelledby="panel-options">
          <div>
            <h2 id="panel-options" className="text-base font-semibold text-foreground">{t("settings.panel.title")}</h2>
            <p className="text-sm text-muted-foreground">{t("settings.panel.subtitle")}</p>
          </div>
          <PreferenceRow
            title={t("settings.panel.themeEditor.title")}
            description={t("settings.panel.themeEditor.description")}
            checked={preferences.themeEditorEnabled}
            disabled={update.isPending}
            onChange={(themeEditorEnabled) => setPreference({ themeEditorEnabled })}
          />
          <PreferenceRow
            title={t("settings.panel.onboarding.title")}
            description={t("settings.panel.onboarding.description")}
            checked={preferences.onboardingEnabled}
            disabled={update.isPending}
            onChange={(onboardingEnabled) => setPreference({ onboardingEnabled })}
          />
          {update.isError && (
            <p role="alert" className="text-sm text-destructive">{errorMessage(tc, update.error)}</p>
          )}
        </section>
      )}
    </div>
  )
}
