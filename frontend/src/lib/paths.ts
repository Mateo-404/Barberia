/** Rutas internas seguras para redirecciones post-login (evita open redirect). */
export function safeInternalPath(path: unknown): string | null {
  if (typeof path !== "string") return null
  if (!path.startsWith("/") || path.startsWith("//") || path.includes("://")) return null
  return path
}
