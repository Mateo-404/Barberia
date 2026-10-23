import { Check } from "lucide-react"
import { useEffect } from "react"
import { useTranslation } from "react-i18next"
import type { Theme } from "@/api/settings"
import { usePresets, type PresetView } from "@/api/theme"
import { fontStack, loadFont } from "@/lib/fonts"
import { loose } from "@/lib/i18n-loose"

/** What picking a preset changes: look and feel, never the business identity. */
export function applyPreset(theme: Theme, preset: PresetView): Theme {
  const { brand, mode, font, shape } = preset.theme
  return { ...theme, preset: preset.id, brand, mode, font, shape, overrides: { light: {}, dark: {} } }
}

/**
 * Each preset is shown as a miniature of itself: its background, a heading
 * in its own typeface and its primary button, so the choice is made by eye.
 */
export function PresetGallery({
  selected,
  scheme,
  onPick,
}: {
  selected: string
  scheme: "light" | "dark"
  onPick: (preset: PresetView) => void
}) {
  const { t } = useTranslation("theme")
  const { data: presets, isError } = usePresets()

  useEffect(() => {
    for (const preset of presets ?? []) {
      void loadFont(preset.theme.font.heading)
      void loadFont(preset.theme.font.body)
    }
  }, [presets])

  if (isError) return <p className="text-sm text-destructive">{t("presets.loadError")}</p>
  if (!presets) return <div className="h-40 animate-pulse rounded-lg bg-secondary" />

  return (
    <ul className="grid grid-cols-2 gap-3">
      {presets.map((preset) => {
        // Each preset is shown in its own mode; "auto" ones follow the editor.
        const presetScheme = preset.theme.mode === "auto" ? scheme : preset.theme.mode
        const colors = presetScheme === "dark" ? preset.dark : preset.light
        const isSelected = preset.id === selected
        return (
          <li key={preset.id}>
            <button
              type="button"
              aria-pressed={isSelected}
              onClick={() => onPick(preset)}
              className="group block w-full text-left focus-visible:outline-none"
            >
              <span
                className={`relative flex h-24 flex-col justify-between p-3 ring-1 transition-shadow group-focus-visible:ring-2 group-focus-visible:ring-focus-ring ${
                  isSelected ? "ring-2 ring-foreground" : "ring-border group-hover:ring-foreground/40"
                }`}
                style={{ backgroundColor: colors.bg, color: colors.text, borderRadius: preset.theme.shape.radius + 4 }}
              >
                <span className="text-2xl leading-none" style={{ fontFamily: fontStack(preset.theme.font.heading) }}>
                  Aa
                </span>
                <span
                  className="self-start px-2.5 py-1 text-[11px] font-medium"
                  style={{
                    backgroundColor: preset.theme.shape.button === "outline" ? "transparent" : colors.primary,
                    color: preset.theme.shape.button === "outline" ? colors.primary : colors.onPrimary,
                    boxShadow: preset.theme.shape.button === "outline" ? `inset 0 0 0 1.5px ${colors.primary}` : undefined,
                    borderRadius: preset.theme.shape.radius,
                    fontFamily: fontStack(preset.theme.font.body),
                  }}
                >
                  {t("presets.sampleButton")}
                </span>
                {isSelected && (
                  <span className="absolute right-2 top-2 inline-flex h-5 w-5 items-center justify-center rounded-full bg-foreground text-background">
                    <Check className="h-3 w-3" aria-hidden />
                  </span>
                )}
              </span>
              <span className="mt-1.5 block text-sm font-medium text-foreground">{loose(t)(`presets.names.${preset.id}`)}</span>
              <span className="block text-xs text-muted-foreground">{loose(t)(`presets.descriptions.${preset.id}`)}</span>
            </button>
          </li>
        )
      })}
    </ul>
  )
}
