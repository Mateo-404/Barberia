/// <reference types="node" />
import { readdirSync, readFileSync, statSync } from "node:fs"
import path from "node:path"
import { describe, expect, it } from "vitest"
import i18n, { resources } from "@/i18n"

const SRC = path.resolve(import.meta.dirname, "../..")

const exists = (key: string, ns: string) => (i18n.exists as (k: string, o: { ns: string }) => boolean)(key, { ns })

function sourceFiles(dir: string): string[] {
  return readdirSync(dir).flatMap((name: string) => {
    const full = path.join(dir, name)
    if (statSync(full).isDirectory()) return name === "__tests__" || name === "locales" ? [] : sourceFiles(full)
    return /\.(ts|tsx)$/.test(name) && !name.endsWith(".d.ts") ? [full] : []
  })
}

/** Maps each `t` variable of a file to its namespace, from useTranslation calls. */
function translators(code: string): Map<string, string> {
  const result = new Map<string, string>()
  for (const m of code.matchAll(/const \{ t(?:: (\w+))? \} = useTranslation\((?:"(\w+)")?\)/g)) {
    result.set(m[1] ?? "t", m[2] ?? "common")
  }
  return result
}

describe("i18n catalog", () => {
  it("has a Spanish entry for every literal key used with t()", () => {
    const missing: string[] = []
    for (const file of sourceFiles(SRC)) {
      const code = readFileSync(file, "utf8")
      for (const [variable, ns] of translators(code)) {
        for (const m of code.matchAll(new RegExp(`\\b${variable}\\(\\s*"([\\w.]+)"`, "g"))) {
          const key = m[1]
          // Plural keys exist as key_one / key_other.
          if (!exists(key, ns) && !exists(`${key}_other`, ns)) {
            missing.push(`${path.relative(SRC, file)}: ${ns}:${key}`)
          }
        }
      }
    }
    expect(missing).toEqual([])
  })

  it("ships the same namespaces it registers", () => {
    expect(Object.keys(resources.es).sort()).toEqual(["admin", "booking", "common", "errors", "shortcuts", "theme"])
  })
})
