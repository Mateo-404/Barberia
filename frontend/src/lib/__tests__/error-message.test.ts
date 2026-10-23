import { describe, expect, it } from "vitest"
import i18n from "@/i18n"
import { ApiError } from "@/lib/api-client"
import { errorMessage, fieldErrorMessage, fieldErrors } from "@/lib/error-message"

const t = i18n.t.bind(i18n)

describe("error messages", () => {
  it("translates a backend code with its params", () => {
    const err = new ApiError("http", 400, "appointment.outside_hours", { params: { open: "13:00", close: "20:00" } })
    expect(errorMessage(t, err)).toBe("El horario debe estar entre las 13:00 y las 20:00.")
  })

  it("falls back to a generic message for unknown codes and non-API errors", () => {
    expect(errorMessage(t, new ApiError("http", 500, "something.new"))).toMatch(/inesperado/)
    expect(errorMessage(t, new Error("boom"))).toMatch(/inesperado/)
  })

  it("translates field codes and maps field errors by path", () => {
    expect(fieldErrorMessage(t, "required")).toBe("Este campo es obligatorio.")
    expect(fieldErrorMessage(t, "never_heard_of_it")).toBe("El valor no es válido.")
    const err = new ApiError("http", 400, "request.invalid_fields", {
      fieldErrors: [{ field: "customerEmail", code: "invalid_email" }],
    })
    expect(fieldErrors(t, err)).toEqual({ customerEmail: "Ingresá un email válido." })
  })
})
