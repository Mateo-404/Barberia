const BASE_URL = import.meta.env.VITE_API_BASE_URL ?? "http://localhost:8080"

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

export async function api<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE_URL}${path}`, {
    headers: { "Content-Type": "application/json", ...init?.headers },
    ...init,
  })

  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    throw new ApiError(
      res.status,
      body.title ?? "Error",
      body.detail ?? res.statusText,
      body.type,
    )
  }

  if (res.status === 204) return undefined as T

  return res.json()
}
