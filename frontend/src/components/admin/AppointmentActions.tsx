import { ArrowRight, Pencil, Trash2 } from "lucide-react"
import { useTranslation } from "react-i18next"
import type { Appointment } from "@/api/appointments"

const ICON_BUTTON = "min-h-10 min-w-10 p-2.5 rounded-lg text-muted-foreground transition-colors active:scale-95"

/** Edit / move / delete buttons of an appointment row. */
export function AppointmentActions({
  appointment,
  onEdit,
  onMove,
  onDelete,
}: {
  appointment: Appointment
  onEdit?: (a: Appointment) => void
  onMove?: (a: Appointment) => void
  onDelete?: (id: number) => void
}) {
  const { t } = useTranslation("admin")
  const who = `${appointment.customerFirstName ?? ""} ${appointment.customerLastName ?? ""}`.trim()
  return (
    <div className="flex items-center gap-1">
      {onEdit && (
        <button
          onClick={() => onEdit(appointment)}
          className={`${ICON_BUTTON} hover:text-foreground hover:bg-accent`}
          aria-label={t("appointments.editFor", { name: who })}
        >
          <Pencil className="w-4 h-4" />
        </button>
      )}
      {onMove && (
        <button
          onClick={() => onMove(appointment)}
          className={`${ICON_BUTTON} hover:text-primary hover:bg-accent`}
          aria-label={t("appointments.moveFor", { name: who })}
          title={t("appointments.moveHint", { name: who })}
        >
          <ArrowRight className="w-4 h-4" />
        </button>
      )}
      {onDelete && (
        <button
          onClick={() => onDelete(appointment.id)}
          className={`${ICON_BUTTON} hover:text-destructive hover:bg-destructive/10`}
          aria-label={t("appointments.deleteFor", { name: who })}
        >
          <Trash2 className="w-4 h-4" />
        </button>
      )}
    </div>
  )
}
