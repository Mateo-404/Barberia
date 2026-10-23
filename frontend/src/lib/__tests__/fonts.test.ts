/// <reference types="node" />
import { readFileSync } from "node:fs"
import path from "node:path"
import { describe, expect, it } from "vitest"
import { FONT_NAMES, fontStack, firstFamily } from "@/lib/fonts"
import { FONT_PAIRS } from "@/lib/font-pairs"

// The backend validates fonts against its own list: the two must match.
const backendSource = readFileSync(path.resolve(import.meta.dirname, "../../../../backend-go/internal/settings/fonts.go"), "utf8")
const backendFonts = [...backendSource.matchAll(/^\t"([^"]+)":\s+"(?:serif|sans-serif)",/gm)].map((m) => m[1])

describe("fonts", () => {
  it("ships exactly the fonts the backend allows", () => {
    expect([...FONT_NAMES].sort()).toEqual([...backendFonts].sort())
  })

  it("only offers pairs made of shipped fonts", () => {
    for (const pair of FONT_PAIRS) {
      expect(FONT_NAMES).toContain(pair.heading)
      expect(FONT_NAMES).toContain(pair.body)
    }
  })

  it("builds a stack with a generic fallback and reads the first family back", () => {
    expect(fontStack("Playfair Display")).toBe('"Playfair Display",serif')
    expect(fontStack("Inter")).toBe('"Inter",sans-serif')
    expect(firstFamily('"Playfair Display",serif')).toBe("Playfair Display")
  })
})
