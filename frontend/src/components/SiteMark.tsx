import { useSiteIdentity } from "@/api/site"

/** The business's logo, or its name set in the heading font when it has none. */
export function SiteMark({ imageClassName = "h-14 max-w-48", className = "" }: { imageClassName?: string; className?: string }) {
  const { name, logo } = useSiteIdentity()
  if (logo) return <img src={logo} alt={name} className={`object-contain ${imageClassName} ${className}`} />
  return (
    <span className={`font-[family-name:var(--font-heading)] text-xl font-bold text-foreground ${className}`}>{name}</span>
  )
}
