import { useEffect } from "react"
import { applyThemeCss } from "@/lib/tenant-theme"
import { isPreview, setPreviewIdentity, type PreviewMessage } from "@/lib/site-preview"

/**
 * Receives unsaved theme changes from the editor when this page runs inside
 * its preview iframe. Only same-origin messages from the parent window are
 * accepted; outside preview mode it does nothing.
 */
export function PreviewBridge() {
  useEffect(() => {
    if (!isPreview()) return
    function onMessage(event: MessageEvent<PreviewMessage>) {
      if (event.origin !== window.location.origin || event.source !== window.parent) return
      if (event.data?.type !== "theme:preview") return
      applyThemeCss(event.data.css)
      setPreviewIdentity(event.data.identity)
      const root = document.documentElement
      if (event.data.scheme) root.dataset.scheme = event.data.scheme
      else delete root.dataset.scheme
    }
    window.addEventListener("message", onMessage)
    const ready: PreviewMessage = { type: "preview:ready" }
    window.parent.postMessage(ready, window.location.origin)
    return () => window.removeEventListener("message", onMessage)
  }, [])
  return null
}
