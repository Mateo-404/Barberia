import {
  createContext, useContext, useState, useCallback, useEffect, type ReactNode,
} from "react"
import { useNavigate } from "react-router-dom"
import { useQueryClient } from "@tanstack/react-query"
import { useLogin, useTOTPChallenge } from "@/api/auth"
import { AUTH_EXPIRED_EVENT } from "@/lib/api-client"
import {
  clearSession, readAdmin, readToken, writeSession, type Admin,
} from "@/lib/auth-storage"

interface AuthContextType {
  admin: Admin | null
  token: string | null
  isAuthenticated: boolean
  isLoading: boolean
  pending2FA: string | null
  login: (email: string, password: string) => Promise<void>
  verify2FA: (code: string) => Promise<void>
  cancel2FA: () => void
  logout: () => void
}

const AuthContext = createContext<AuthContextType | null>(null)

/** Exported so the token expiry policy can be tested. */
export function isTokenExpired(token: string): boolean {
  try {
    const payload = JSON.parse(atob(token.split(".")[1]))
    // A token without `exp` is never valid: the backend requires expiry.
    if (typeof payload.exp !== "number") return true
    return Date.now() >= payload.exp * 1000
  } catch {
    return true
  }
}

/** Returns the token when present and unexpired; otherwise clears the session. */
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
  const [pending2FA, setPending2FA] = useState<string | null>(null)

  const { mutateAsync, isPending } = useLogin()
  const { mutateAsync: challengeAsync, isPending: isChallengePending } = useTOTPChallenge()

  /** Ends the session and drops ALL remote state, so the next session cannot
   * see the previous one's data. */
  const terminateSession = useCallback(() => {
    setAdmin(null)
    setToken(null)
    setPending2FA(null)
    clearSession()
    queryClient.clear()
  }, [queryClient])

  // The HTTP client reports a 401 (invalid or expired token on any request).
  useEffect(() => {
    const onExpired = () => {
      terminateSession()
      navigate("/login", { replace: true })
    }
    window.addEventListener(AUTH_EXPIRED_EVENT, onExpired)
    return () => window.removeEventListener(AUTH_EXPIRED_EVENT, onExpired)
  }, [terminateSession, navigate])

  const login = useCallback(
    async (email: string, password: string) => {
      const data: any = await mutateAsync({ email, password })
      if (data.requires2fa && data.tempToken) {
        setPending2FA(data.tempToken)
        return
      }
      const { token: jwt, ...adminData } = data
      queryClient.clear()
      setAdmin(adminData)
      setToken(jwt)
      writeSession(adminData, jwt)
    },
    [mutateAsync, queryClient],
  )

  const verify2FA = useCallback(
    async (code: string) => {
      if (!pending2FA) throw new Error("no pending 2fa")
      const data: any = await challengeAsync({ tempToken: pending2FA, code })
      const { token: jwt, ...adminData } = data
      setPending2FA(null)
      queryClient.clear()
      setAdmin(adminData)
      setToken(jwt)
      writeSession(adminData, jwt)
    },
    [pending2FA, challengeAsync, queryClient],
  )

  const cancel2FA = useCallback(() => setPending2FA(null), [])

  const logout = useCallback(() => {
    terminateSession()
  }, [terminateSession])

  return (
    <AuthContext.Provider
      value={{
        admin,
        token,
        isAuthenticated: admin !== null && token !== null,
        isLoading: isPending || isChallengePending,
        pending2FA,
        login,
        verify2FA,
        cancel2FA,
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
