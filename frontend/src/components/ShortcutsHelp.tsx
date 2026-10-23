import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"
import { useShortcutsRegistry, type ShortcutDef } from "@/hooks/useShortcutsRegistry"
import { useHotkeys } from "@/hooks/useHotkeys"

function kbd(key: string) {
  return (
    <kbd className="px-1.5 py-0.5 rounded bg-muted text-[11px] font-mono text-foreground whitespace-nowrap">
      {key}
    </kbd>
  )
}

export function ShortcutsHelp({ scope }: { scope: "global" | "customer" | "admin" }) {
  const { t } = useTranslation("shortcuts")
  const { t: tc } = useTranslation()
  const [open, setOpen] = useState(false)
  const all = useShortcutsRegistry()
  const own = all.filter((e) => e.scope === scope)

  useHotkeys(
    {
      "Shift+?": () => setOpen(true),
    },
    true,
  )

  useEffect(() => {
    if (!open) return
    function onKey(e: KeyboardEvent) {
      if (e.key === "Escape" || e.key === "?") setOpen(false)
    }
    document.addEventListener("keydown", onKey)
    return () => document.removeEventListener("keydown", onKey)
  }, [open])

  if (!open) return null

  const grouped = own.reduce<Record<string, ShortcutDef[]>>((acc, e) => {
    ;(acc[e.scope] ??= []).push(e)
    return acc
  }, {})
  const scopes = (["global", "customer", "admin"] as const).filter((s) => grouped[s]?.length)

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4"
      role="dialog"
      aria-modal="true"
      onClick={() => setOpen(false)}
    >
      <div
        className="w-full max-w-md bg-card rounded-2xl border border-border p-5 space-y-4 max-h-[80vh] overflow-y-auto"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex items-center justify-between">
          <h2 className="text-lg font-bold text-foreground">{t("title")}</h2>
          <button
            onClick={() => setOpen(false)}
            className="text-muted-foreground hover:text-foreground rounded-lg px-2 py-1 transition-colors"
            aria-label={tc("close")}
          >
            ✕
          </button>
        </div>
        {scopes.map((s) => (
          <div key={s} className="space-y-1.5">
            <h3 className="text-xs uppercase tracking-wide text-muted-foreground font-medium">
              {t(`scopes.${s}`)}
            </h3>
            {grouped[s].map((d, i) => (
              <div key={i} className="flex items-center justify-between gap-4 text-sm">
                <span className="text-foreground">{d.description}</span>
                <span>{kbd(d.key)}</span>
              </div>
            ))}
          </div>
        ))}
        {own.length === 0 && (
          <p className="text-sm text-muted-foreground">{t("empty")}</p>
        )}
      </div>
    </div>
  )
}