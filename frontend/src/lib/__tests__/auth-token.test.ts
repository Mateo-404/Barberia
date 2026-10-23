import { describe, expect, it } from "vitest"
import { isTokenExpired } from "@/context/AuthContext"

/**
 * A token without `exp` is never valid: the backend requires expiry
 * (jwt.WithExpirationRequired) and accepting it client-side would bypass
 * session expiry.
 */
function tokenWith(payload: Record<string, unknown>): string {
  const b64 = (o: unknown) =>
    btoa(JSON.stringify(o)).replace(/\+/g, "-").replace(/\//g, "_").replace(/=+$/, "")
  return `${b64({ alg: "HS256", typ: "JWT" })}.${b64(payload)}.firma`
}

describe("isTokenExpired", () => {
  it("a token without exp is expired", () => {
    expect(isTokenExpired(tokenWith({ sub: "1", email: "a@b.com" }))).toBe(true)
  })

  it("a token with a future exp is valid", () => {
    const exp = Math.floor(Date.now() / 1000) + 3600
    expect(isTokenExpired(tokenWith({ sub: "1", exp }))).toBe(false)
  })

  it("a token with a past exp is expired", () => {
    const exp = Math.floor(Date.now() / 1000) - 60
    expect(isTokenExpired(tokenWith({ sub: "1", exp }))).toBe(true)
  })

  it("a malformed token is expired", () => {
    expect(isTokenExpired("not-a-jwt")).toBe(true)
    expect(isTokenExpired("")).toBe(true)
  })

  it("a non-numeric exp is expired", () => {
    expect(isTokenExpired(tokenWith({ exp: "tomorrow" }))).toBe(true)
  })
})
