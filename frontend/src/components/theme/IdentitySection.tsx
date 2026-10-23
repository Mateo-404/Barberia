import { ImagePlus, Trash2 } from "lucide-react"
import { useState, type ChangeEvent } from "react"
import { useTranslation } from "react-i18next"
import type { Theme } from "@/api/settings"
import { FieldMessage } from "@/components/theme/Section"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"

const MAX_LOGO_BYTES = 300 * 1024
const LOGO_TYPES = ["image/png", "image/jpeg", "image/webp"]

function readAsDataUri(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => resolve(String(reader.result ?? ""))
    reader.onerror = () => reject(reader.error ?? new Error("read failed"))
    reader.readAsDataURL(file)
  })
}

export function IdentitySection({
  theme,
  onChange,
  errors,
}: {
  theme: Theme
  onChange: (change: (t: Theme) => Theme) => void
  errors: Record<string, string>
}) {
  const { t } = useTranslation("theme")
  const [logoError, setLogoError] = useState("")

  async function onLogo(event: ChangeEvent<HTMLInputElement>) {
    const file = event.target.files?.[0]
    event.target.value = ""
    if (!file) return
    if (!LOGO_TYPES.includes(file.type)) return setLogoError(t("identity.logoType"))
    if (file.size > MAX_LOGO_BYTES) return setLogoError(t("identity.logoSize"))
    try {
      const logo = await readAsDataUri(file)
      setLogoError("")
      onChange((current) => ({ ...current, logo }))
    } catch {
      setLogoError(t("identity.logoRead"))
    }
  }

  return (
    <>
      <div className="space-y-1.5">
        <Label htmlFor="theme-name">{t("identity.name")}</Label>
        <Input
          id="theme-name"
          value={theme.name}
          maxLength={60}
          aria-invalid={errors.name ? true : undefined}
          onChange={(e) => onChange((current) => ({ ...current, name: e.target.value }))}
          className="border border-border"
        />
        <FieldMessage message={errors.name} />
      </div>

      <div className="space-y-1.5">
        <Label htmlFor="theme-tagline">{t("identity.tagline")}</Label>
        <Input
          id="theme-tagline"
          value={theme.tagline}
          maxLength={120}
          placeholder={t("identity.taglinePlaceholder")}
          aria-invalid={errors.tagline ? true : undefined}
          onChange={(e) => onChange((current) => ({ ...current, tagline: e.target.value }))}
          className="border border-border"
        />
        <FieldMessage message={errors.tagline} />
      </div>

      <div className="space-y-2">
        <p className="text-sm font-medium text-foreground">{t("identity.logo")}</p>
        <div className="flex items-center gap-3">
          <span className="flex h-14 w-14 shrink-0 items-center justify-center overflow-hidden rounded-lg bg-secondary ring-1 ring-border">
            {theme.logo ? (
              <img src={theme.logo} alt="" className="h-full w-full object-contain" />
            ) : (
              <ImagePlus className="h-5 w-5 text-muted-foreground" aria-hidden />
            )}
          </span>
          <div className="flex flex-wrap gap-2">
            <label className="inline-flex cursor-pointer items-center rounded-md border border-border px-3 py-1.5 text-sm font-medium text-foreground transition-colors hover:bg-accent focus-within:ring-2 focus-within:ring-focus-ring">
              {theme.logo ? t("identity.replaceLogo") : t("identity.uploadLogo")}
              <input type="file" accept={LOGO_TYPES.join(",")} onChange={(e) => void onLogo(e)} className="sr-only" />
            </label>
            {theme.logo && (
              <button
                type="button"
                onClick={() => onChange((current) => ({ ...current, logo: "" }))}
                className="inline-flex items-center gap-1.5 rounded-md px-3 py-1.5 text-sm text-muted-foreground transition-colors hover:text-destructive focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-focus-ring"
              >
                <Trash2 className="h-3.5 w-3.5" aria-hidden />
                {t("identity.removeLogo")}
              </button>
            )}
          </div>
        </div>
        <p className="text-xs text-muted-foreground">{t("identity.logoHint")}</p>
        <FieldMessage message={logoError || errors.logo} />
      </div>
    </>
  )
}
