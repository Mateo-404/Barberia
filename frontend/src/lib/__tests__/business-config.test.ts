import { describe, expect, it } from "vitest"
import { businessConfig, timeSlots, isValidSlotTime } from "@/lib/business-config"

describe("business-config", () => {
  it("genera la grilla original de horarios", () => {
    expect(timeSlots()).toEqual([
      "13:00", "13:30", "14:00", "14:30", "15:00", "15:30",
      "16:00", "16:30", "17:00", "17:30", "18:00", "18:30",
      "19:00", "19:30",
    ])
  })

  it("valida slots", () => {
    expect(isValidSlotTime("13:30")).toBe(true)
    expect(isValidSlotTime("12:00")).toBe(false)
    expect(isValidSlotTime("20:00")).toBe(false)
  })

  it("expone branding configurable", () => {
    expect(businessConfig.branding.nombre).toBe("TH Barber Club")
  })
})
