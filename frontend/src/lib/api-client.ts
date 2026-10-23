/**
 * Único punto de acceso HTTP de la aplicación.
 *
 * - Respeta el `AbortSignal` del llamador (React Query cancela en unmount) y
 *   lo compone con un timeout propio: cancelación y timeout son distinguibles.
 * - Errores clasificados en `kind`: "http" (respuesta del servidor),
 *   "network" (fetch rechazó), "timeout" y "cancelled".
 * - Ante un 401 fuera del login emite `auth:expired`; AuthContext limpia la
 *   sesión y la caché de React Query y el router se encarga de navegar.
 *   Nunca se loguean datos personales ni tokens.
 */

import { readToken } from "@/lib/auth-storage"

const BASE_URL = import.meta.env.VITE_API_BASE_URL ?? "http://localhost:8080"
const REQUEST_TIMEOUT_MS = 15_000

/** URL base del API. Se exporta para los recursos que no van por `api()`
 *  (CSS del tema, previews) y que necesitan una URL absoluta. */
export const API_BASE_URL = BASE_URL

/** Disparado cuando el backend responde 401 (token ausente/inválido/expirado). */
export const AUTH_EXPIRED_EVENT = "auth:expired"

export type ApiErrorKind = "http" | "network" | "timeout" | "cancelled"

export class ApiError extends Error {
  readonly kind: ApiErrorKind
  readonly status: number
  readonly title: string
  readonly detail: string
  readonly type?: string
  /** Body JSON parseado de la respuesta de error (ej. `details[]` del 422). */
  readonly body?: unknown

  constructor(
    kind: ApiErrorKind,
    status: number,
    title: string,
    detail: string,
    type?: string,
    body?: unknown,
  ) {
    super(detail)
    this.name = "ApiError"
    this.kind = kind
    this.status = status
    this.title = title
    this.detail = detail
    this.type = type
    this.body = body
  }

  /** true si un reintento automático tiene sentido (no en 4xx). */
  get isRetriable(): boolean {
    return this.kind === "network" || this.kind === "timeout" || this.status >= 500
  }

  /** El backend rechazó la operación (ej. turno tomado por otro usuario). */
  get isConflict(): boolean {
    return this.status === 409
  }
}

export type ApiOptions = RequestInit & { timeoutMs?: number }

function buildHeaders(init: ApiOptions): Record<string, string> {
  const headers: Record<string, string> = {}
  // Content-Type sólo cuando hay cuerpo: no ensuciar GET/HEAD.
  if (init.body !== undefined && init.body !== null) {
    headers["Content-Type"] = "application/json"
  }
  if (init.headers && typeof init.headers === "object" && !("forEach" in init.headers)) {
    for (const [k, v] of Object.entries(init.headers as Record<string, string>)) {
      if (typeof v === "string") headers[k] = v
    }
  }
  return headers
}

function notifyAuthExpired() {
  if (typeof window !== "undefined") {
    window.dispatchEvent(new CustomEvent(AUTH_EXPIRED_EVENT))
  }
}

/** Lee el body de error tolerando respuestas no JSON o vacías. */
async function readErrorBody(res: Response): Promise<{
  title: string
  detail: string
  type?: string
  body?: unknown
}> {
  try {
    const body: unknown = await res.json()
    if (body && typeof body === "object") {
      const b = body as Record<string, unknown>
      // En 5xx el detalle interno puede filtrar información (traza, SQL):
      // se muestra siempre el mensaje genérico, no el del servidor.
      const detail =
        res.status >= 500
          ? "Ocurrió un error inesperado. Intentá de nuevo más tarde."
          : typeof b.detail === "string"
            ? b.detail
            : res.statusText || "Error del servidor"
      return {
        title: typeof b.title === "string" ? b.title : "Error",
        detail,
        type: typeof b.type === "string" ? b.type : undefined,
        body,
      }
    }
  } catch {
    // body no JSON: cae al fallback
  }
  return { title: "Error", detail: res.statusText || "Error del servidor" }
}

/**
 * Rutas que el backend expone sin autenticación: no se les manda el Bearer
 * (evita filtrar el token en peticiones que no lo necesitan).
 */
const RUTAS_PUBLICAS = [/^\/ping$/, /^\/healthz$/, /^\/readyz$/, /^\/turnos\/findDateTimes$/]

function esRutaProtegida(path: string, method: string): boolean {
  const soloPath = path.split("?")[0]
  if (RUTAS_PUBLICAS.some((re) => re.test(soloPath))) return false
  if (soloPath === "/servicios" || soloPath.startsWith("/servicios/")) return false
  if (soloPath === "/administradores/login") return false
  if (method === "POST" && soloPath === "/turnos") return false
  return true
}

export async function api<T>(path: string, init: ApiOptions = {}): Promise<T> {
  const { timeoutMs = REQUEST_TIMEOUT_MS, headers: callerHeaders, signal: callerSignal, ...rest } = init

  const timeoutSignal = AbortSignal.timeout(timeoutMs)
  const signal = callerSignal ? AbortSignal.any([callerSignal, timeoutSignal]) : timeoutSignal

  const headers = buildHeaders({ ...rest, headers: callerHeaders })
  const metodo = (rest.method ?? "GET").toUpperCase()
  if (esRutaProtegida(path, metodo)) {
    const token = readToken()
    if (token) headers["Authorization"] = `Bearer ${token}`
  }

  let res: Response
  try {
    res = await fetch(`${BASE_URL}${path}`, {
      credentials: "same-origin",
      ...rest,
      headers,
      signal,
    })
  } catch {
    if (timeoutSignal.aborted) {
      throw new ApiError("timeout", 0, "Tiempo de espera agotado", "La conexión tardó demasiado. Intentá de nuevo.")
    }
    if (callerSignal?.aborted) {
      throw new ApiError("cancelled", 0, "Petición cancelada", "La petición fue cancelada.")
    }
    throw new ApiError("network", 0, "Sin conexión", "No se pudo conectar con el servidor. Revisá tu conexión.")
  }

  if (!res.ok) {
    const { title, detail, type, body } = await readErrorBody(res)
    const error = new ApiError("http", res.status, title, detail, type, body)
    if (res.status === 401 && path !== "/administradores/login") {
      notifyAuthExpired()
    }
    throw error
  }

  if (res.status === 204 || res.status === 205) return undefined as T

  const text = await res.text()
  if (!text) return undefined as T
  try {
    return JSON.parse(text) as T
  } catch {
    throw new ApiError("http", res.status, "Respuesta inválida", "El servidor devolvió una respuesta que no se pudo interpretar.")
  }
}
