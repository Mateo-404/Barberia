import { describe, expect, it } from "vitest"
import { applyBranding, applyDefaultBranding } from "@/lib/theme"

describe("theme", () => {
  it("aplica tokens por defecto", () => {
    applyDefaultBranding()
    expect(document.documentElement.style.getPropertyValue("--primary")).toBe("#d4a574")
  })

  it("sobrescribe tokens parcialmente", () => {
    applyBranding({ primary: "#ff0000" })
    expect(document.documentElement.style.getPropertyValue("--primary")).toBe("#ff0000")
    expect(document.documentElement.style.getPropertyValue("--background")).toBe("#1c1c1c")
  })
})
