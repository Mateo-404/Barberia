import { useState } from "react"
import { useTranslation } from "react-i18next"
import { useRollback, useVersions } from "@/api/settings"
import { Button } from "@/components/ui/button"
import { Dialog } from "@/components/ui/dialog"
import { errorMessage } from "@/lib/error-message"

/** History of published versions, with one-step restore. */
export function VersionsDialog({ current, onClose }: { current: number; onClose: () => void }) {
  const { t, i18n } = useTranslation("theme")
  const { t: tc } = useTranslation()
  const { data: versions, isLoading, isError } = useVersions()
  const rollback = useRollback()
  const [confirming, setConfirming] = useState<number | null>(null)
  const when = new Intl.DateTimeFormat(i18n.language, { dateStyle: "medium", timeStyle: "short" })

  async function restore(version: number) {
    try {
      await rollback.mutateAsync(version)
      onClose()
    } catch {
      // The error shows below the list.
    }
  }

  return (
    <Dialog open onClose={onClose} title={t("versions.title")} closeLabel={tc("close")}>
      {isLoading && <p className="text-sm text-muted-foreground">{tc("loading")}</p>}
      {isError && <p role="alert" className="text-sm text-destructive">{t("versions.loadError")}</p>}
      {versions && versions.length === 0 && <p className="text-sm text-muted-foreground">{t("versions.empty")}</p>}
      {versions && versions.length > 0 && (
        <ul className="max-h-72 divide-y divide-border overflow-y-auto">
          {versions.map((v) => (
            <li key={v.version} className="flex items-center justify-between gap-3 py-2.5">
              <div>
                <p className="text-sm font-medium text-foreground">
                  {t("versions.version", { number: v.version })}
                  {v.version === current && (
                    <span className="ml-2 rounded bg-secondary px-1.5 py-0.5 text-[11px] font-medium text-foreground">
                      {t("versions.current")}
                    </span>
                  )}
                </p>
                <p className="text-xs text-muted-foreground">{when.format(new Date(v.createdAt))}</p>
              </div>
              {v.version !== current &&
                (confirming === v.version ? (
                  <Button size="sm" onClick={() => void restore(v.version)} disabled={rollback.isPending}>
                    {rollback.isPending ? t("versions.restoring") : t("versions.confirmRestore")}
                  </Button>
                ) : (
                  <Button size="sm" variant="secondary" onClick={() => setConfirming(v.version)}>
                    {t("versions.restore")}
                  </Button>
                ))}
            </li>
          ))}
        </ul>
      )}
      {rollback.isError && <p role="alert" className="text-sm text-destructive">{errorMessage(tc, rollback.error)}</p>}
      <p className="text-xs text-muted-foreground">{t("versions.hint")}</p>
    </Dialog>
  )
}
