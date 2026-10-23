import { afterEach, describe, expect, it } from "vitest"
import { applyBusiness, businessConfig, isValidSlotTime, timeSlots } from "@/lib/business-config"

const defaults = { ...businessConfig }
afterEach(() => Object.assign(businessConfig, defaults))

describe("business-config", () => {
  it("builds the default slot grid (13:00–20:00, 30 minutes)", () => {
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

  it("follows the business data once applied", () => {
    applyBusiness({ tz: "America/Montevideo", open: "09:00", close: "11:00", slot: 45 })
    expect(businessConfig.timeZone).toBe("America/Montevideo")
    expect(timeSlots()).toEqual(["09:00", "09:45"]) // 10:30 would end after closing
    expect(isValidSlotTime("10:30")).toBe(false)
  })
})
