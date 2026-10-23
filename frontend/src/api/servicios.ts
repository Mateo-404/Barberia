import { useQuery } from "@tanstack/react-query"
import { api } from "@/lib/api-client"
import { queryKeys } from "@/lib/query-keys"
import { useTenant } from "@/context/TenantContext"
import type { components } from "@/types/api"

type Servicio = components["schemas"]["ServicioResponseDTO"]

export function useServicios() {
  const tenant = useTenant()
  return useQuery({
    queryKey: queryKeys.servicios(tenant),
    queryFn: ({ signal }) => api<Servicio[]>("/servicios", { signal }),
    // 4xx no se reintenta (el retry por defecto de Query sólo hace falta
    // acotarlo para errores no transitorios).
    retry: 1,
    staleTime: 5 * 60_000,
  })
}
