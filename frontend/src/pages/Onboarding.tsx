/**
 * Wizard de primer login (F-01): configura la marca del negocio en 6 pasos.
 *
 * Al terminar guarda el borrador (`PUT /api/admin/settings/draft`) y recién
 * después lo publica (`POST /api/admin/settings/publish`). Recién ahí entra
 * al panel. Si el backend responde 422, los `details` se muestran en el campo
 * que corresponde.
 */

import { useMemo, useState } from "react"
import { useNavigate } from "react-router-dom"
import { ApiError } from "@/lib/api-client"
import { usePublishSettings, useSaveDraft, validationDetails } from "@/api/settings"
import { DEFAULT_SETTINGS } from "@/lib/settings"
import { buildPreviewHtml } from "@/lib/settings-preview"
import {
  BrandFields,
  BusinessFields,
  ColorFields,
  StyleFields,
  TypographyFields,
} from "@/components/settings/SettingsFields"
import { Button } from "@/components/ui/button"
import type { Settings, SettingsFieldErrors } from "@/types/settings"

const STEPS = ["Marca", "Colores", "Tipografía", "Estilo", "Tu negocio", "Listo"]

/** Paso del wizard donde vive cada campo (para volver al error). */
function stepForField(field: string): number {
  if (field === "name" || field === "tagline" || field === "logo") return 0
  if (field.startsWith("c.")) return 1
  if (field === "hf" || field === "bf") return 2
  if (field === "r" || field === "btn") return 3
  if (field.startsWith("biz.")) return 4
  return 5
}

export default function Onboarding() {
  const [step, setStep] = useState(0)
  const [settings, setSettings] = useState<Settings>(DEFAULT_SETTINGS)
  const [errors, setErrors] = useState<SettingsFieldErrors>({})
  const [generalError, setGeneralError] = useState("")
  const navigate = useNavigate()

  const saveDraft = useSaveDraft()
  const publish = usePublishSettings()

  const previewHtml = useMemo(() => buildPreviewHtml(settings), [settings])

  const saving = saveDraft.isPending || publish.isPending

  async function onFinish() {
    setGeneralError("")
    setErrors({})
    try {
      await saveDraft.mutateAsync(settings)
      await publish.mutateAsync()
      navigate("/admin", { replace: true })
    } catch (err) {
      const details = validationDetails(err)
      if (details.length > 0) {
        const byField: SettingsFieldErrors = {}
        for (const detail of details) byField[detail.field] = detail.message
        setErrors(byField)
        setGeneralError("Revisá los campos marcados y volvé a guardar.")
        setStep(stepForField(details[0].field))
        return
      }
      setGeneralError(
        err instanceof ApiError ? err.code : "No se pudo guardar la configuración.",
      )
    }
  }

  function onNext() {
    if (step < STEPS.length - 1) {
      setStep((current) => current + 1)
      return
    }
    void onFinish()
  }

  return (
    <div className="min-h-screen bg-background">
      <div className="mx-auto grid max-w-5xl grid-cols-1 gap-6 p-4 lg:grid-cols-[minmax(0,1fr)_400px]">
        <main className="lg:order-none lg:col-start-1 lg:row-start-1">
          <header className="mb-4">
            <h1 className="text-xl font-bold text-foreground sm:text-2xl">Configurá tu barbería</h1>
            <p className="mt-1 text-sm text-muted-foreground">
              Paso {step + 1} de {STEPS.length}: {STEPS[step]}
            </p>
          </header>

          <div className="mb-4 flex gap-1" aria-hidden="true">
            {STEPS.map((title, index) => (
              <div
                key={title}
                className={`h-1 flex-1 rounded ${index <= step ? "bg-primary" : "bg-border"}`}
              />
            ))}
          </div>

          {generalError && (
            <div role="alert" className="mb-4 rounded-lg bg-destructive/10 p-3 text-sm text-destructive">
              {generalError}
            </div>
          )}

          <div className="rounded-2xl border border-border bg-card p-4 sm:p-6">
            {step === 0 && (
              <BrandFields value={settings} onChange={setSettings} errors={errors} />
            )}
            {step === 1 && (
              <ColorFields value={settings} onChange={setSettings} errors={errors} />
            )}
            {step === 2 && (
              <TypographyFields value={settings} onChange={setSettings} errors={errors} />
            )}
            {step === 3 && <StyleFields value={settings} onChange={setSettings} errors={errors} />}
            {step === 4 && (
              <BusinessFields value={settings} onChange={setSettings} errors={errors} />
            )}
            {step === 5 && (
              <div className="space-y-4">
                <p className="text-sm text-foreground">
                  <strong>Todo listo.</strong> Podés cambiar cualquiera de estas opciones más tarde
                  desde el panel.
                </p>
                <dl className="grid grid-cols-1 gap-2 text-sm sm:grid-cols-2">
                  <div>
                    <dt className="text-muted-foreground">Nombre</dt>
                    <dd className="text-foreground">{settings.name}</dd>
                  </div>
                  <div>
                    <dt className="text-muted-foreground">Frase</dt>
                    <dd className="text-foreground">{settings.tagline}</dd>
                  </div>
                  <div>
                    <dt className="text-muted-foreground">Títulos</dt>
                    <dd className="text-foreground">{settings.hf}</dd>
                  </div>
                  <div>
                    <dt className="text-muted-foreground">Textos</dt>
                    <dd className="text-foreground">{settings.bf}</dd>
                  </div>
                  <div>
                    <dt className="text-muted-foreground">Horario</dt>
                    <dd className="text-foreground">
                      {settings.biz.open} a {settings.biz.close}
                    </dd>
                  </div>
                  <div>
                    <dt className="text-muted-foreground">Turnos de</dt>
                    <dd className="text-foreground">{settings.biz.slot} minutos</dd>
                  </div>
                </dl>
              </div>
            )}

            <div className="mt-6 flex flex-col-reverse gap-3 sm:flex-row sm:justify-between">
              <Button
                type="button"
                variant="secondary"
                onClick={() => setStep((current) => Math.max(0, current - 1))}
                disabled={step === 0 || saving}
                className="w-full sm:w-auto"
              >
                Atrás
              </Button>
              <Button type="button" onClick={onNext} disabled={saving} className="w-full sm:w-auto">
                {saving
                  ? "Guardando..."
                  : step === STEPS.length - 1
                    ? "Guardar y entrar al panel"
                    : "Siguiente"}
              </Button>
            </div>
          </div>
        </main>

        <aside
          className={`lg:order-none lg:col-start-2 lg:row-start-1 ${step === STEPS.length - 1 ? "order-first" : "order-last"}`}
        >
          <h2 className="mb-2 text-xs font-semibold uppercase tracking-widest text-muted-foreground">
            Vista previa en vivo
          </h2>
          <iframe
            title="Vista previa de la marca"
            srcDoc={previewHtml}
            className="h-[420px] w-full rounded-2xl border border-border bg-card"
          />
        </aside>
      </div>
    </div>
  )
}
