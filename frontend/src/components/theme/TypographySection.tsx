import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"
import type { Theme } from "@/api/settings"
import { FONT_NAMES, fontStack, loadFont } from "@/lib/fonts"
import { FONT_PAIRS } from "@/lib/font-pairs"

const pairKey = (heading: string, body: string) => `${heading}|${body}`

function FontGrid({
  label,
  value,
  onPick,
}: {
  label: string
  value: string
  onPick: (name: string) => void
}) {
  return (
    <div role="radiogroup" aria-label={label} className="grid grid-cols-2 gap-1.5">
      {FONT_NAMES.map((name) => (
        <button
          key={name}
          type="button"
          role="radio"
          aria-checked={name === value}
          onClick={() => onPick(name)}
          className={`rounded-md border px-3 py-2 text-left text-sm transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-focus-ring ${
            name === value ? "border-foreground bg-card" : "border-border hover:bg-accent"
          }`}
          style={{ fontFamily: fontStack(name) }}
        >
          {name}
        </button>
      ))}
    </div>
  )
}

export function TypographySection({
  theme,
  onChange,
  errors,
}: {
  theme: Theme
  onChange: (change: (t: Theme) => Theme) => void
  errors: Record<string, string>
}) {
  const { t } = useTranslation("theme")
  const [separate, setSeparate] = useState(false)
  const current = pairKey(theme.font.heading, theme.font.body)

  // The specimens render in their own typeface, so load them all once.
  useEffect(() => {
    for (const name of FONT_NAMES) void loadFont(name)
  }, [])

  const setFont = (heading: string, body: string) =>
    onChange((t0) => ({ ...t0, preset: "custom", font: { heading, body } }))

  return (
    <>
      <div role="radiogroup" aria-label={t("typography.pairs")} className="grid grid-cols-1 gap-2">
        {FONT_PAIRS.map((pair) => {
          const selected = pairKey(pair.heading, pair.body) === current
          return (
            <button
              key={pairKey(pair.heading, pair.body)}
              type="button"
              role="radio"
              aria-checked={selected}
              onClick={() => setFont(pair.heading, pair.body)}
              className={`rounded-lg border px-4 py-3 text-left transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-focus-ring ${
                selected ? "border-foreground bg-card" : "border-border hover:bg-accent"
              }`}
            >
              <span className="block text-xl leading-tight text-foreground" style={{ fontFamily: fontStack(pair.heading) }}>
                {t("typography.sampleHeading")}
              </span>
              <span className="mt-0.5 block text-sm text-muted-foreground" style={{ fontFamily: fontStack(pair.body) }}>
                {t("typography.sampleBody")}
              </span>
              <span className="mt-1 block text-[11px] text-muted-foreground">
                {pair.heading === pair.body ? pair.heading : `${pair.heading} + ${pair.body}`}
              </span>
            </button>
          )
        })}
      </div>

      <button
        type="button"
        aria-expanded={separate}
        onClick={() => setSeparate((v) => !v)}
        className="text-sm font-medium text-foreground underline underline-offset-4 hover:no-underline focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-focus-ring"
      >
        {separate ? t("typography.hideSeparate") : t("typography.showSeparate")}
      </button>

      {separate && (
        <div className="space-y-4">
          <div className="space-y-2">
            <p className="text-sm font-medium text-foreground">{t("typography.heading")}</p>
            <FontGrid label={t("typography.heading")} value={theme.font.heading} onPick={(name) => setFont(name, theme.font.body)} />
            {errors["font.heading"] && <p role="alert" className="text-xs text-destructive">{errors["font.heading"]}</p>}
          </div>
          <div className="space-y-2">
            <p className="text-sm font-medium text-foreground">{t("typography.body")}</p>
            <FontGrid label={t("typography.body")} value={theme.font.body} onPick={(name) => setFont(theme.font.heading, name)} />
            {errors["font.body"] && <p role="alert" className="text-xs text-destructive">{errors["font.body"]}</p>}
          </div>
        </div>
      )}
    </>
  )
}
