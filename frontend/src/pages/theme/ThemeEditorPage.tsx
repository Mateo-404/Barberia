import { ArrowLeft, Check, History, Loader2, Monitor, Moon, Smartphone, Sun, TriangleAlert } from "lucide-react"
import { useCallback, useEffect, useMemo, useState } from "react"
import { useTranslation } from "react-i18next"
import { Navigate, useNavigate } from "react-router-dom"
import { useAdminSettings, usePublish } from "@/api/settings"
import { ColorSection } from "@/components/theme/ColorSection"
import { IdentitySection } from "@/components/theme/IdentitySection"
import { PresetGallery, applyPreset } from "@/components/theme/PresetGallery"
import { PreviewCanvas, type Device } from "@/components/theme/PreviewCanvas"
import { PublishDialog } from "@/components/theme/PublishDialog"
import { Section } from "@/components/theme/Section"
import { Segmented } from "@/components/theme/Segmented"
import { ShapeSection } from "@/components/theme/ShapeSection"
import { TypographySection } from "@/components/theme/TypographySection"
import { VersionsDialog } from "@/components/theme/VersionsDialog"
import { Button } from "@/components/ui/button"
import { useHotkeys } from "@/hooks/useHotkeys"
import { useRegisterShortcuts } from "@/hooks/useShortcutsRegistry"
import { useThemeDraft, type SaveStatus } from "@/hooks/useThemeDraft"
import { useThemePreview } from "@/hooks/useThemePreview"
import { fieldErrorMessage } from "@/lib/error-message"
import { loadFont } from "@/lib/fonts"
import { loose } from "@/lib/i18n-loose"

type SectionId = "presets" | "color" | "typography" | "shape" | "identity"

const STATUS_ICON: Record<SaveStatus, React.ReactNode> = {
  saved: <Check className="h-3.5 w-3.5" aria-hidden />,
  dirty: <Loader2 className="h-3.5 w-3.5" aria-hidden />,
  saving: <Loader2 className="h-3.5 w-3.5 animate-spin" aria-hidden />,
  error: <TriangleAlert className="h-3.5 w-3.5 text-destructive" aria-hidden />,
}

/**
 * Full-screen theme editor. Lives outside the admin layout on purpose: it is
 * a workshop for one task, with the business's real booking page as the hero
 * and a quiet neutral interface around it.
 */
export default function ThemeEditorPage() {
  const { t } = useTranslation("theme")
  const { t: tc } = useTranslation()
  const { t: ts } = useTranslation("shortcuts")
  const navigate = useNavigate()
  const settings = useAdminSettings()
  const publish = usePublish()

  const draft = useThemeDraft(settings.data?.draft)
  const preview = useThemePreview(draft.theme)
  const { setCanSave } = draft
  useEffect(() => setCanSave(preview.valid), [preview.valid, setCanSave])

  const [open, setOpen] = useState<SectionId | null>("presets")
  const [device, setDevice] = useState<Device>("mobile")
  const [scheme, setScheme] = useState<"light" | "dark" | null>(null)
  const [dialog, setDialog] = useState<"publish" | "versions" | null>(null)
  const [justPublished, setJustPublished] = useState(false)

  useEffect(() => void loadFont("DM Sans"), [])
  useEffect(() => {
    if (!justPublished) return
    const timer = setTimeout(() => setJustPublished(false), 4000)
    return () => clearTimeout(timer)
  }, [justPublished])

  const errors = useMemo(
    () => Object.fromEntries(preview.errors.map((e) => [e.field, fieldErrorMessage(tc, e.code)])),
    [preview.errors, tc],
  )
  const warnings = preview.result?.warnings ?? []
  const theme = draft.theme

  const flushAndLeave = useCallback(async () => {
    const saved = await draft.flush()
    if (saved || window.confirm(t("exit.discard"))) navigate("/admin/settings")
  }, [draft, navigate, t])

  async function confirmPublish() {
    if (!(await draft.flush())) return
    try {
      await publish.mutateAsync()
      setDialog(null)
      setJustPublished(true)
    } catch {
      // The dialog shows the failure.
    }
  }

  useHotkeys({ "Ctrl+s": () => void draft.flush() })
  useRegisterShortcuts("admin", [{ key: "Ctrl+s", description: ts("theme.save") }])

  if (settings.isLoading || !theme) {
    return (
      <div className="neutral-ui flex h-dvh items-center justify-center text-sm text-muted-foreground" role="status">
        {t("loading")}
      </div>
    )
  }
  if (settings.data && !settings.data.preferences.themeEditorEnabled) return <Navigate to="/admin/settings" replace />

  const effectiveScheme = scheme ?? (theme.mode === "dark" ? "dark" : "light")
  const canPublish = preview.valid && draft.status !== "error" && (draft.status !== "saved" || !!settings.data?.hasUnpublishedChanges)
  const identity = { name: theme.name, tagline: theme.tagline, logo: theme.logo }

  const toggle = (id: SectionId) => setOpen((current) => (current === id ? null : id))
  const presetName = theme.preset === "custom" ? t("presets.custom") : loose(t)(`presets.names.${theme.preset}`)

  return (
    <div className="neutral-ui flex h-dvh flex-col">
      <header className="flex flex-wrap items-center gap-x-4 gap-y-2 border-b border-border bg-card px-4 py-2.5">
        <button
          type="button"
          onClick={() => void flushAndLeave()}
          className="inline-flex items-center gap-1.5 rounded-md px-2 py-1.5 text-sm font-medium text-foreground hover:bg-accent focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-focus-ring"
        >
          <ArrowLeft className="h-4 w-4" aria-hidden />
          {t("exit.label")}
        </button>

        <div className="min-w-0">
          <h1 className="truncate text-sm font-semibold text-foreground">{t("title")}</h1>
          <p className="flex items-center gap-1.5 text-xs text-muted-foreground" role="status" aria-live="polite">
            {STATUS_ICON[draft.status]}
            {justPublished ? t("status.published") : t(`status.${draft.status}`)}
          </p>
        </div>

        <div className="ml-auto flex flex-wrap items-center gap-2">
          <div className="hidden lg:block">
            <Segmented
              label={t("device.label")}
              value={device}
              onChange={setDevice}
              options={[
                { value: "mobile", label: t("device.mobile"), icon: <Smartphone className="h-3.5 w-3.5" aria-hidden /> },
                { value: "desktop", label: t("device.desktop"), icon: <Monitor className="h-3.5 w-3.5" aria-hidden /> },
              ]}
            />
          </div>
          <Segmented
            label={t("scheme.label")}
            value={effectiveScheme}
            onChange={setScheme}
            options={[
              { value: "light", label: t("scheme.light"), icon: <Sun className="h-3.5 w-3.5" aria-hidden /> },
              { value: "dark", label: t("scheme.dark"), icon: <Moon className="h-3.5 w-3.5" aria-hidden /> },
            ]}
          />
          <Button variant="secondary" size="sm" onClick={() => setDialog("versions")} className="gap-1.5">
            <History className="h-4 w-4" aria-hidden />
            {t("versions.open")}
          </Button>
          <Button size="sm" onClick={() => setDialog("publish")} disabled={!canPublish} className="px-5">
            {t("publish.action")}
          </Button>
        </div>
      </header>

      <div className="grid min-h-0 flex-1 grid-rows-[minmax(18rem,46%)_1fr] lg:grid-cols-[23.5rem_1fr] lg:grid-rows-1">
        <div className="min-h-0 lg:order-2">
          <PreviewCanvas css={preview.result?.css ?? null} identity={identity} device={device} scheme={scheme} />
        </div>

        <aside className="min-h-0 overflow-y-auto border-border bg-card lg:order-1 lg:border-r" aria-label={t("sidebar")}>
          {warnings.length > 0 && (
            <div className="flex gap-2 border-b border-border bg-warning/15 px-5 py-3 text-xs text-foreground" role="status">
              <TriangleAlert className="mt-0.5 h-3.5 w-3.5 shrink-0" aria-hidden />
              <span>{t("warnings.lowContrast")}</span>
            </div>
          )}

          <Section title={t("sections.presets")} summary={presetName} open={open === "presets"} onToggle={() => toggle("presets")}>
            <PresetGallery
              selected={theme.preset}
              scheme={effectiveScheme}
              onPick={(preset) => draft.update((current) => applyPreset(current, preset))}
            />
          </Section>

          <Section
            title={t("sections.color")}
            summary={`${theme.brand} · ${t(`color.modes.${theme.mode}`)}`}
            open={open === "color"}
            onToggle={() => toggle("color")}
          >
            <ColorSection theme={theme} onChange={draft.update} preview={preview.result} errors={errors} />
          </Section>

          <Section
            title={t("sections.typography")}
            summary={theme.font.heading === theme.font.body ? theme.font.heading : `${theme.font.heading} + ${theme.font.body}`}
            open={open === "typography"}
            onToggle={() => toggle("typography")}
          >
            <TypographySection theme={theme} onChange={draft.update} errors={errors} />
          </Section>

          <Section
            title={t("sections.shape")}
            summary={`${theme.shape.radius} px · ${t(`shape.${theme.shape.button}`)}`}
            open={open === "shape"}
            onToggle={() => toggle("shape")}
          >
            <ShapeSection theme={theme} onChange={draft.update} />
          </Section>

          <Section title={t("sections.identity")} summary={theme.name} open={open === "identity"} onToggle={() => toggle("identity")}>
            <IdentitySection theme={theme} onChange={draft.update} errors={errors} />
          </Section>
        </aside>
      </div>

      {dialog === "publish" && (
        <PublishDialog
          warnings={warnings}
          publishing={publish.isPending}
          failed={publish.isError}
          onCancel={() => {
            publish.reset()
            setDialog(null)
          }}
          onConfirm={() => void confirmPublish()}
        />
      )}
      {dialog === "versions" && (
        <VersionsDialog current={settings.data?.publishedVersion ?? 0} onClose={() => setDialog(null)} />
      )}
    </div>
  )
}
