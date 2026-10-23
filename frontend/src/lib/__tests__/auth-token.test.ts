import { describe, expect, it } from "vitest"
import { isTokenExpired } from "@/context/AuthContext"

/**
 * Un token sin `exp` NO puede tratarse como válido: el backend exige
 * expiración (jwt.WithExpirationRequired) y aceptarlo en el cliente sería un
 * bypass de la expiración de sesión.
 */
function tokenCon(payload: Record<string, unknown>): string {
  const b64 = (o: unknown) =>
    btoa(JSON.stringify(o)).replace(/\+/g, "-").replace(/\//g, "_").replace(/=+$/, "")
  return `${b64({ alg: "HS256", typ: "JWT" })}.${b64(payload)}.firma`
}

describe("isTokenExpired", () => {
  it("token sin exp se considera expirado", () => {
    expect(isTokenExpired(tokenCon({ sub: "1", email: "a@b.com" }))).toBe(true)
  })

  it("token con exp futuro es válido", () => {
    const exp = Math.floor(Date.now() / 1000) + 3600
    expect(isTokenExpired(tokenCon({ sub: "1", exp }))).toBe(false)
  })

  it("token con exp pasado está expirado", () => {
    const exp = Math.floor(Date.now() / 1000) - 60
    expect(isTokenExpired(tokenCon({ sub: "1", exp }))).toBe(true)
  })

  it("token malformado se considera expirado", () => {
    expect(isTokenExpired("no-es-un-jwt")).toBe(true)
    expect(isTokenExpired("")).toBe(true)
  })

  it("exp no numérico se considera expirado", () => {
    expect(isTokenExpired(tokenCon({ exp: "mañana" }))).toBe(true)
  })
})
