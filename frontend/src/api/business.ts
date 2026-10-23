import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { api } from "@/lib/api-client"
import { applyBusiness } from "@/lib/business-config"
import { useTenant } from "@/context/TenantContext"
import type { components } from "@/types/api"

export type Business = components["schemas"]["Business"]

/** Operating data of the business (hours, currency, contact). */
export function useBusiness() {
  const tenant = useTenant()
  return useQuery({
    queryKey: ["business", tenant] as const,
    queryFn: ({ signal }) => api<Business>("/api/admin/business", { signal }),
  })
}

/** Business data is saved directly: it is not versioned or published. */
export function useSaveBusiness() {
  const tenant = useTenant()
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (business: Business) =>
      api<Business>("/api/admin/business", { method: "PUT", body: JSON.stringify(business) }),
    onSuccess: (saved) => {
      applyBusiness(saved)
      void queryClient.invalidateQueries({ queryKey: ["business", tenant] })
      void queryClient.invalidateQueries({ queryKey: ["public-settings", tenant] })
    },
  })
}
