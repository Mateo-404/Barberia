import { useQuery } from "@tanstack/react-query"
import { api } from "@/lib/api-client"
import type { components } from "@/types/api"

type Servicio = components["schemas"]["ServicioResponseDTO"]

export function useServicios() {
  return useQuery({
    queryKey: ["servicios"],
    queryFn: () => api<Servicio[]>("/servicios"),
  })
}
