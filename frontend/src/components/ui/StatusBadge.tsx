import type { ReactNode } from "react"
import { cn } from "@/lib/utils"

type BadgeVariant = "success" | "error" | "warning" | "info"

const variants: Record<BadgeVariant, string> = {
  success: "bg-success text-white",
  error: "bg-destructive text-white",
  warning: "bg-warning text-black",
  info: "bg-primary text-primary-foreground",
}

export function StatusBadge({ variant = "info", children }: { variant?: BadgeVariant; children: ReactNode }) {
  return (
    <span className={cn("inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-semibold", variants[variant])}>
      {children}
    </span>
  )
}
