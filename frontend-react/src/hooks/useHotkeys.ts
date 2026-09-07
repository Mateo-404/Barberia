import { useEffect, useCallback } from "react"

type HotkeyMap = Record<string, () => void>

type RawMods = { ctrl: boolean; alt: boolean; shift: boolean; meta: boolean }

function parseSpec(spec: string): { key: string; mods: RawMods } {
  const mods: RawMods = { ctrl: false, alt: false, shift: false, meta: false }
  let key = ""
  for (const part of spec.split("+")) {
    const q = part.trim().toLowerCase()
    if (q === "ctrl" || q === "control") mods.ctrl = true
    else if (q === "alt") mods.alt = true
    else if (q === "shift") mods.shift = true
    else if (q === "meta") mods.meta = true
    else key = part.trim()
  }
  return { key, mods }
}

function keyMatches(key: string, e: KeyboardEvent): boolean {
  return key.length === 1
    ? e.key.toLowerCase() === key.toLowerCase()
    : e.key === key
}

// Combo explícito ("Ctrl+k") => matcher estricto sobre modificadores;
// tecla simple => matcher clásico (sin ctrl/alt/meta). "?" relaja Shift.
function makeMatcher(spec: string) {
  const { key, mods } = parseSpec(spec)
  const hasMods =
    mods.ctrl || mods.alt || mods.shift || mods.meta || key === "?"
  const relaxedShift = key === "?"
  return (e: KeyboardEvent): { hit: boolean; modded: boolean } => {
    if (!keyMatches(key, e)) return { hit: false, modded: hasMods }
    if (mods.ctrl !== e.ctrlKey) return { hit: false, modded: hasMods }
    if (mods.alt !== e.altKey) return { hit: false, modded: hasMods }
    if (mods.meta !== e.metaKey) return { hit: false, modded: hasMods }
    if (mods.shift && !e.shiftKey) return { hit: false, modded: hasMods }
    if (!mods.shift && !relaxedShift && e.shiftKey)
      return { hit: false, modded: hasMods }
    return { hit: true, modded: hasMods }
  }
}

function isInteractionTarget(el: EventTarget | null): boolean {
  if (!el || !(el instanceof HTMLElement)) return false
  const tag = el.tagName
  if (tag === "INPUT" || tag === "TEXTAREA" || tag === "SELECT") return true
  if (el.isContentEditable) return true
  if (el.closest("dialog") || el.closest("[role='dialog']")) return true
  return false
}

function dialogAbierto(): boolean {
  return document.querySelector("[role='dialog'],[aria-modal='true']") !== null
}

export function useHotkeys(
  map: HotkeyMap,
  enabled = true,
  options: { ignoreWhenDialogOpen?: boolean } = {},
) {
  const handler = useCallback(
    (e: KeyboardEvent) => {
      if (!enabled) return
      if (options.ignoreWhenDialogOpen && dialogAbierto()) return

      for (const spec of Object.keys(map)) {
        const { hit, modded } = makeMatcher(spec)(e)
        if (!hit) continue
        if (!modded && isInteractionTarget(e.target)) return
        e.preventDefault()
        map[spec]()
        return
      }
    },
    [map, enabled, options],
  )

  useEffect(() => {
    window.addEventListener("keydown", handler)
    return () => window.removeEventListener("keydown", handler)
  }, [handler])
}