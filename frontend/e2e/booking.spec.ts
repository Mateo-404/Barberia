import { test, expect } from "@playwright/test"

test("booking wizard shows a translated error when services fail to load", async ({ page }) => {
  await page.route("**/api/services", (route) =>
    route.fulfill({ status: 500, contentType: "application/problem+json", body: JSON.stringify({ code: "server.internal", status: 500 }) }),
  )
  await page.goto("/")
  await expect(page.getByText(/no pudimos cargar los servicios/i)).toBeVisible()
})

test("booking wizard lists services and moves to date and time", async ({ page }) => {
  await page.route("**/api/services", (route) =>
    route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify([{ id: 1, name: "Corte", price: 5000 }]) }),
  )
  await page.route("**/api/appointments/booked-slots", (route) =>
    route.fulfill({ status: 200, contentType: "application/json", body: "[]" }),
  )
  await page.goto("/")
  await page.getByText("Corte").click()
  await page.getByRole("button", { name: /siguiente/i }).click()
  await expect(page.getByText(/elegí fecha y horario/i)).toBeVisible()
})

test("a slot taken by someone else shows the conflict message", async ({ page }) => {
  await page.route("**/api/services", (route) =>
    route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify([{ id: 1, name: "Corte", price: 5000 }]) }),
  )
  await page.route("**/api/appointments/booked-slots", (route) =>
    route.fulfill({ status: 200, contentType: "application/json", body: "[]" }),
  )
  await page.route("**/api/appointments", (route) =>
    route.request().method() === "POST"
      ? route.fulfill({ status: 409, contentType: "application/problem+json", body: JSON.stringify({ code: "appointment.slot_taken", status: 409 }) })
      : route.continue(),
  )
  await page.goto("/")
  await page.getByText("Corte").click()
  await page.getByRole("button", { name: /siguiente/i }).click()
  const tomorrow = new Date(Date.now() + 2 * 86_400_000).toISOString().slice(0, 10)
  await page.getByLabel(/fecha/i).fill(tomorrow)
  await page.getByRole("button", { name: "14:00" }).click()
  await page.getByRole("button", { name: /siguiente/i }).click()
  await page.getByLabel("Nombre").fill("Ana")
  await page.getByLabel("Apellido").fill("Paz")
  await page.getByLabel("Teléfono").fill("1155551234")
  await page.getByRole("button", { name: /reservar turno/i }).click()
  await expect(page.getByRole("alert")).toContainText(/ya fue reservado/i)
})
