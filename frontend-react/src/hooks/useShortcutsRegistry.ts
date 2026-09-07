import { useEffect, useState } from "react"

export type ShortcutDef = {
  key: string
  description: string
}

type Entry = ShortcutDef & { scope: string; uid: number }

const listeners = new Set<() => void>()
let entries: Entry[] = []
let uidSeq = 0

function emitAll() {
  for (const l of listeners) l()
}

export function useShortcutsRegistry(): Entry[] {
  const [, force] = useState(0)
  useEffect(() => {
    const fn = () => force((n) => n + 1)
    listeners.add(fn)
    return () => {
      listeners.delete(fn)
    }
  }, [])
  return entries
}

// Registra atajos vigentes mientras el componente está montado.
export function useRegisterShortcuts(scope: string, defs: ShortcutDef[]) {
  useEffect(() => {
    const uid = ++uidSeq
    const batch: Entry[] = defs.map((d) => ({ ...d, scope, uid }))
    entries = [...entries, ...batch]
    emitAll()
    return () => {
      entries = entries.filter((e) => e.uid !== uid)
      emitAll()
    }
  }, [scope, defs])
}