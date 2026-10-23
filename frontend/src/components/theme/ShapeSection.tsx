import { useId } from "react"
import { useTranslation } from "react-i18next"
import type { Theme } from "@/api/settings"
import { Segmented } from "@/components/theme/Segmented"

export function ShapeSection({
  theme,
  onChange,
}: {
  theme: Theme
  onChange: (change: (t: Theme) => Theme) => void
}) {
  const { t } = useTranslation("theme")
  const sliderId = useId()
  const set = (shape: Partial<Theme["shape"]>) =>
    onChange((current) => ({ ...current, preset: "custom", shape: { ...current.shape, ...shape } }))
  const outline = theme.shape.button === "outline"

  return (
    <>
      <div className="space-y-2">
        <div className="flex items-baseline justify-between">
          <label htmlFor={sliderId} className="text-sm font-medium text-foreground">
            {t("shape.radius")}
          </label>
          <span className="text-xs tabular-nums text-muted-foreground">{theme.shape.radius} px</span>
        </div>
        <input
          id={sliderId}
          type="range"
          min={0}
          max={24}
          step={1}
          value={theme.shape.radius}
          onChange={(e) => set({ radius: Number(e.target.value) })}
          className="w-full accent-foreground"
        />
        <p className="text-xs text-muted-foreground">{t("shape.radiusHint")}</p>
      </div>

      <div className="space-y-2">
        <p className="text-sm font-medium text-foreground">{t("shape.button")}</p>
        <Segmented
          label={t("shape.button")}
          value={theme.shape.button as "solid" | "outline"}
          onChange={(button) => set({ button })}
          options={[
            { value: "solid", label: t("shape.solid") },
            { value: "outline", label: t("shape.outline") },
          ]}
        />
        <div className="flex items-center gap-3 rounded-lg bg-secondary p-3">
          <span
            aria-hidden
            className="px-4 py-1.5 text-sm font-medium"
            style={{
              borderRadius: theme.shape.radius,
              backgroundColor: outline ? "transparent" : "var(--foreground)",
              color: outline ? "var(--foreground)" : "var(--background)",
              boxShadow: outline ? "inset 0 0 0 2px var(--foreground)" : undefined,
            }}
          >
            {t("shape.sample")}
          </span>
          <span aria-hidden className="h-8 w-14 border border-foreground/40" style={{ borderRadius: theme.shape.radius }} />
        </div>
      </div>
    </>
  )
}
