import { forwardRef, type InputHTMLAttributes } from "react"
import { cn } from "@/lib/utils"

const Input = forwardRef<HTMLInputElement, InputHTMLAttributes<HTMLInputElement>>(
  ({ className, type, ...props }, ref) => {
    return (
      <input
        type={type}
        className={cn(
          "flex w-full rounded-xl border-2 border-primary bg-input px-4 py-2.5 text-sm text-foreground placeholder:text-muted-foreground focus-visible:border-primary/80 focus-visible:outline-none focus-visible:shadow-[0_0_0_3px_var(--focus-ring)] disabled:cursor-not-allowed disabled:opacity-50",
          className,
        )}
        ref={ref}
        {...props}
      />
    )
  },
)
Input.displayName = "Input"

export { Input }
