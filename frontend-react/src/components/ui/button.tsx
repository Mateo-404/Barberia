import { forwardRef, type ButtonHTMLAttributes } from "react"
import { cva, type VariantProps } from "class-variance-authority"
import { cn } from "@/lib/utils"

const buttonVariants = cva(
  "inline-flex items-center justify-center whitespace-nowrap font-bold transition-all duration-200 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-focus-ring disabled:pointer-events-none disabled:opacity-50",
  {
    variants: {
      variant: {
        default:
          "bg-primary text-primary-foreground hover:bg-primary/90 rounded-[30px] px-8 py-3",
        secondary:
          "border border-primary text-primary bg-transparent hover:bg-accent rounded-[10px] px-6 py-2 text-sm font-normal",
        destructive:
          "bg-destructive text-white hover:bg-destructive/90 rounded-lg px-6 py-2",
        ghost:
          "hover:bg-accent hover:text-accent-foreground rounded-lg px-4 py-2",
        link:
          "text-primary underline-offset-4 hover:underline",
        icon:
          "h-10 w-10 rounded-lg hover:bg-accent",
      },
      size: {
        default: "",
        sm: "px-4 py-1.5 text-sm",
        lg: "px-10 py-4 text-lg",
        icon: "h-10 w-10",
      },
    },
    defaultVariants: {
      variant: "default",
      size: "default",
    },
  },
)

interface ButtonProps
  extends ButtonHTMLAttributes<HTMLButtonElement>,
    VariantProps<typeof buttonVariants> {}

const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant, size, ...props }, ref) => {
    return (
      <button
        className={cn(buttonVariants({ variant, size, className }))}
        ref={ref}
        {...props}
      />
    )
  },
)
Button.displayName = "Button"

export { Button, buttonVariants }
