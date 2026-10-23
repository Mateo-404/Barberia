/** Safe internal paths for post-login redirects (prevents open redirects). */
export function safeInternalPath(path: unknown): string | null {
  if (typeof path !== "string") return null
  if (!path.startsWith("/") || path.startsWith("//") || path.includes("://")) return null
  return path
}
