import { useQuery } from "@tanstack/react-query"
import { api } from "@/lib/api-client"
import type { components } from "@/types/api"

type PanelEstadistica = components["schemas"]["PanelEstadisticaDTO"]
type TurnoResponse = components["schemas"]["TurnoResponseDTO"]

export function usePanelEstadisticas() {
  return useQuery({
    queryKey: ["estadisticas", "panel"],
    queryFn: () => api<PanelEstadistica>("/estadisticas/panel"),
  })
}

export function useUltimosTurnos(size = 10) {
  return useQuery({
    queryKey: ["turnos", "ultimos", size],
    queryFn: async () => {
      const page = await api<components["schemas"]["PageTurno"]>(
        `/turnos/0/${size}`,
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
