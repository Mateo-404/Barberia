const BASE_URL = import.meta.env.VITE_API_BASE_URL ?? "http://localhost:8080"
const TOKEN_KEY = "auth_token"
const REQUEST_TIMEOUT_MS = 15_000

export class ApiError extends Error {
  status: number
  title: string
  detail: string
  type?: string

  constructor(status: number, title: string, detail: string, type?: string) {
    super(detail)
    this.name = "ApiError"
    this.status = status
    this.title = title
    this.detail = detail
    this.type = type
  }
}

function getToken(): string | null {
  return sessionStorage.getItem(TOKEN_KEY)
}

function clearSession() {
  sessionStorage.removeItem(TOKEN_KEY)
  sessionStorage.removeItem("admin")
}

export async function api<T>(path: string, init?: RequestInit): Promise<T> {
  const token = getToken()
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...(init?.headers as Record<string, string> | undefined),
  }
  if (token) {
    headers["Authorization"] = `Bearer ${token}`
  }

  const controller = new AbortController()
  const timeout = setTimeout(() => controller.abort(), REQUEST_TIMEOUT_MS)
  let res: Response
  try {
    res = await fetch(`${BASE_URL}${path}`, {
      ...init,
      headers,
      signal: controller.signal,
    })
  } catch (err) {
    clearTimeout(timeout)
    if (err instanceof DOMException && err.name === "AbortError") {
      throw new ApiError(0, "Tiempo de espera agotado", "La conexión tardó demasiado. Intentá de nuevo.")
    }
    throw err
  }
  clearTimeout(timeout)

  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    const error = new ApiError(
      res.status,
      body.title ?? "Error",
      body.detail ?? res.statusText,
      body.type,
    )
    if (res.status === 401 && path !== "/administradores/login") {
      clearSession()
      window.location.href = "/login"
    }
    throw error
  }

  if (res.status === 204) return undefined as T

  return res.json()
}
