import { useEffect, useRef, useState } from "react"
import { useTranslation } from "react-i18next"
import type { SiteIdentity, PreviewMessage } from "@/lib/site-preview"

export type Device = "mobile" | "desktop"

/** Logical size of each device; the frame is scaled down to fit the canvas. */
const SIZES: Record<Device, { width: number; height: number }> = {
  mobile: { width: 390, height: 800 },
  desktop: { width: 1280, height: 800 },
}

/**
 * The real booking page in an iframe, restyled live: each new stylesheet from
 * the server is posted to the page, which applies it without reloading.
 */
export function PreviewCanvas({
  css,
  identity,
  device,
  scheme,
}: {
  css: string | null
  identity: SiteIdentity
  device: Device
  scheme: "light" | "dark" | null
}) {
  const { t } = useTranslation("theme")
  const frame = useRef<HTMLIFrameElement>(null)
  const canvas = useRef<HTMLDivElement>(null)
  const [ready, setReady] = useState(false)
  const [available, setAvailable] = useState({ width: 0, height: 0 })

  useEffect(() => {
    const element = canvas.current
    if (!element) return
    const observer = new ResizeObserver(([entry]) =>
      setAvailable({ width: entry.contentRect.width, height: entry.contentRect.height }),
    )
    observer.observe(element)
    return () => observer.disconnect()
  }, [])

  useEffect(() => {
    function onMessage(event: MessageEvent<PreviewMessage>) {
      if (event.origin !== window.location.origin || event.source !== frame.current?.contentWindow) return
      if (event.data?.type === "preview:ready") setReady(true)
    }
    window.addEventListener("message", onMessage)
    return () => window.removeEventListener("message", onMessage)
  }, [])

  useEffect(() => {
    if (!ready || css === null) return
    const message: PreviewMessage = { type: "theme:preview", css, identity, scheme }
    frame.current?.contentWindow?.postMessage(message, window.location.origin)
  }, [ready, css, identity, scheme])

  const size = SIZES[device]
  const scale = Math.min(1, (available.width - 32) / size.width, (available.height - 32) / size.height) || 1

  return (
    <div ref={canvas} className="relative flex h-full w-full items-center justify-center overflow-hidden bg-secondary/60">
      <div
        className={`overflow-hidden bg-background shadow-[0_18px_50px_-18px_rgba(0,0,0,0.45)] ring-1 ring-border ${
          device === "mobile" ? "rounded-[2rem]" : "rounded-lg"
        }`}
        style={{ width: size.width * scale, height: size.height * scale }}
      >
        <iframe
          ref={frame}
          src="/?preview=1"
          title={t("preview.title")}
          className="block border-0 bg-background"
          style={{
            width: size.width,
            height: size.height,
            transform: `scale(${scale})`,
            transformOrigin: "top left",
          }}
        />
      </div>
      {!ready && (
        <p className="absolute text-sm text-muted-foreground" role="status">
          {t("preview.loading")}
        </p>
      )}
    </div>
  )
}
