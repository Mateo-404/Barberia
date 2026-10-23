/**
 * Query keys jerárquicas y tipadas. Toda key lleva el tenant primero, así
 * - `invalidateQueries({ queryKey: queryKeys.turnos.all(tenant) })` refrésca
 *   sólo turnos de ese negocio;
 * - `queryClient.clear()` (logout / cambio de cuenta) descarta todo.
 */
export type TenantSlug = string

export interface FiltrosTurno {
  busqueda?: string
  estado?: string
  servicio?: number | "todos"
  desde?: string
  hasta?: string
}

export const queryKeys = {
  servicios: (tenant: TenantSlug) => ["servicios", tenant] as const,
  /** Disponibilidad/fechas ocupadas: invalidar tras cada reserva exitosa. */
  disponibilidad: (tenant: TenantSlug) => ["disponibilidad", tenant] as const,
  turnos: {
    all: (tenant: TenantSlug) => ["turnos", tenant] as const,
    lista: (tenant: TenantSlug, filtros: FiltrosTurno) =>
      ["turnos", tenant, "lista", filtros] as const,
    pagina: (tenant: TenantSlug, page: number, size: number) =>
      ["turnos", tenant, "pagina", { page, size }] as const,
    ultimos: (tenant: TenantSlug, size: number) =>
      ["turnos", tenant, "ultimos", size] as const,
  },
  estadisticas: {
    panel: (tenant: TenantSlug) => ["estadisticas", tenant, "panel"] as const,
  },
  /** Marca del negocio: estado del admin, borrador y versiones publicadas. */
  settings: {
    admin: (tenant: TenantSlug) => ["settings", tenant, "admin"] as const,
    versiones: (tenant: TenantSlug) => ["settings", tenant, "versiones"] as const,
  },
} as const
