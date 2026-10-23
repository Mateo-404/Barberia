import { test, expect } from "@playwright/test"

test("wizard de reserva carga servicios y muestra error si falla", async ({ page }) => {
  await page.route("**/servicios", async (route) => {
    await route.fulfill({ status: 500, contentType: "application/json", body: JSON.stringify({ title: "Error", detail: "boom", status: 500 }) })
  })
  await page.goto("/")
  await expect(page.getByText(/no se pudieron cargar los servicios/i)).toBeVisible()
})

test("wizard de reserva muestra servicios y permite seleccionar", async ({ page }) => {
  await page.route("**/servicios", async (route) => {
    await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify([{ id: 1, tipo: "Corte", precio: 5000 }]) })
  })
  await page.route("**/turnos/findDateTimes", async (route) => {
    await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify([]) })
  })
  await page.goto("/")
  await expect(page.getByText("Corte")).toBeVisible()
  await page.getByText("Corte").click()
  await page.getByRole("button", { name: /siguiente/i }).click()
  await expect(page.getByText(/elegí fecha y horario/i)).toBeVisible()
})
