import { createContext, useContext, type ReactNode } from "react"

/**
 * Identifier of the business (tenant) whose data the app requests.
 *
 * IMPORTANT: the slug is a cache *data scope*, NEVER an authorization. Tenant
 * isolation is enforced by the API (Go backend).
 *
 * Today it comes from `VITE_TENANT_SLUG` (single-business deployment). With
 * per-business public routes it would be resolved from the URL in this
 * provider; the query keys would not change.
 */
const TenantContext = createContext<string>(import.meta.env.VITE_TENANT_SLUG ?? "default")

export function TenantProvider({ slug, children }: { slug: string; children: ReactNode }) {
  return <TenantContext.Provider value={slug}>{children}</TenantContext.Provider>
}

export function useTenant(): string {
  return useContext(TenantContext)
}
