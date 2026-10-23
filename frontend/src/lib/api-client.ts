/**
 * The single HTTP entry point of the app.
 *
 * - Honors the caller's `AbortSignal` (React Query cancels on unmount) and
 *   combines it with its own timeout, so cancellation and timeout differ.
 * - Errors are classified by `kind`: "http" (server answered), "network"
 *   (fetch rejected), "timeout" and "cancelled".
 * - Every error carries a stable `code` (and `params`); the UI translates it
 *   (see lib/error-message.ts). The server's English `detail` is never shown.
 * - A 401 outside login emits `auth:expired`; AuthContext clears the session
 *   and the query cache and the router navigates. Personal data and tokens
 *   are never logged.
 */

import { readToken } from "@/lib/auth-storage"

const BASE_URL = import.meta.env.VITE_API_BASE_URL ?? "http://localhost:8080"
const REQUEST_TIMEOUT_MS = 15_000

/** API base URL, exported for resources that do not go through `api()`
 *  (theme CSS, previews) and need an absolute URL. */
export const API_BASE_URL = BASE_URL

/** Fired when the backend answers 401 (missing, invalid or expired token). */
export const AUTH_EXPIRED_EVENT = "auth:expired"

export type ApiErrorKind = "http" | "network" | "timeout" | "cancelled"

/** A field-level validation error returned by the API. */
export interface FieldError {
  field: string
  code: string
}

export class ApiError extends Error {
  readonly kind: ApiErrorKind
  readonly status: number
  /** Stable identifier to translate, e.g. "appointment.slot_taken". */
  readonly code: string
  readonly params: Record<string, unknown>
  readonly fieldErrors: FieldError[]
  /** Problem `type`, e.g. "/errors/invalid-token". */
  readonly type?: string

  constructor(
    kind: ApiErrorKind,
    status: number,
    code: string,
    options: { params?: Record<string, unknown>; fieldErrors?: FieldError[]; type?: string } = {},
  ) {
    super(code)
    this.name = "ApiError"
    this.kind = kind
    this.status = status
    this.code = code
    this.params = options.params ?? {}
    this.fieldErrors = options.fieldErrors ?? []
    this.type = options.type
  }

  /** True when an automatic retry makes sense (never on 4xx). */
  get isRetriable(): boolean {
    return this.kind === "network" || this.kind === "timeout" || this.status >= 500
  }

  /** The server rejected the operation as a conflict (e.g. slot taken). */
  get isConflict(): boolean {
    return this.status === 409
  }
}

export type ApiOptions = RequestInit & { timeoutMs?: number }

function buildHeaders(init: ApiOptions): Record<string, string> {
  const headers: Record<string, string> = {}
  // Content-Type only when there is a body.
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

/** Reads an RFC 7807 error body, tolerating empty or non-JSON responses. */
async function readProblem(res: Response): Promise<ApiError> {
  try {
    const body: unknown = await res.json()
    if (body && typeof body === "object") {
      const b = body as Record<string, unknown>
      const errors = Array.isArray(b.errors) ? (b.errors as FieldError[]) : []
      return new ApiError("http", res.status, typeof b.code === "string" ? b.code : "server.unknown", {
        params: b.params && typeof b.params === "object" ? (b.params as Record<string, unknown>) : {},
        fieldErrors: errors.filter((e) => typeof e?.field === "string" && typeof e?.code === "string"),
        type: typeof b.type === "string" ? b.type : undefined,
      })
    }
  } catch {
    // not JSON: fall through
  }
  return new ApiError("http", res.status, res.status >= 500 ? "server.internal" : "server.unknown")
}

const LOGIN_PATH = "/api/auth/login"

/** Routes the backend serves without authentication get no Bearer header
 *  (the token is not leaked to requests that do not need it). */
function isProtected(path: string, method: string): boolean {
  const p = path.split("?")[0]
  if (p === "/healthz" || p === "/readyz" || p === LOGIN_PATH) return false
  if (p.startsWith("/api/public/") || p === "/api/appointments/booked-slots") return false
  if (method === "GET" && (p === "/api/services" || p.startsWith("/api/services/"))) return false
  if (method === "POST" && p === "/api/appointments") return false
  return true
}

export async function api<T>(path: string, init: ApiOptions = {}): Promise<T> {
  const { timeoutMs = REQUEST_TIMEOUT_MS, headers: callerHeaders, signal: callerSignal, ...rest } = init

  const timeoutSignal = AbortSignal.timeout(timeoutMs)
  const signal = callerSignal ? AbortSignal.any([callerSignal, timeoutSignal]) : timeoutSignal

  const headers = buildHeaders({ ...rest, headers: callerHeaders })
  const method = (rest.method ?? "GET").toUpperCase()
  if (isProtected(path, method)) {
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
      throw new ApiError("timeout", 0, "client.timeout")
    }
    if (callerSignal?.aborted) {
      throw new ApiError("cancelled", 0, "client.cancelled")
    }
    throw new ApiError("network", 0, "client.network")
  }

  if (!res.ok) {
    const error = await readProblem(res)
    if (res.status === 401 && path !== LOGIN_PATH) {
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
    throw new ApiError("http", res.status, "client.invalid_response")
  }
}
