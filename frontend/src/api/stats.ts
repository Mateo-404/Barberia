import { useQuery } from "@tanstack/react-query"
import { api } from "@/lib/api-client"
import { queryKeys } from "@/lib/query-keys"
import { useTenant } from "@/context/TenantContext"
import type { components } from "@/types/api"

export type Dashboard = components["schemas"]["Dashboard"]

export function useDashboard() {
  const tenant = useTenant()
  return useQuery({
    queryKey: queryKeys.stats.dashboard(tenant),
    queryFn: ({ signal }) => api<Dashboard>("/api/stats/dashboard", { signal }),
  })
}
