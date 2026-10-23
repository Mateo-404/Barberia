import { useTranslation } from "react-i18next"
import { Button } from "@/components/ui/button"
import { Dialog } from "@/components/ui/dialog"
import { fieldErrorMessage } from "@/lib/error-message"
import type { FieldError } from "@/lib/api-client"

export function PublishDialog({
  warnings,
  publishing,
  failed,
  onCancel,
  onConfirm,
}: {
  warnings: FieldError[]
  publishing: boolean
  failed: boolean
  onCancel: () => void
  onConfirm: () => void
}) {
  const { t } = useTranslation("theme")
  const { t: tc } = useTranslation()
  return (
    <Dialog
      open
      onClose={onCancel}
      title={t("publish.title")}
      closeLabel={tc("close")}
      footer={
        <>
          <Button variant="secondary" onClick={onCancel}>
            {tc("cancel")}
          </Button>
          <Button onClick={onConfirm} disabled={publishing}>
            {publishing ? t("publish.publishing") : t("publish.confirm")}
          </Button>
        </>
      }
    >
      <p className="text-sm text-muted-foreground">{t("publish.body")}</p>
      {warnings.length > 0 && (
        <div className="space-y-1 rounded-lg bg-warning/15 p-3 text-sm text-foreground">
          <p className="font-medium">{t("publish.warnings")}</p>
          <ul className="list-disc pl-5 text-muted-foreground">
            {warnings.map((w) => (
              <li key={`${w.field}-${w.code}`}>{fieldErrorMessage(tc, w.code)}</li>
            ))}
          </ul>
        </div>
      )}
      {failed && (
        <p role="alert" className="text-sm text-destructive">
          {t("publish.failed")}
        </p>
      )}
    </Dialog>
  )
}
