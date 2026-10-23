/**
 * Genera el HTML de la vista previa de la marca.
 *
 * F-04: este HTML se inyecta en un `<iframe srcDoc>`. TODO texto que venga
 * del usuario se escapa con `escapeHtml` y el logo sólo se usa si es una URL
 * o un data URI de imagen admitidos por el backend. Nunca se usa
 * `dangerouslySetInnerHTML` con estos valores.
 */

import { escapeHtml, hexToRgba, onPrimaryColor, previewSlots, safeLogoSrc, formatPrice } from "@/lib/settings"
import type { Settings } from "@/types/settings"

const PREVIEW_CSS = `
*{box-sizing:border-box}
body{margin:0;background:var(--bg);color:var(--t);font:15px/1.5 var(--bf),system-ui,sans-serif}
#pv{font-family:var(--bf),system-ui,sans-serif}
#pv nav{display:flex;align-items:center;gap:10px;padding:12px 16px;background:var(--s)}
#pv nav img{height:30px;width:auto;border-radius:4px}
#pv .nm{font-family:var(--hf),serif;font-weight:700;font-size:18px;color:var(--t)}
#pv .hero{padding:28px 16px}
#pv h3{font-family:var(--hf),serif;font-size:30px;line-height:1.1;margin:0 0 6px}
#pv .tag{opacity:.8;margin:0 0 16px}
#pv .cta{display:inline-block;padding:10px 18px;font-weight:700;border-radius:var(--r);background:var(--p);color:var(--on);border:2px solid var(--p)}
#pv .cta.outline{background:transparent;color:var(--p)}
#pv .sv{display:grid;grid-template-columns:repeat(3,1fr);gap:8px;padding:0 16px 16px}
#pv .sv div{background:var(--s);border-radius:var(--r);padding:10px;font-size:13px}
#pv .sv b{display:block;font-family:var(--hf),serif;font-size:15px}
#pv .sv span{color:var(--a);font-weight:700}
#pv .sl{display:flex;flex-wrap:wrap;gap:6px;padding:0 16px 16px}
#pv .sl em{font-style:normal;font-size:12px;padding:4px 10px;border:1px solid var(--a);color:var(--a);border-radius:var(--r)}
#pv footer{padding:12px 16px;background:var(--s);font-size:12px;opacity:.9}
`

/** Variables CSS vivas (las mismas que emite `RenderCSS` del backend). */
export function previewCssVars(settings: Settings): string {
  const { c, r, hf, bf } = settings
  return [
    `--p:${c.primary}`,
    `--a:${c.accent}`,
    `--bg:${c.bg}`,
    `--s:${c.surface}`,
    `--t:${c.text}`,
    `--on:${onPrimaryColor(c.primary)}`,
    `--r:${r}px`,
    `--hf:'${hf}'`,
    `--bf:'${bf}'`,
  ].join(";")
}

/** Construye el documento completo del preview. */
export function buildPreviewHtml(settings: Settings): string {
  const { c, btn, biz } = settings
  const logo = safeLogoSrc(settings.logo)
  const slots = previewSlots(settings)
  const services = [
    { name: "Corte", minutes: biz.slot, price: 8000 },
    { name: "Barba", minutes: 20, price: 5000 },
    { name: "Combo", minutes: biz.slot + 20, price: 11000 },
  ]

  const nav = logo
    ? `<img src="${escapeHtml(logo)}" alt="">`
    : ""
  const servicesHtml = services
    .map(
      (s) =>
        `<div><b>${escapeHtml(s.name)}</b>${s.minutes} min<br><span>${escapeHtml(
          formatPrice(s.price, biz.currency),
        )}</span></div>`,
    )
    .join("")
  const slotsHtml = slots.map((slot) => `<em>${escapeHtml(slot)}</em>`).join("")
  const footerParts = [
    biz.address,
    biz.whatsapp ? `WhatsApp ${biz.whatsapp}` : "",
    biz.instagram,
  ].filter(Boolean)

  return `<!doctype html>
<html lang="es">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<style>:root{${previewCssVars(settings)};--tint:${hexToRgba(c.accent, 0.2)}}${PREVIEW_CSS}</style>
</head>
<body>
<div id="pv">
  <nav>${nav}<span class="nm">${escapeHtml(settings.name)}</span></nav>
  <div class="hero">
    <h3>${escapeHtml(settings.name)}</h3>
    <p class="tag">${escapeHtml(settings.tagline)}</p>
    <span class="cta${btn === "outline" ? " outline" : ""}">Reservar turno</span>
  </div>
  <div class="sv">${servicesHtml}</div>
  <div class="sl">${slotsHtml}</div>
  <footer>${escapeHtml(footerParts.join(" · "))}</footer>
</div>
</body>
</html>`
}
