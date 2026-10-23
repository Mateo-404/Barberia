import { test, expect, type Page } from "@playwright/test"

const json = (body: unknown) => ({ status: 200, contentType: "application/json", body: JSON.stringify(body) })

const scheme = (primary: string) => ({
  bg: "#ffffff", surface: "#ffffff", text: "#111111", muted: "#555555", primary, onPrimary: "#ffffff", accent: "#eeeeee", border: "#dddddd",
  ratios: { text: 18, muted: 7.5, primary: 4.6, onPrimary: 4.6 },
})

const theme = (brand: string) => ({
  v: 2, name: "Barbería Test", tagline: "", logo: "", preset: "classic", mode: "light", brand,
  font: { heading: "Inter", body: "Inter" }, shape: { radius: 8, button: "solid" }, overrides: { light: {}, dark: {} },
})

const css = (primary: string) =>
  `:root{--theme-mode:light;--radius:8px;--button-style:solid;--font-heading:"Inter",sans-serif;--font-body:"Inter",sans-serif;--color-bg:#ffffff;--color-surface:#ffffff;--color-text:#111111;--color-muted:#555555;--color-primary:${primary};--color-on-primary:#ffffff;--color-accent:#eeeeee;--color-border:#dddddd;}`

/** Signs the browser in as an owner and mocks the whole API. */
async function mockOwnerApi(page: Page) {
  const payload = Buffer.from(JSON.stringify({ sub: "1", exp: Math.floor(Date.now() / 1000) + 3600 })).toString("base64url")
  await page.addInitScript((token) => {
    sessionStorage.setItem("auth_token", token)
    sessionStorage.setItem("admin", JSON.stringify({ id: 1, firstName: "Admin", lastName: "Test", email: "a@b.com", role: "owner" }))
  }, `e30.${payload}.sig`)

  const published = { theme: theme("#336699"), business: { whatsapp: "", instagram: "", address: "", currency: "ARS", slot: 30, notice: 1, open: "13:00", close: "20:00", tz: "America/Argentina/Buenos_Aires" }, hash: "h1" }
  let hasUnpublished = false

  await page.route("**/api/public/theme.css", (r) => r.fulfill({ status: 200, contentType: "text/css", body: css("#336699") }))
  await page.route("**/api/public/settings", (r) => r.fulfill(json(published)))
  await page.route("**/api/services", (r) => r.fulfill(json([{ id: 1, name: "Corte", price: 5000 }])))
  await page.route("**/api/appointments/booked-slots", (r) => r.fulfill(json([])))
  await page.route("**/api/admin/settings", (r) =>
    r.fulfill(json({ onboarded: true, draft: theme("#336699"), publishedVersion: 1, hash: "h1", hasUnpublishedChanges: hasUnpublished, preferences: { themeEditorEnabled: true, onboardingEnabled: true } })),
  )
  await page.route("**/api/admin/theme/presets", (r) =>
    r.fulfill(json([
      { id: "classic", theme: { ...theme("#336699"), name: "Barbería" }, light: scheme("#336699"), dark: scheme("#336699") },
      { id: "vintage", theme: { ...theme("#d4145a"), preset: "vintage", name: "Barbería" }, light: scheme("#d4145a"), dark: scheme("#d4145a") },
    ])),
  )
  await page.route("**/api/admin/theme/preview", async (r) => {
    const body = r.request().postDataJSON() as { brand: string }
    await r.fulfill(json({ css: css(body.brand), hash: "x", warnings: [], light: scheme(body.brand), dark: scheme(body.brand) }))
  })
  await page.route("**/api/admin/settings/draft", async (r) => {
    hasUnpublished = true
    await r.fulfill(json({ draft: r.request().postDataJSON(), warnings: [] }))
  })
  await page.route("**/api/admin/settings/publish", async (r) => {
    hasUnpublished = false
    await r.fulfill(json({ version: 2, hash: "h2" }))
  })
  await page.route("**/api/admin/settings/versions", (r) => r.fulfill(json([{ version: 1, hash: "h1", createdAt: "2026-10-01T12:00:00Z" }])))
}

test("picking a preset restyles the live preview, saves the draft and publishes", async ({ page }) => {
  await mockOwnerApi(page)
  await page.goto("/admin/theme")

  await expect(page.getByRole("heading", { name: "Editor de tema" })).toBeVisible()
  const nextButton = page.frameLocator("iframe").getByRole("button", { name: "Siguiente" })
  await expect(nextButton).toHaveCSS("background-color", "rgb(51, 102, 153)")

  const saved = page.waitForRequest((req) => req.url().endsWith("/api/admin/settings/draft") && req.method() === "PUT")
  await page.getByRole("button", { name: /Vintage/ }).click()

  // The preview iframe takes the new color without reloading.
  await expect(nextButton).toHaveCSS("background-color", "rgb(212, 20, 90)")
  expect((await saved).postDataJSON()).toMatchObject({ preset: "vintage", brand: "#d4145a" })
  await expect(page.getByRole("status").filter({ hasText: "Guardado" })).toBeVisible()

  await page.getByRole("button", { name: "Publicar", exact: true }).click()
  await page.getByRole("dialog").getByRole("button", { name: "Publicar" }).click()
  await expect(page.getByRole("status").filter({ hasText: "Publicado" })).toBeVisible()
})

test("the preview page cannot submit a booking", async ({ page }) => {
  await mockOwnerApi(page)
  await page.goto("/?preview=1")
  await page.getByText("Corte").click()
  await page.getByRole("button", { name: /siguiente/i }).click()
  await page.getByLabel(/fecha/i).fill(new Date(Date.now() + 2 * 86_400_000).toISOString().slice(0, 10))
  await page.getByRole("button", { name: "14:00" }).click()
  await page.getByRole("button", { name: /siguiente/i }).click()
  await expect(page.getByRole("button", { name: /reservar turno/i })).toBeDisabled()
})

test("turning the theme editor off in settings hides it and redirects the route", async ({ page }) => {
  await mockOwnerApi(page)
  await page.route("**/api/admin/settings", (r) =>
    r.fulfill(json({ onboarded: true, draft: theme("#336699"), publishedVersion: 1, hash: "h1", hasUnpublishedChanges: false, preferences: { themeEditorEnabled: false, onboardingEnabled: true } })),
  )
  await page.goto("/admin/theme")
  await expect(page).toHaveURL(/\/admin\/settings$/)
  await expect(page.getByRole("link", { name: "Tema" })).toHaveCount(0)
})

test("the welcome wizard names the business, picks a style and publishes", async ({ page }) => {
  await mockOwnerApi(page)
  await page.route("**/api/admin/settings", (r) =>
    r.fulfill(json({ onboarded: false, draft: theme("#336699"), publishedVersion: 0, hash: "h0", hasUnpublishedChanges: true, preferences: { themeEditorEnabled: true, onboardingEnabled: true } })),
  )
  let published = false
  await page.route("**/api/admin/settings/publish", (r) => {
    published = true
    return r.fulfill(json({ version: 1, hash: "h1" }))
  })
  await page.goto("/welcome")

  await expect(page.getByText("Paso 1 de 3")).toBeVisible()
  await page.getByLabel("Nombre del negocio").fill("Don Mateo")
  await page.getByRole("button", { name: "Siguiente" }).click()
  await expect(page.getByText("Paso 2 de 3")).toBeVisible()
  await page.getByRole("button", { name: /Vintage/ }).click()
  await page.getByRole("button", { name: "Siguiente" }).click()
  await page.getByRole("button", { name: /Publicar y entrar/ }).click()
  await expect(page).toHaveURL(/\/admin$/)
  expect(published).toBe(true)
})

test("skipping the wizard turns it off for good", async ({ page }) => {
  await mockOwnerApi(page)
  await page.route("**/api/admin/settings", (r) =>
    r.fulfill(json({ onboarded: false, draft: theme("#336699"), publishedVersion: 0, hash: "h0", hasUnpublishedChanges: true, preferences: { themeEditorEnabled: true, onboardingEnabled: true } })),
  )
  const patch = page.waitForRequest((req) => req.url().endsWith("/preferences") && req.method() === "PATCH")
  await page.route("**/api/admin/settings/preferences", (r) => r.fulfill(json({ themeEditorEnabled: true, onboardingEnabled: false })))
  await page.goto("/welcome")
  await page.getByRole("button", { name: /Saltar/ }).click()
  expect((await patch).postDataJSON()).toEqual({ onboardingEnabled: false })
  await expect(page).toHaveURL(/\/admin$/)
})
