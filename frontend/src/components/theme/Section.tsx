import { ChevronDown } from "lucide-react"
import { useId, type ReactNode } from "react"

/**
 * One collapsible group of controls. The closed state still tells the story:
 * its summary shows the current choice ("Playfair Display + Inter") so the
 * owner can scan the whole theme without opening anything.
 */
export function Section({
  title,
  summary,
  open,
  onToggle,
  children,
}: {
  title: string
  summary: ReactNode
  open: boolean
  onToggle: () => void
  children: ReactNode
}) {
  const panelId = useId()
  return (
    <section className="border-b border-border">
      <h2>
        <button
          type="button"
          onClick={onToggle}
          aria-expanded={open}
          aria-controls={panelId}
          className="flex w-full items-center gap-3 px-5 py-4 text-left transition-colors hover:bg-accent focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-focus-ring"
        >
          <span className="text-sm font-semibold text-foreground">{title}</span>
          <span className="ml-auto min-w-0 truncate text-xs text-muted-foreground">{summary}</span>
          <ChevronDown className={`h-4 w-4 shrink-0 text-muted-foreground transition-transform ${open ? "rotate-180" : ""}`} aria-hidden />
        </button>
      </h2>
      {open && (
        <div id={panelId} className="space-y-5 px-5 pb-6 pt-1">
          {children}
        </div>
      )}
    </section>
  )
}

export function FieldMessage({ message }: { message?: string }) {
  if (!message) return null
  return (
    <p role="alert" className="text-xs text-destructive">
      {message}
    </p>
  )
}
