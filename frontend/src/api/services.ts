import { useQuery } from "@tanstack/react-query"
import { api } from "@/lib/api-client"
import { queryKeys } from "@/lib/query-keys"
import { useTenant } from "@/context/TenantContext"
import type { components } from "@/types/api"

export type Service = components["schemas"]["Service"]

export function useServices() {
  const tenant = useTenant()
  return useQuery({
    queryKey: queryKeys.services(tenant),
    queryFn: ({ signal }) => api<Service[]>("/api/services", { signal }),
    retry: 1,
    staleTime: 5 * 60_000,
  })
}
