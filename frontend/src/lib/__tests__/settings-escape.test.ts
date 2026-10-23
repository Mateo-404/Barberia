import { describe, expect, it } from "vitest"
import { escapeHtml } from "@/lib/settings"
import { buildPreviewHtml } from "@/lib/settings-preview"
import { DEFAULT_SETTINGS } from "@/lib/settings"

describe("escapeHtml (F-04)", () => {
  it("escapa <script> y comillas", () => {
    const out = escapeHtml('<script>alert(1)</script>"\'&')
    expect(out).not.toContain("<script>")
    expect(out).toContain("&lt;script&gt;")
    expect(out).toContain("&quot;")
    expect(out).toContain("&#39;")
    expect(out).toContain("&amp;")
  })

  it("buildPreviewHtml no deja crudo el nombre con script", () => {
    const s = { ...DEFAULT_SETTINGS, name: "<script>alert(1)</script>" }
    const html = buildPreviewHtml(s)
    expect(html).not.toContain("<script>alert(1)</script>")
    expect(html).toContain("&lt;script&gt;")
  })
})
