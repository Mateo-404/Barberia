import { Moon, Sun, SunMoon } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"
import type { Theme } from "@/api/settings"
import type { PreviewResult, SchemeView } from "@/api/theme"
import { ColorField } from "@/components/theme/ColorField"
import { Segmented } from "@/components/theme/Segmented"

type Scheme = "light" | "dark"
type Token = "bg" | "surface" | "text" | "primary" | "accent"
const TOKENS: Token[] = ["bg", "surface", "text", "primary", "accent"]

/** WCAG level of a contrast ratio: 7 / 4.5 for text, 3 for interface parts. */
function level(token: Token, ratio: number): "aaa" | "aa" | "low" {
  if (token === "primary") return ratio >= 3 ? "aa" : "low"
  if (ratio >= 7) return "aaa"
  return ratio >= 4.5 ? "aa" : "low"
}

export function ColorSection({
  theme,
  onChange,
  preview,
  errors,
}: {
  theme: Theme
  onChange: (change: (t: Theme) => Theme) => void
  preview: PreviewResult | null
  errors: Record<string, string>
}) {
  const { t } = useTranslation("theme")
  const [scheme, setScheme] = useState<Scheme>(theme.mode === "dark" ? "dark" : "light")
  const [advanced, setAdvanced] = useState(false)

  const generated: SchemeView | undefined = preview?.[scheme]
  const overrides = theme.overrides[scheme]

  const setOverride = (token: Token, value: string) =>
    onChange((current) => ({
      ...current,
      preset: "custom",
      overrides: { ...current.overrides, [scheme]: { ...current.overrides[scheme], [token]: value } },
    }))
  const clearOverride = (token: Token) =>
    onChange((current) => {
      const next = { ...current.overrides[scheme] }
      delete next[token]
      return { ...current, overrides: { ...current.overrides, [scheme]: next } }
    })

  const ratioFor = (token: Token): number | undefined => {
    if (!generated) return undefined
    return { text: generated.ratios.text, primary: generated.ratios.primary }[token as "text" | "primary"]
  }

  return (
    <>
      <div className="space-y-2">
        <p className="text-sm font-medium text-foreground">{t("color.brand")}</p>
        <ColorField
          label={t("color.brand")}
          value={theme.brand}
          error={errors.brand}
          onChange={(brand) => onChange((current) => ({ ...current, preset: "custom", brand }))}
        />
        <p className="text-xs text-muted-foreground">{t("color.brandHint")}</p>
      </div>

      <div className="space-y-2">
        <p className="text-sm font-medium text-foreground">{t("color.mode")}</p>
        <Segmented
          label={t("color.mode")}
          value={theme.mode as "light" | "dark" | "auto"}
          onChange={(mode) => onChange((current) => ({ ...current, mode }))}
          options={[
            { value: "light", label: t("color.modes.light"), icon: <Sun className="h-3.5 w-3.5" aria-hidden /> },
            { value: "dark", label: t("color.modes.dark"), icon: <Moon className="h-3.5 w-3.5" aria-hidden /> },
            { value: "auto", label: t("color.modes.auto"), icon: <SunMoon className="h-3.5 w-3.5" aria-hidden /> },
          ]}
        />
        <p className="text-xs text-muted-foreground">{t(`color.modeHints.${theme.mode}`)}</p>
      </div>

      <div className="space-y-3">
        <button
          type="button"
          aria-expanded={advanced}
          onClick={() => setAdvanced((v) => !v)}
          className="text-sm font-medium text-foreground underline underline-offset-4 hover:no-underline focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-focus-ring"
        >
          {advanced ? t("color.hideAdvanced") : t("color.showAdvanced")}
        </button>

        {advanced && (
          <div className="space-y-4 rounded-lg border border-border p-4">
            <Segmented
              label={t("color.scheme")}
              value={scheme}
              onChange={setScheme}
              options={[
                { value: "light", label: t("color.modes.light") },
                { value: "dark", label: t("color.modes.dark") },
              ]}
            />
            <ul className="space-y-3">
              {TOKENS.map((token) => {
                const effective = overrides[token] ?? generated?.[token] ?? "#000000"
                const ratio = ratioFor(token)
                const grade = ratio === undefined ? null : level(token, ratio)
                return (
                  <li key={token} className="space-y-1">
                    <div className="flex items-center justify-between gap-2">
                      <span className="text-sm text-foreground">{t(`color.tokens.${token}`)}</span>
                      <span className="flex items-center gap-2">
                        {grade && ratio !== undefined && (
                          <span
                            className={`rounded px-1.5 py-0.5 text-[11px] font-medium ${
                              grade === "low" ? "bg-destructive/15 text-destructive" : "bg-secondary text-foreground"
                            }`}
                            title={t("color.ratio", { ratio })}
                          >
                            {t(`color.grades.${grade}`)} · {ratio}:1
                          </span>
                        )}
                        {overrides[token] && (
                          <button
                            type="button"
                            onClick={() => clearOverride(token)}
                            className="text-xs text-muted-foreground underline underline-offset-2 hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-focus-ring"
                          >
                            {t("color.reset")}
                          </button>
                        )}
                      </span>
                    </div>
                    <ColorField
                      size="sm"
                      label={t(`color.tokens.${token}`)}
                      value={effective}
                      error={errors[`overrides.${scheme}.${token}`]}
                      onChange={(hex) => setOverride(token, hex)}
                    />
                  </li>
                )
              })}
            </ul>
            <p className="text-xs text-muted-foreground">{t("color.advancedHint")}</p>
          </div>
        )}
      </div>
    </>
  )
}
