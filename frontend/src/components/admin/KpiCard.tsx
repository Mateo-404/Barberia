import { useEffect, useRef, useState } from "react"

type KpiTone =
  | "blue"
  | "green"
  | "ocre"
  | "cyan"
  | "yellow"
  | "red"

const TONES: Record<KpiTone, { bg: string; badge: string }> = {
  blue: { bg: "#1E3A5F", badge: "#2563eb" },
  green: { bg: "#1E5631", badge: "#22c55e" },
  ocre: { bg: "#3D3419", badge: "#eab308" },
  cyan: { bg: "#133B44", badge: "#06b6d4" },
  yellow: { bg: "#4A3F15", badge: "#facc15" },
  red: { bg: "#5F1E1E", badge: "#ef4444" },
}

function useCountUp(target: number, duration = 600) {
  const [display, setDisplay] = useState(0)
  const raf = useRef<number>(0)
  const prefersReduced = useRef(
    typeof window !== "undefined" &&
      window.matchMedia("(prefers-reduced-motion: reduce)").matches,
  )

  useEffect(() => {
    if (prefersReduced.current) {
      setDisplay(target)
      return
    }
    const start = performance.now()
    const from = 0
    function tick(now: number) {
      const progress = Math.min((now - start) / duration, 1)
      const eased = 1 - Math.pow(1 - progress, 3)
      setDisplay(Math.round(from + (target - from) * eased))
      if (progress < 1) raf.current = requestAnimationFrame(tick)
    }
    raf.current = requestAnimationFrame(tick)
    return () => cancelAnimationFrame(raf.current)
  }, [target, duration])

  return display
}

export function KpiCard({
  label,
  value,
  hint,
  tone = "blue",
  animated = false,
}: {
  label: string
  value: string
  hint?: string
  tone?: KpiTone
  animated?: boolean
}) {
  const t = TONES[tone]

  const numericValue = animated ? parseInt(value.replace(/\D/g, ""), 10) : 0
  const count = useCountUp(animated ? numericValue : 0)
  const formattedCount = `${value.replace(/[\d.,\s]+/g, "")}${count.toLocaleString("es-AR")}`
  const animatable = animated && !isNaN(numericValue)

  return (
    <div
      className="rounded-2xl p-5 border border-border transition-transform duration-200 hover:scale-[1.02]"
      style={{ backgroundColor: t.bg }}
    >
      <div
        className="inline-flex items-center gap-2 rounded-full px-3 py-1 mb-3"
        style={{ backgroundColor: "rgba(0,0,0,0.35)" }}
      >
        <span
          className="w-2 h-2 rounded-full"
          style={{ backgroundColor: t.badge }}
        />
        <span className="text-xs text-white/80 uppercase tracking-wide font-medium">
          {label}
        </span>
      </div>
      <p className="text-2xl font-bold text-white">
        {animatable ? formattedCount : value}
      </p>
      {hint && (
        <p className="text-xs mt-1 text-white/70">{hint}</p>
      )}
    </div>
  )
}
