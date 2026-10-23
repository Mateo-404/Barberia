import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"
import { Navigate, useNavigate } from "react-router-dom"
import { useAdminSettings, usePublish, useUpdatePreferences } from "@/api/settings"
import { IdentitySection } from "@/components/theme/IdentitySection"
import { PresetGallery, applyPreset } from "@/components/theme/PresetGallery"
import { PreviewCanvas } from "@/components/theme/PreviewCanvas"
import { Button } from "@/components/ui/button"
import { useThemeDraft } from "@/hooks/useThemeDraft"
import { useThemePreview } from "@/hooks/useThemePreview"
import { errorMessage, fieldErrorMessage } from "@/lib/error-message"
import { loadFont } from "@/lib/fonts"
import { loose } from "@/lib/i18n-loose"

const STEPS = ["identity", "style", "done"] as const

/**
 * First-login wizard: name the business, pick a style, publish. It is a short
 * guided version of the theme editor (same draft, same live preview) and it
 * can be skipped for good or turned off from Settings.
 */
export default function Welcome() {
  const { t } = useTranslation("theme")
  const { t: tc } = useTranslation()
  const navigate = useNavigate()
  const settings = useAdminSettings()
  const publish = usePublish()
  const preferences = useUpdatePreferences()

  const draft = useThemeDraft(settings.data?.draft)
  const preview = useThemePreview(draft.theme)
  const { setCanSave } = draft
  useEffect(() => setCanSave(preview.valid), [preview.valid, setCanSave])
  useEffect(() => void loadFont("DM Sans"), [])

  const [step, setStep] = useState(0)
  const theme = draft.theme

  if (settings.isLoading || !theme) {
    return (
      <div className="neutral-ui flex h-dvh items-center justify-center text-sm text-muted-foreground" role="status">
        {t("loading")}
      </div>
    )
  }
  if (settings.data && (!settings.data.preferences.onboardingEnabled || settings.data.onboarded)) {
    return <Navigate to="/admin" replace />
  }

  const errors = Object.fromEntries(preview.errors.map((e) => [e.field, fieldErrorMessage(tc, e.code)]))
  const identityValid = !errors.name && theme.name.trim().length >= 2
  const scheme = theme.mode === "dark" ? "dark" : "light"

  async function finish(destination: string) {
    if (!(await draft.flush())) return
    try {
      await publish.mutateAsync()
      navigate(destination, { replace: true })
    } catch {
      // The error shows under the buttons.
    }
  }

  async function skipForGood() {
    await preferences.mutateAsync({ onboardingEnabled: false })
    navigate("/admin", { replace: true })
  }

  return (
    <div className="neutral-ui grid h-dvh grid-rows-[auto_minmax(16rem,40%)] lg:grid-cols-[26rem_1fr] lg:grid-rows-1">
      <main className="flex min-h-0 flex-col overflow-y-auto border-border bg-card lg:border-r">
        <div className="flex items-center justify-between px-6 pt-5 text-xs text-muted-foreground">
          <span>{t("welcome.step", { current: step + 1, total: STEPS.length })}</span>
          <button
            type="button"
            onClick={() => void skipForGood()}
            disabled={preferences.isPending}
            className="rounded px-1 underline underline-offset-4 hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-focus-ring"
          >
            {t("welcome.skip")}
          </button>
        </div>

        <div className="flex-1 space-y-5 px-6 py-6">
          <div className="space-y-1">
            <h1 className="text-2xl font-semibold text-foreground">{t(`welcome.${STEPS[step]}.title`)}</h1>
            <p className="text-sm text-muted-foreground">{t(`welcome.${STEPS[step]}.body`)}</p>
          </div>

          {step === 0 && <IdentitySection theme={theme} onChange={draft.update} errors={errors} />}
          {step === 1 && (
            <PresetGallery
              selected={theme.preset}
              scheme={scheme}
              onPick={(preset) => draft.update((current) => applyPreset(current, preset))}
            />
          )}
          {step === 2 && (
            <div className="space-y-3 text-sm text-muted-foreground">
              <p>{t("welcome.done.summary", { name: theme.name, style: loose(t)(`presets.names.${theme.preset}`) })}</p>
              {publish.isError && (
                <p role="alert" className="text-destructive">
                  {errorMessage(tc, publish.error)}
                </p>
              )}
            </div>
          )}
        </div>

        <div className="flex flex-wrap items-center justify-between gap-2 border-t border-border px-6 py-4">
          <Button variant="ghost" size="sm" onClick={() => setStep((s) => s - 1)} disabled={step === 0}>
            {tc("back")}
          </Button>
          {step < STEPS.length - 1 ? (
            <Button size="sm" onClick={() => setStep((s) => s + 1)} disabled={step === 0 && !identityValid}>
              {t("welcome.next")}
            </Button>
          ) : (
            <div className="flex flex-wrap gap-2">
              <Button variant="secondary" size="sm" onClick={() => void finish("/admin/theme")} disabled={publish.isPending}>
                {t("welcome.done.customize")}
              </Button>
              <Button size="sm" onClick={() => void finish("/admin")} disabled={publish.isPending}>
                {publish.isPending ? t("publish.publishing") : t("welcome.done.publish")}
              </Button>
            </div>
          )}
        </div>
      </main>

      <div className="min-h-0 lg:order-last">
        <PreviewCanvas
          css={preview.result?.css ?? null}
          identity={{ name: theme.name, tagline: theme.tagline, logo: theme.logo }}
          device="mobile"
          scheme={null}
        />
      </div>
    </div>
  )
}
