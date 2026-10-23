import { describe, expect, it, vi, beforeEach } from "vitest"
import { api } from "@/lib/api-client"

const json = (body: unknown, status = 200) => ({
  ok: status < 400,
  status,
  headers: new Headers({ "content-type": "application/json" }),
  json: async () => body,
  text: async () => JSON.stringify(body),
})

describe("api-client", () => {
  beforeEach(() => {
    vi.restoreAllMocks()
    sessionStorage.clear()
  })

  it("returns JSON on success", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(json({ hello: "world" })))
    expect((await api<{ hello: string }>("/x")).hello).toBe("world")
  })

  it("returns undefined on 204", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue({ ok: true, status: 204, headers: new Headers(), text: async () => "" }))
    expect(await api<void>("/x", { method: "DELETE" })).toBeUndefined()
  })

  it("classifies an HTTP problem by its code and params", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(
      json({ type: "/errors/conflict", code: "appointment.slot_taken", params: { startsAt: "2026-10-22T15:00" } }, 409),
    ))
    await expect(api("/x")).rejects.toMatchObject({
      kind: "http",
      status: 409,
      code: "appointment.slot_taken",
      params: { startsAt: "2026-10-22T15:00" },
      isConflict: true,
    })
  })

  it("exposes field errors of a validation problem", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(
      json({ code: "request.invalid_fields", errors: [{ field: "customerPhone", code: "required" }, { nope: 1 }] }, 400),
    ))
    await expect(api("/x")).rejects.toMatchObject({ fieldErrors: [{ field: "customerPhone", code: "required" }] })
  })

  it("falls back to server.unknown / server.internal without a problem body", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue({ ok: false, status: 502, json: async () => { throw new Error("no json") } }))
    await expect(api("/x")).rejects.toMatchObject({ code: "server.internal", isRetriable: true })
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue({ ok: false, status: 418, json: async () => { throw new Error("no json") } }))
    await expect(api("/x")).rejects.toMatchObject({ code: "server.unknown", isRetriable: false })
  })

  it("classifies network errors", async () => {
    vi.stubGlobal("fetch", vi.fn().mockRejectedValue(new TypeError("Failed to fetch")))
    await expect(api("/x")).rejects.toMatchObject({ kind: "network", code: "client.network", status: 0, isRetriable: true })
  })

  it("honors the caller's AbortSignal", async () => {
    const controller = new AbortController()
    vi.stubGlobal("fetch", vi.fn().mockImplementation((_u: unknown, init: { signal?: AbortSignal }) =>
      new Promise((_resolve, reject) => {
        if (init?.signal?.aborted) reject(new DOMException("aborted", "AbortError"))
      }),
    ))
    controller.abort()
    await expect(api("/x", { signal: controller.signal })).rejects.toMatchObject({ kind: "cancelled" })
  })

  describe("Authorization header", () => {
    const stub = () => {
      const fetchMock = vi.fn().mockResolvedValue(json([]))
      vi.stubGlobal("fetch", fetchMock)
      return fetchMock
    }

    it("is sent on protected routes", async () => {
      sessionStorage.setItem("auth_token", "tok-123")
      const fetchMock = stub()
      await api("/api/appointments")
      await api("/api/appointments/1", { method: "PATCH", body: "{}" })
      await api("/api/services", { method: "POST", body: "{}" })
      for (const call of fetchMock.mock.calls) expect(call[1].headers.Authorization).toBe("Bearer tok-123")
    })

    it("is not sent on public routes", async () => {
      sessionStorage.setItem("auth_token", "tok-secret")
      const fetchMock = stub()
      await api("/api/services")
      await api("/api/services/1")
      await api("/api/appointments/booked-slots")
      await api("/api/public/settings")
      await api("/healthz")
      await api("/api/auth/login", { method: "POST", body: "{}" })
      await api("/api/appointments", { method: "POST", body: "{}" })
      for (const call of fetchMock.mock.calls) expect(call[1].headers.Authorization).toBeUndefined()
    })
  })

  it("emits auth:expired on a 401 outside login, but not on login", async () => {
    const onExpired = vi.fn()
    window.addEventListener("auth:expired", onExpired)
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(json({ code: "auth.invalid_token" }, 401)))
    await expect(api("/api/appointments")).rejects.toMatchObject({ status: 401 })
    expect(onExpired).toHaveBeenCalledTimes(1)
    await expect(api("/api/auth/login", { method: "POST", body: "{}" })).rejects.toMatchObject({ status: 401 })
    expect(onExpired).toHaveBeenCalledTimes(1)
    window.removeEventListener("auth:expired", onExpired)
  })
})
