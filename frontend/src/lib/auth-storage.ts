import type { components } from "@/types/api"

export type Admin = components["schemas"]["Admin"]

/**
 * The only place that reads or writes the session in `sessionStorage`.
 * `api-client` only knows the token; `AuthContext` the admin + token.
 * Moving to HttpOnly cookies would only touch this module.
 */
const TOKEN_KEY = "auth_token"
const ADMIN_KEY = "admin"

export function readToken(): string | null {
  return sessionStorage.getItem(TOKEN_KEY)
}

export function readAdmin(): Admin | null {
  const stored = sessionStorage.getItem(ADMIN_KEY)
  if (!stored) return null
  try {
    return JSON.parse(stored) as Admin
  } catch {
    sessionStorage.removeItem(ADMIN_KEY)
    return null
  }
}

export function writeSession(admin: Admin, token: string): void {
  sessionStorage.setItem(ADMIN_KEY, JSON.stringify(admin))
  sessionStorage.setItem(TOKEN_KEY, token)
}

export function clearSession(): void {
  sessionStorage.removeItem(ADMIN_KEY)
  sessionStorage.removeItem(TOKEN_KEY)
}
