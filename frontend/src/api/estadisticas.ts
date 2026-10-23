import { useQuery } from "@tanstack/react-query"
import { api } from "@/lib/api-client"
import { queryKeys } from "@/lib/query-keys"
import { useTenant } from "@/context/TenantContext"
import type { components } from "@/types/api"

type PanelEstadistica = components["schemas"]["PanelEstadisticaDTO"]
type TurnoResponse = components["schemas"]["TurnoResponseDTO"]

export function usePanelEstadisticas() {
  const tenant = useTenant()
  return useQuery({
    queryKey: queryKeys.estadisticas.panel(tenant),
    queryFn: ({ signal }) => api<PanelEstadistica>("/estadisticas/panel", { signal }),
  })
}

export function useUltimosTurnos(size = 10) {
  const tenant = useTenant()
  return useQuery({
    queryKey: queryKeys.turnos.ultimos(tenant, size),
    queryFn: async ({ signal }) => {
      const page = await api<components["schemas"]["PageTurno"]>(
        `/turnos/0/${size}`,
        { signal },
      )
      const content = page.content ?? []
      return content.map<TurnoResponse>((t) => ({
        id: t.id ?? 0,
        fechaHora: t.fechaHora ?? "",
        telefonoCliente: t.cliente?.telefono ?? "",
        nombreCliente: t.cliente?.nombre ?? "",
        apellidoCliente: t.cliente?.apellido ?? "",
        emailCliente: t.cliente?.email,
        idServicio: t.servicio?.id ?? 0,
        tipoServicio: t.servicio?.tipo ?? "",
        precioServicio: t.servicio?.precio ?? 0,
      }))
    },
  })
}
