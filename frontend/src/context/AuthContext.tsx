import {
  createContext, useContext, useState, useCallback, useEffect, type ReactNode,
} from "react"
import { useNavigate } from "react-router-dom"
import { useQueryClient } from "@tanstack/react-query"
import { useLogin, type LoginResponse } from "@/api/auth"
import { AUTH_EXPIRED_EVENT } from "@/lib/api-client"
import {
  clearSession, readAdmin, readToken, writeSession, type Admin,
} from "@/lib/auth-storage"

interface AuthContextType {
  admin: Admin | null
  token: string | null
  isAuthenticated: boolean
  isLoading: boolean
  login: (email: string, contrasenia: string) => Promise<void>
  logout: () => void
}

const AuthContext = createContext<AuthContextType | null>(null)

/** Exportado para poder testear la política de expiración del token. */
export function isTokenExpired(token: string): boolean {
  try {
    const payload = JSON.parse(atob(token.split(".")[1]))
    // Un token sin `exp` no puede considerarse válido: el backend exige
    // expiración (WithExpirationRequired), así que aceptarlo sería un bypass.
    if (typeof payload.exp !== "number") return true
    return Date.now() >= payload.exp * 1000
  } catch {
    return true
  }
}

/** Devuelve el token si existe y no expiró; si no, limpia la sesión. */
function loadValidToken(): string | null {
  const token = readToken()
  if (!token || isTokenExpired(token)) {
    clearSession()
    return null
  }
  return token
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const navigate = useNavigate()
  const queryClient = useQueryClient()

  const [token, setToken] = useState<string | null>(loadValidToken)
  const [admin, setAdmin] = useState<Admin | null>(() => (readToken() ? readAdmin() : null))

  const { mutateAsync, isPending } = useLogin()

  /** Cierra sesión y descarta TODO el estado remoto: la siguiente sesión no
   * puede ver datos de la anterior (turnos, clientes, estadísticas). */
  const terminateSession = useCallback(() => {
    setAdmin(null)
    setToken(null)
    clearSession()
    queryClient.clear()
  }, [queryClient])

  // El cliente HTTP avisa de un 401 (token inválido/expirado en cualquier request).
  useEffect(() => {
    const onExpired = () => {
      terminateSession()
      navigate("/login", { replace: true })
    }
    window.addEventListener(AUTH_EXPIRED_EVENT, onExpired)
    return () => window.removeEventListener(AUTH_EXPIRED_EVENT, onExpired)
  }, [terminateSession, navigate])

  const login = useCallback(
    async (email: string, contrasenia: string) => {
      const data: LoginResponse = await mutateAsync({ email, contrasenia })
      const { id, nombre, apellido, email: adminEmail, token: jwt } = data
      const adminData: Admin = { id, nombre, apellido, email: adminEmail }
      // Una sesión nueva arranca sin restos de la anterior.
      queryClient.clear()
      setAdmin(adminData)
      setToken(jwt)
      writeSession(adminData, jwt)
    },
    [mutateAsync, queryClient],
  )

  const logout = useCallback(() => {
    terminateSession()
  }, [terminateSession])

  return (
    <AuthContext.Provider
      value={{
        admin,
        token,
        isAuthenticated: admin !== null && token !== null,
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
