import type { TFunction } from "i18next"
import { ApiError } from "@/lib/api-client"
import { loose } from "@/lib/i18n-loose"

/** Translated, user-facing message for any error the UI can catch. */
export function errorMessage(t: TFunction, err: unknown): string {
    const fallback = loose(t)("unknown", { ns: "errors" })
  if (!(err instanceof ApiError)) return fallback
  return loose(t)(err.code, { ns: "errors", defaultValue: fallback, ...err.params })
}

/** Translated message for one field-level validation code ("required", …). */
export function fieldErrorMessage(t: TFunction, code: string): string {
    return loose(t)(`field.${code}`, { ns: "errors", defaultValue: loose(t)("field.invalid_value", { ns: "errors" }) })
}

/** Field errors of a 400/422 response keyed by field path ("customerPhone"). */
export function fieldErrors(t: TFunction, err: unknown): Record<string, string> {
  if (!(err instanceof ApiError)) return {}
  return Object.fromEntries(err.fieldErrors.map((e) => [e.field, fieldErrorMessage(t, e.code)]))
}
