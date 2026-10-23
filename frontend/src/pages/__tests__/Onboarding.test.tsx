import { describe, expect, it } from "vitest"
import { render, screen } from "@testing-library/react"
import { MemoryRouter } from "react-router-dom"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import Onboarding from "@/pages/Onboarding"

describe("Onboarding (F-01/F-05)", () => {
  it("renderiza el paso 1 en español", () => {
    const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } })
    render(
      <QueryClientProvider client={qc}>
        <MemoryRouter>
          <Onboarding />
        </MemoryRouter>
      </QueryClientProvider>,
    )
    expect(screen.getByText(/Configurá tu barbería/i)).toBeInTheDocument()
    expect(screen.getByText(/Paso 1 de 6/i)).toBeInTheDocument()
  })
})
