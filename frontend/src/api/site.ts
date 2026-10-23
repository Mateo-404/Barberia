import { useQuery } from "@tanstack/react-query"
import { api } from "@/lib/api-client"
import { useTenant } from "@/context/TenantContext"
import { usePreviewIdentity, type SiteIdentity } from "@/lib/site-preview"
import type { components } from "@/types/api"

export type PublicSettings = components["schemas"]["PublicSettings"]

/** Published theme and business data of the business. */
export function usePublicSettings() {
  const tenant = useTenant()
  return useQuery({
    queryKey: ["public-settings", tenant] as const,
    queryFn: ({ signal }) => api<PublicSettings>("/api/public/settings", { signal }),
    staleTime: 60_000,
  })
}

/** Name, tagline and logo to show: the editor's unsaved ones while previewing. */
export function useSiteIdentity(): SiteIdentity {
  const preview = usePreviewIdentity()
  const { data } = usePublicSettings()
  return preview ?? { name: data?.theme?.name ?? "", tagline: data?.theme?.tagline ?? "", logo: data?.theme?.logo ?? "" }
}
