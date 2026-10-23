import { useEffect, useId, useState } from "react"

const HEX = /^#[0-9a-fA-F]{6}$/

/**
 * A color as a swatch plus a hex field. The swatch opens the system picker;
 * typing is free-form and only commits when the value is a full #rrggbb.
 */
export function ColorField({
  label,
  value,
  onChange,
  error,
  size = "md",
}: {
  label: string
  value: string
  onChange: (hex: string) => void
  error?: string
  size?: "md" | "sm"
}) {
  const id = useId()
  const [draft, setDraft] = useState(value)
  useEffect(() => setDraft(value), [value])

  const dimension = size === "md" ? "h-11 w-11" : "h-8 w-8"
  return (
    <div className="flex items-center gap-3">
      <span className={`relative shrink-0 overflow-hidden rounded-full ring-1 ring-border ${dimension}`} style={{ backgroundColor: HEX.test(value) ? value : "transparent" }}>
        <input
          type="color"
          aria-label={label}
          value={HEX.test(value) ? value : "#000000"}
          onChange={(e) => onChange(e.target.value)}
          className="absolute inset-0 h-full w-full cursor-pointer opacity-0"
        />
      </span>
      <div className="min-w-0 flex-1">
        <label htmlFor={id} className="sr-only">
          {label}
        </label>
        <input
          id={id}
          value={draft}
          spellCheck={false}
          maxLength={7}
          aria-invalid={error ? true : undefined}
          onChange={(e) => {
            setDraft(e.target.value)
            if (HEX.test(e.target.value)) onChange(e.target.value.toLowerCase())
          }}
          onBlur={() => setDraft(value)}
          className="w-full rounded-md border border-border bg-input px-3 py-2 font-mono text-sm text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-focus-ring"
        />
        {error && (
          <p role="alert" className="mt-1 text-xs text-destructive">
            {error}
          </p>
        )}
      </div>
    </div>
  )
}
