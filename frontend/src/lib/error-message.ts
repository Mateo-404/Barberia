import type { TFunction } from "i18next"
import { ApiError } from "@/lib/api-client"

type LooseT = (key: string, options?: Record<string, unknown>) => string

/** Translated, user-facing message for any error the UI can catch. */
export function errorMessage(t: TFunction, err: unknown): string {
  const loose = t as unknown as LooseT
  const fallback = loose("unknown", { ns: "errors" })
  if (!(err instanceof ApiError)) return fallback
  return loose(err.code, { ns: "errors", defaultValue: fallback, ...err.params })
}

/** Translated message for one field-level validation code ("required", …). */
export function fieldErrorMessage(t: TFunction, code: string): string {
  const loose = t as unknown as LooseT
  return loose(`field.${code}`, { ns: "errors", defaultValue: loose("field.invalid_value", { ns: "errors" }) })
}

/** Field errors of a 400/422 response keyed by field path ("customerPhone"). */
export function fieldErrors(t: TFunction, err: unknown): Record<string, string> {
  if (!(err instanceof ApiError)) return {}
  return Object.fromEntries(err.fieldErrors.map((e) => [e.field, fieldErrorMessage(t, e.code)]))
}
