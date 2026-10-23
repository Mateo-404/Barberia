/**
 * Editor de marca del panel (F-03).
 *
 * Permite editar el borrador, ver la vista previa en vivo, publicar y
 * restaurar versiones anteriores. Todos los valores vienen del backend:
 * no se guarda nada de la marca en el navegador.
 */

import { useEffect, useMemo, useState } from "react"
import { useNavigate } from "react-router-dom"
import { User } from "lucide-react"
import { ApiError } from "@/lib/api-client"
import {
  useAdminSettings,
  usePublishSettings,
  useSaveDraft,
  validationDetails,
} from "@/api/settings"
import { DEFAULT_SETTINGS, normalizeSettings, safeLogoSrc } from "@/lib/settings"
import { buildPreviewHtml } from "@/lib/settings-preview"
import {
  BrandFields,
  BusinessFields,
  ColorFields,
  StyleFields,
  TypographyFields,
} from "@/components/settings/SettingsFields"
import type {
  Settings,
  SettingsFieldErrors,
  SettingsPublishResponse,
  SettingsValidationDetail,
} from "@/types/settings"

const SECTIONS = [
  { id: "marca", label: "Marca" },
  { id: "colores", label: "Colores" },
  { id: "tipo", label: "Tipografía" },
  { id: "estilo", label: "Estilo" },
  { id: "negocio", label: "Negocio" },
]

export default function MarcaAdmin() {
  const navigate = useNavigate()
  const { data: state, isLoading, isError, error, refetch } = useAdminSettings()

  const [settings, setSettings] = useState<Settings>(DEFAULT_SETTINGS)
  const [errors, setErrors] = useState<SettingsFieldErrors>({})
  const [warnings, setWarnings] = useState<SettingsValidationDetail[]>([])
  const [generalError, setGeneralError] = useState("")
  const [published, setPublished] = useState<SettingsPublishResponse | null>(null)
  const [dirty, setDirty] = useState(false)
  const [active, setActive] = useState("marca")
  const logoSrc = safeLogoSrc(settings.logo)

  const saveDraft = useSaveDraft()
  const publish = usePublishSettings()

  useEffect(() => {
    if (!state?.draft) return
    setSettings(normalizeSettings(state.draft))
    setDirty(false)
  }, [state?.draft])

  const previewHtml = useMemo(() => buildPreviewHtml(settings), [settings])

  function update(next: Settings) {
    setSettings(next)
    setDirty(true)
    setPublished(null)
  }

  async function onSaveDraft() {
    setGeneralError("")
    setErrors({})
    setWarnings([])
    try {
      const result = await saveDraft.mutateAsync(settings)
      setSettings(normalizeSettings(result.draft))
      setWarnings(result.warnings ?? [])
      setDirty(false)
    } catch (err) {
      const details = validationDetails(err)
      if (details.length > 0) {
        const byField: SettingsFieldErrors = {}
        for (const detail of details) byField[detail.field] = detail.message
        setErrors(byField)
        setGeneralError("Revisá los campos marcados.")
        return
      }
      setGeneralError(err instanceof ApiError ? err.detail : "No se pudo guardar el borrador.")
    }
  }

  async function onPublish() {
    setGeneralError("")
    try {
      const result = await publish.mutateAsync()
      setPublished(result)
    } catch (err) {
      setGeneralError(err instanceof ApiError ? err.detail : "No se pudo publicar.")
    }
  }


  if (isLoading) {
    return <p className="p-6 text-sm text-muted-foreground">Cargando la configuración…</p>
  }

  if (isError) {
    return (
      <div className="space-y-4 p-6">
        <div role="alert" className="border-l-2 border-destructive pl-4 text-sm text-destructive">
          {error instanceof ApiError ? error.detail : "No se pudo cargar la configuración."}
        </div>
        <button onClick={() => void refetch()} className="text-sm underline underline-offset-4">
          Reintentar
        </button>
      </div>
    )
  }

  return (
    <div className="mx-auto w-full max-w-[1400px] px-4 py-6 sm:px-6 lg:px-8">
      {/* Header */}
      <header className="mb-8 flex flex-wrap items-end justify-between gap-4 border-b border-border/60 pb-6">
        <div className="flex items-center gap-4">
          <span className="inline-flex h-12 w-12 items-center justify-center overflow-hidden rounded-full bg-foreground/5 ring-1 ring-border">
            {logoSrc ? (
              <img src={logoSrc} alt="Logo" className="h-full w-full object-cover" />
            ) : (
              <User className="h-5 w-5 text-foreground/70" />
            )}
          </span>
          <div>
            <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground/70">Estudio de marca</p>
            <h1 className="mt-1 text-2xl font-semibold tracking-tight text-foreground sm:text-3xl">
              {settings.name || "Sin nombre"}
            </h1>
            <p className="mt-1 text-sm text-muted-foreground">
              {state?.onboarded
                ? `Publicado · v${state.published_version} · ${state.hash}`
                : "Borrador sin publicar"}
              {dirty ? " · cambios sin guardar" : ""}
            </p>
          </div>
        </div>
        <button
          onClick={() => navigate("/admin")}
          className="text-sm text-muted-foreground transition-colors hover:text-foreground"
        >
          ← Volver al panel
        </button>
      </header>

      {generalError && (
        <div role="alert" className="mb-6 border-l-2 border-destructive pl-4 text-sm text-destructive">
          {generalError}
        </div>
      )}

      {published && (
        <div role="status" className="mb-6 border-l-2 border-primary pl-4 text-sm text-primary">
          Versión {published.version} publicada · hash {published.hash}
        </div>
      )}

      {warnings.length > 0 && (
        <div className="mb-6 border-l-2 border-warning pl-4 text-sm text-warning">
          <p>Avisos de contraste:</p>
          <ul className="mt-1 space-y-0.5">
            {warnings.map((w) => (
              <li key={`${w.field}-${w.message}`}>
                {w.field}: {w.message}
              </li>
            ))}
          </ul>
        </div>
      )}

      {/* Layout asimétrico: nav | form | preview */}
      <div className="grid grid-cols-1 gap-8 lg:grid-cols-[180px_minmax(0,1fr)_360px]">
        {/* Nav de secciones */}
        <aside className="lg:sticky lg:top-6 lg:self-start">
          <nav className="flex flex-row gap-1 overflow-x-auto lg:flex-col lg:gap-0">
            {SECTIONS.map((s) => (
              <button
                key={s.id}
                type="button"
                onClick={() => setActive(s.id)}
                className={`relative px-3 py-2 text-left text-sm transition-colors ${
                  active === s.id
                    ? "text-foreground"
                    : "text-muted-foreground hover:text-foreground/80"
                }`}
              >
                {active === s.id && (
                  <span className="absolute left-0 top-1/2 h-4 w-0.5 -translate-y-1/2 bg-primary" />
                )}
                {s.label}
              </button>
            ))}
          </nav>

        </aside>

        {/* Formulario */}
        <main>
          <div className="border-t border-border/40 pt-6">
            {active === "marca" && <BrandFields value={settings} onChange={update} errors={errors} />}
            {active === "colores" && <ColorFields value={settings} onChange={update} errors={errors} />}
            {active === "tipo" && <TypographyFields value={settings} onChange={update} errors={errors} />}
            {active === "estilo" && <StyleFields value={settings} onChange={update} errors={errors} />}
            {active === "negocio" && <BusinessFields value={settings} onChange={update} errors={errors} />}
          </div>

          <div className="mt-8 flex flex-wrap items-center gap-3 border-t border-border/40 pt-6">
            <button
              type="button"
              onClick={() => void onSaveDraft()}
              disabled={saveDraft.isPending}
              className="rounded-sm bg-foreground px-4 py-2 text-sm font-medium text-background transition-opacity hover:opacity-90 disabled:opacity-50"
            >
              {saveDraft.isPending ? "Guardando…" : "Guardar borrador"}
            </button>
            <button
              type="button"
              onClick={() => void onPublish()}
              disabled={publish.isPending}
              className="rounded-sm border border-primary px-4 py-2 text-sm font-medium text-primary transition-colors hover:bg-primary hover:text-primary-foreground disabled:opacity-50"
            >
              {publish.isPending ? "Publicando…" : "Publicar"}
            </button>
            {dirty && (
              <span className="text-xs text-muted-foreground">Hay cambios sin guardar.</span>
            )}
          </div>
        </main>

        {/* Preview */}
        <aside className="lg:sticky lg:top-6 lg:self-start">
          <p className="mb-2 text-xs text-muted-foreground/60">Vista previa</p>
          <div
            className="overflow-hidden rounded-sm border border-border/60 bg-card"
            style={{ boxShadow: "8px 8px 0 0 var(--color-primary, #c9a24b)" }}
          >
            <iframe
              title="Vista previa de la marca"
              srcDoc={previewHtml}
              className="h-[460px] w-full"
            />
          </div>
          <p className="mt-3 text-[11px] text-muted-foreground/50">
            Lo que publicás es lo que ven tus clientes.
          </p>
        </aside>
      </div>

    </div>
  )
}
