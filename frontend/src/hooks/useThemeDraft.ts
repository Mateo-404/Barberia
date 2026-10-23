import { useCallback, useEffect, useMemo, useRef, useState } from "react"
import { useSaveDraft, type Theme } from "@/api/settings"

export type SaveStatus = "saved" | "dirty" | "saving" | "error"

const AUTOSAVE_DELAY_MS = 1000

/**
 * Local copy of the theme draft being edited, saved to the server by itself
 * shortly after the last change. `setCanSave(false)` holds the save back
 * while the draft is invalid (it would only be rejected).
 */
export function useThemeDraft(initial: Theme | undefined) {
  const [canSave, setCanSave] = useState(true)
  const [theme, setTheme] = useState<Theme | null>(null)
  const [status, setStatus] = useState<SaveStatus>("saved")
  const saved = useRef("")
  const latest = useRef<Theme | null>(null)
  const timer = useRef<ReturnType<typeof setTimeout> | null>(null)
  const { mutateAsync, isPending } = useSaveDraft()

  // Take the server's draft once, when it first arrives.
  useEffect(() => {
    if (initial && latest.current === null) {
      latest.current = initial
      saved.current = JSON.stringify(initial)
      setTheme(initial)
    }
  }, [initial])

  const save = useCallback(async (): Promise<boolean> => {
    const current = latest.current
    if (!current || JSON.stringify(current) === saved.current) return true
    setStatus("saving")
    try {
      await mutateAsync(current)
      saved.current = JSON.stringify(current)
      setStatus(JSON.stringify(latest.current) === saved.current ? "saved" : "dirty")
      return true
    } catch {
      setStatus("error")
      return false
    }
  }, [mutateAsync])

  const update = useCallback((change: (t: Theme) => Theme) => {
    const next = change(latest.current as Theme)
    latest.current = next
    setTheme(next)
    setStatus(JSON.stringify(next) === saved.current ? "saved" : "dirty")
  }, [])

  // Autosave: wait for a pause in editing.
  useEffect(() => {
    if (!theme || !canSave || JSON.stringify(theme) === saved.current) return
    timer.current = setTimeout(() => void save(), AUTOSAVE_DELAY_MS)
    return () => {
      if (timer.current) clearTimeout(timer.current)
    }
  }, [theme, canSave, save])

  // Leaving the page with unsaved edits asks for confirmation.
  useEffect(() => {
    const onBeforeUnload = (event: BeforeUnloadEvent) => {
      if (JSON.stringify(latest.current) !== saved.current) event.preventDefault()
    }
    window.addEventListener("beforeunload", onBeforeUnload)
    return () => window.removeEventListener("beforeunload", onBeforeUnload)
  }, [])

  /** Saves right now (before publishing or leaving). */
  const flush = useCallback(async () => {
    if (timer.current) clearTimeout(timer.current)
    return save()
  }, [save])

  return useMemo(
    () => ({ theme, update, status: isPending ? ("saving" as const) : status, flush, setCanSave }),
    [theme, update, status, isPending, flush],
  )
}
