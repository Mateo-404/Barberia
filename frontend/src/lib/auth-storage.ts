import type { components } from "@/types/api"

export type Admin = components["schemas"]["AdministradorResponseDTO"]

/**
 * Único punto de lectura/escritura de la sesión en `sessionStorage`.
 * `api-client` sólo conoce el token; `AuthContext` admin + token.
 * Al migrar a cookies HttpOnly (backend Go), este es el único módulo a tocar.
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
