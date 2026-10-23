/**
 * Hierarchical, typed query keys. Every key starts with the tenant, so
 * - `invalidateQueries({ queryKey: queryKeys.appointments.all(tenant) })`
 *   refreshes only that business's appointments;
 * - `queryClient.clear()` (logout / account switch) drops everything.
 */
export type TenantSlug = string

export interface AppointmentFilters {
  search?: string
  status?: string
  service?: number | "all"
  from?: string
  to?: string
}

export const queryKeys = {
  services: (tenant: TenantSlug) => ["services", tenant] as const,
  /** Booked slots: invalidate after every successful booking. */
  availability: (tenant: TenantSlug) => ["availability", tenant] as const,
  appointments: {
    all: (tenant: TenantSlug) => ["appointments", tenant] as const,
    list: (tenant: TenantSlug, filters: AppointmentFilters) =>
      ["appointments", tenant, "list", filters] as const,
    page: (tenant: TenantSlug, page: number, size: number) =>
      ["appointments", tenant, "page", { page, size }] as const,
    latest: (tenant: TenantSlug, size: number) =>
      ["appointments", tenant, "latest", size] as const,
  },
  stats: {
    dashboard: (tenant: TenantSlug) => ["stats", tenant, "dashboard"] as const,
  },
  /** Brand settings: admin state, draft and published versions. */
  settings: {
    admin: (tenant: TenantSlug) => ["settings", tenant, "admin"] as const,
    versions: (tenant: TenantSlug) => ["settings", tenant, "versions"] as const,
  },
} as const
