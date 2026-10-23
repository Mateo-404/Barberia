import { useEffect, useId, useRef, type ReactNode } from "react"
import { X } from "lucide-react"

const FOCUSABLE =
  'a[href],button:not([disabled]),input:not([disabled]),select:not([disabled]),textarea:not([disabled]),[tabindex]:not([tabindex="-1"])'

interface DialogProps {
  open: boolean
  onClose: () => void
  /** Título visible y nombre accesible del diálogo (aria-labelledby). */
  title: string
  children: ReactNode
  footer?: ReactNode
  /** Label del botón de cierre (default "Cerrar"). */
  closeLabel?: string
}

/**
 * Diálogo accesible: `role="dialog"` + `aria-modal` + `aria-labelledby`,
 * foco inicial dentro del diálogo, Tab atrapado, Escape cierra y el foco
 * vuelve al elemento que lo abrió. Sin dependencias nuevas.
 */
export function Dialog({ open, onClose, title, children, footer, closeLabel = "Cerrar" }: DialogProps) {
  const dialogRef = useRef<HTMLDivElement>(null)
  const restoreRef = useRef<HTMLElement | null>(null)
  const titleId = useId()
  // Estabiliza el handler: el efecto no debe re-ejecutarse (ni robar el foco)
  // cuando el padre re-renderiza.
  const onCloseRef = useRef(onClose)
  useEffect(() => {
    onCloseRef.current = onClose
  })

  useEffect(() => {
    if (!open) return
    restoreRef.current = document.activeElement as HTMLElement | null
    dialogRef.current?.focus()
    if (!open) return
    restoreRef.current = document.activeElement as HTMLElement | null
    dialogRef.current?.focus()

    function onKeyDown(e: KeyboardEvent) {
      if (e.key === "Escape") {
        e.preventDefault()
        onCloseRef.current()
        return
      }
      if (e.key !== "Tab") return
      const dialog = dialogRef.current
      if (!dialog) return
      const nodes = Array.from(dialog.querySelectorAll<HTMLElement>(FOCUSABLE))
      if (nodes.length === 0) {
        e.preventDefault()
        return
      }
      const first = nodes[0]
      const last = nodes[nodes.length - 1]
      const active = document.activeElement
      if (e.shiftKey && (active === first || active === dialog)) {
        e.preventDefault()
        last.focus()
      } else if (!e.shiftKey && active === last) {
        e.preventDefault()
        first.focus()
      }
    }

    document.addEventListener("keydown", onKeyDown)
    return () => {
      document.removeEventListener("keydown", onKeyDown)
      restoreRef.current?.focus()
    }
  }, [open])

  if (!open) return null

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm animate-in fade-in-0"
      onClick={onClose}
    >
      <div
        ref={dialogRef}
        role="dialog"
        aria-modal="true"
        aria-labelledby={titleId}
        tabIndex={-1}
        className="bg-card rounded-2xl border border-border p-6 max-w-sm w-full space-y-4 animate-in fade-in-0 zoom-in-95 duration-200 focus:outline-none"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex items-center justify-between">
          <h3 id={titleId} className="text-lg font-semibold text-foreground">
            {title}
          </h3>
          <button
            type="button"
            onClick={onClose}
            className="p-2 -m-2 rounded-lg text-muted-foreground hover:text-foreground hover:bg-accent"
            aria-label={closeLabel}
          >
            <X className="w-5 h-5" aria-hidden="true" />
          </button>
        </div>
        {children}
        {footer && <div className="flex justify-end gap-2">{footer}</div>}
      </div>
    </div>
  )
}
