import { createContext, useContext, useState, useCallback, type ReactNode } from "react"
import { useLogin, type LoginResponse } from "@/api/auth"
import type { components } from "@/types/api"

type Admin = components["schemas"]["AdministradorResponseDTO"]

const TOKEN_KEY = "auth_token"
const ADMIN_KEY = "admin"

interface AuthContextType {
  admin: Admin | null
  token: string | null
  isAuthenticated: boolean
  isLoading: boolean
  login: (email: string, contrasenia: string) => Promise<void>
  logout: () => void
}

const AuthContext = createContext<AuthContextType | null>(null)

function isTokenExpired(token: string): boolean {
  try {
    const payload = JSON.parse(atob(token.split(".")[1]))
    if (typeof payload.exp !== "number") return false
    return Date.now() >= payload.exp * 1000
  } catch {
    return true
  }
}

/** Devuelve el token si es válido; si expiró o no existe, limpia la sesión y devuelve null. */
function loadValidToken(): string | null {
  const token = sessionStorage.getItem(TOKEN_KEY)
  if (!token || isTokenExpired(token)) {
    clearSession()
    return null
  }
  return token
}

function loadAdmin(): Admin | null {
  if (!loadValidToken()) return null
  const stored = sessionStorage.getItem(ADMIN_KEY)
  if (!stored) return null
  try {
    return JSON.parse(stored) as Admin
  } catch {
    sessionStorage.removeItem(ADMIN_KEY)
    return null
  }
}

function loadToken(): string | null {
  return loadValidToken()
}

function saveSession(admin: Admin, token: string) {
  sessionStorage.setItem(ADMIN_KEY, JSON.stringify(admin))
  sessionStorage.setItem(TOKEN_KEY, token)
}

function clearSession() {
  sessionStorage.removeItem(ADMIN_KEY)
  sessionStorage.removeItem(TOKEN_KEY)
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [admin, setAdmin] = useState<Admin | null>(loadAdmin)
  const [token, setToken] = useState<string | null>(loadToken)

  const { mutateAsync, isPending } = useLogin()

  const login = useCallback(
    async (email: string, contrasenia: string) => {
      const data: LoginResponse = await mutateAsync({ email, contrasenia })
      const { id, nombre, apellido, email: adminEmail, token: jwt } = data
      const adminData: Admin = { id, nombre, apellido, email: adminEmail }
      setAdmin(adminData)
      setToken(jwt)
      saveSession(adminData, jwt)
    },
    [mutateAsync],
  )

  const logout = useCallback(() => {
    setAdmin(null)
    setToken(null)
    clearSession()
  }, [])

  return (
    <AuthContext.Provider
      value={{
        admin,
        token,
        isAuthenticated: admin !== null,
        isLoading: isPending,
        login,
        logout,
      }}
    >
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth(): AuthContextType {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error("useAuth must be used within AuthProvider")
  return ctx
}
