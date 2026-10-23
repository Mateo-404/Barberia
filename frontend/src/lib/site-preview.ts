import { useSyncExternalStore } from "react"

/**
 * The theme editor shows the real booking page in an iframe
 * (`/?preview=1`) and streams unsaved changes to it. This module holds what
 * the page needs to know about that mode.
 */

export interface SiteIdentity {
  name: string
  tagline: string
  logo: string
}

/** True when this page was opened by the theme editor. */
export const isPreview = (): boolean => new URLSearchParams(window.location.search).get("preview") === "1"

let override: SiteIdentity | null = null
const listeners = new Set<() => void>()

export function setPreviewIdentity(identity: SiteIdentity | null): void {
  override = identity
  for (const listener of listeners) listener()
}

/** Identity streamed by the editor, or null outside preview mode. */
export function usePreviewIdentity(): SiteIdentity | null {
  return useSyncExternalStore(
    (listener) => {
      listeners.add(listener)
      return () => listeners.delete(listener)
    },
    () => override,
  )
}

/** Messages exchanged between the editor and the preview iframe. */
export type PreviewMessage =
  | { type: "preview:ready" }
  | { type: "theme:preview"; css: string; identity: SiteIdentity; scheme: "light" | "dark" | null }
