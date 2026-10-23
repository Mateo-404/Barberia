import { useEffect, useRef, useState } from "react"
import { ApiError } from "@/lib/api-client"
import { previewTheme, type PreviewResult } from "@/api/theme"
import type { Theme } from "@/api/settings"
import type { FieldError } from "@/lib/api-client"

const PREVIEW_DELAY_MS = 150

export interface ThemePreview {
  /** Last successful render (kept while a newer draft is invalid). */
  result: PreviewResult | null
  /** Field errors of the current draft; empty when it is valid. */
  errors: FieldError[]
  valid: boolean
}

/**
 * Renders the draft on the server shortly after each change. The server is
 * the only place that knows how a theme turns into colors, so the editor
 * never computes a palette itself.
 */
export function useThemePreview(theme: Theme | null): ThemePreview {
  const [result, setResult] = useState<PreviewResult | null>(null)
  const [errors, setErrors] = useState<FieldError[]>([])
  const [valid, setValid] = useState(true)
  const requestId = useRef(0)

  useEffect(() => {
    if (!theme) return
    const id = ++requestId.current
    const controller = new AbortController()
    const timer = setTimeout(() => {
      previewTheme(theme, controller.signal)
        .then((next) => {
          if (id !== requestId.current) return
          setResult(next)
          setErrors([])
          setValid(true)
        })
        .catch((err: unknown) => {
          if (id !== requestId.current || !(err instanceof ApiError) || err.fieldErrors.length === 0) return
          setErrors(err.fieldErrors)
          setValid(false)
        })
    }, PREVIEW_DELAY_MS)
    return () => {
      clearTimeout(timer)
      controller.abort()
    }
  }, [theme])

  return { result, errors, valid }
}
