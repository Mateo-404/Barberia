import { createContext, useContext, type ReactNode } from "react"

/**
 * Identificador del negocio (tenant) cuyos datos se piden a la API.
 *
 * IMPORTANTE: el slug es un *scope de datos* para la caché, NUNCA una
 * autorización. El aislamiento entre negocios lo impone la API (backend Go).
 *
 * Hoy el valor sale de `VITE_TENANT_SLUG` (despliegue de un solo negocio).
 * Cuando existan rutas públicas por negocio (`/:slug/reservar`), se resuelve
 * desde la URL dentro de este provider — las query keys no cambian.
 */
const TenantContext = createContext<string>(import.meta.env.VITE_TENANT_SLUG ?? "default")

export function TenantProvider({ slug, children }: { slug: string; children: ReactNode }) {
  return <TenantContext.Provider value={slug}>{children}</TenantContext.Provider>
}

export function useTenant(): string {
  return useContext(TenantContext)
}
