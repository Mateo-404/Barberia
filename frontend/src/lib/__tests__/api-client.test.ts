import { describe, expect, it, vi, beforeEach } from "vitest"
import { api } from "@/lib/api-client"

describe("api-client", () => {
  beforeEach(() => {
    vi.restoreAllMocks()
    sessionStorage.clear()
  })

  it("devuelve JSON en éxito", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue({
      ok: true, status: 200,
      headers: new Headers({ "content-type": "application/json" }),
      text: async () => JSON.stringify({ hello: "world" }),
    }))
    const res = await api<{ hello: string }>("/x")
    expect(res.hello).toBe("world")
  })

  it("devuelve undefined en 204", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue({
      ok: true, status: 204, headers: new Headers(),
      text: async () => "",
    }))
    const res = await api<void>("/x", { method: "DELETE" })
    expect(res).toBeUndefined()
  })

  it("clasifica error HTTP (RFC 9457)", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue({
      ok: false, status: 409,
      headers: new Headers({ "content-type": "application/json" }),
      json: async () => ({ title: "Conflict", detail: "turno tomado", type: "urn:conflict" }),
      text: async () => "",
    }))
    await expect(api("/x")).rejects.toMatchObject({ kind: "http", status: 409, detail: "turno tomado", isConflict: true })
  })

  it("clasifica error de red", async () => {
    vi.stubGlobal("fetch", vi.fn().mockRejectedValue(new TypeError("Failed to fetch")))
    await expect(api("/x")).rejects.toMatchObject({ kind: "network", status: 0, isRetriable: true })
  })

  it("respeta AbortSignal del caller (cancelado)", async () => {
    const ctrl = new AbortController()
    vi.stubGlobal("fetch", vi.fn().mockImplementation((_u: unknown, init: { signal?: AbortSignal }) => {
      return new Promise((_r, reject) => {
        if (init?.signal?.aborted) reject(new DOMException("aborted", "AbortError"))
        init?.signal?.addEventListener("abort", () => reject(new DOMException("aborted", "AbortError")))
      })
    }))
    ctrl.abort()
    await expect(api("/x", { signal: ctrl.signal })).rejects.toMatchObject({ kind: "cancelled" })
  })

  it("envía Authorization cuando hay token", async () => {
    sessionStorage.setItem("auth_token", "tok-123")
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true, status: 200, headers: new Headers({ "content-type": "application/json" }),
      text: async () => JSON.stringify({}),
    })
    vi.stubGlobal("fetch", fetchMock)
    await api("/x")
    const headers = fetchMock.mock.calls[0][1].headers
    expect(headers.Authorization).toBe("Bearer tok-123")
  })

  it("no envía Authorization en endpoints públicos", async () => {
    sessionStorage.setItem("auth_token", "tok-secreto")
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true, status: 200, headers: new Headers({ "content-type": "application/json" }),
      text: async () => JSON.stringify([]),
    })
    vi.stubGlobal("fetch", fetchMock)

    await api("/servicios")
    await api("/turnos/findDateTimes")
    await api("/ping")
    await api("/administradores/login", {
      method: "POST",
      body: JSON.stringify({ email: "a@b.com", contrasenia: "12345678" }),
    })
    await api("/turnos", { method: "POST", body: "{}" })

    for (const llamada of fetchMock.mock.calls) {
      expect(llamada[1].headers.Authorization).toBeUndefined()
    }
  })

  it("envía Authorization en endpoints protegidos", async () => {
    sessionStorage.setItem("auth_token", "tok-123")
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true, status: 200, headers: new Headers({ "content-type": "application/json" }),
      text: async () => JSON.stringify([]),
    })
    vi.stubGlobal("fetch", fetchMock)

    await api("/turnos")
    await api("/turnos/1", { method: "PATCH", body: "{}" })

    for (const llamada of fetchMock.mock.calls) {
      expect(llamada[1].headers.Authorization).toBe("Bearer tok-123")
    }
  })

  it("en 5xx no expone el detalle interno del backend", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue({
      ok: false, status: 500,
      headers: new Headers({ "content-type": "application/problem+json" }),
      json: async () => ({
        title: "Error interno del servidor",
        detail: "pq: relation \"turnos\" does not exist (SQLSTATE 42P01)",
      }),
    }))
    await expect(api("/turnos")).rejects.toMatchObject({
      status: 500,
      detail: expect.not.stringContaining("SQLSTATE"),
    })
    await expect(api("/turnos")).rejects.toMatchObject({
      detail: expect.stringContaining("error inesperado"),
    })
  })

  it("mantiene el detalle del backend en 4xx", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue({
      ok: false, status: 400,
      headers: new Headers({ "content-type": "application/problem+json" }),
      json: async () => ({ title: "Bad Request", detail: "La hora debe estar entre 13:00 y 20:00." }),
    }))
    await expect(api("/turnos", { method: "POST", body: "{}" }))
      .rejects.toMatchObject({ detail: expect.stringContaining("13:00") })
  })
})
