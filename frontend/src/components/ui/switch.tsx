import { cn } from "@/lib/utils"

/** An on/off setting. Name it with `aria-labelledby` or `aria-label`. */
export function Switch({
  checked,
  onChange,
  disabled,
  ...aria
}: {
  checked: boolean
  onChange: (checked: boolean) => void
  disabled?: boolean
  "aria-labelledby"?: string
  "aria-describedby"?: string
  "aria-label"?: string
}) {
  return (
    <button
      type="button"
      role="switch"
      aria-checked={checked}
      disabled={disabled}
      onClick={() => onChange(!checked)}
      className={cn(
        "relative inline-flex h-6 w-11 shrink-0 items-center rounded-full transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-focus-ring disabled:opacity-50",
        checked ? "bg-primary" : "bg-muted-foreground/40",
      )}
      {...aria}
    >
      <span
        className={cn(
          "inline-block h-5 w-5 rounded-full bg-background shadow transition-transform",
          checked ? "translate-x-[1.4rem]" : "translate-x-0.5",
        )}
      />
    </button>
  )
}
