import { describe, expect, it } from "vitest"
import { businessConfig, timeSlots, isValidSlotTime } from "@/lib/business-config"

describe("business-config", () => {
  it("builds the original slot grid", () => {
    expect(timeSlots()).toEqual([
      "13:00", "13:30", "14:00", "14:30", "15:00", "15:30",
      "16:00", "16:30", "17:00", "17:30", "18:00", "18:30",
      "19:00", "19:30",
    ])
  })

  it("validates slots", () => {
    expect(isValidSlotTime("13:30")).toBe(true)
    expect(isValidSlotTime("12:00")).toBe(false)
    expect(isValidSlotTime("20:00")).toBe(false)
  })

  it("exposes configurable branding", () => {
    expect(businessConfig.branding.name).toBe("TH Barber Club")
  })
})
