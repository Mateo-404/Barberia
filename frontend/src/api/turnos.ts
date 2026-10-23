import { useMutation, useQuery, useQueryClient, type QueryClient } from "@tanstack/react-query"
import { api } from "@/lib/api-client"
import { queryKeys, type FiltrosTurno } from "@/lib/query-keys"
import { useTenant } from "@/context/TenantContext"
import type { components } from "@/types/api"

type TurnoRequest = components["schemas"]["TurnoRequestDTO"]
type TurnoResponse = components["schemas"]["TurnoResponseDTO"]

/** Fechas ocupadas: la grilla de disponibilidad se reconstruye con esto. */
export function useFechasOcupadas() {
  const tenant = useTenant()
  return useQuery({
    queryKey: queryKeys.disponibilidad(tenant),
    queryFn: ({ signal }) => api<string[]>("/turnos/findDateTimes", { signal }),
  })
}

export function useCrearTurno() {
  const tenant = useTenant()
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (data: TurnoRequest) =>
      api<TurnoResponse>("/turnos", {
        method: "POST",
        body: JSON.stringify(data),
      }),
    onSuccess: () => {
      // El turno ya no está libre: refrescar la disponibilidad del negocio.
      void queryClient.invalidateQueries({ queryKey: queryKeys.disponibilidad(tenant) })
      void queryClient.invalidateQueries({ queryKey: queryKeys.turnos.all(tenant) })
    },
  })
}

export function useTodosTurnos() {
  const tenant = useTenant()
  return useQuery({
    queryKey: queryKeys.turnos.all(tenant),
    queryFn: ({ signal }) => api<TurnoResponse[]>("/turnos", { signal }),
  })
}

export function useTurnosPaginados(page: number, size: number) {
  const tenant = useTenant()
  return useQuery({
    queryKey: queryKeys.turnos.pagina(tenant, page, size),
    queryFn: ({ signal }) =>
      api<components["schemas"]["PageTurno"]>(`/turnos/${page}/${size}`, { signal }),
  })
}

type TurnoPatch = {
  fechaHora?: string
  servicio?: { id: number }
}

export function useActualizarTurno() {
  const tenant = useTenant()
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, patch }: { id: number; patch: TurnoPatch }) =>
      api<TurnoResponse>(`/turnos/${id}`, {
        method: "PATCH",
        body: JSON.stringify(patch),
      }),
    onSuccess: () => invalidateTurnos(queryClient, tenant),
  })
}

export function useEliminarTurno() {
  const tenant = useTenant()
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: number) =>
      api<void>(`/turnos/${id}`, { method: "DELETE" }),
    onSuccess: () => invalidateTurnos(queryClient, tenant),
  })
}

/** Refresca todo lo que depende de turnos de ESTE negocio. */
function invalidateTurnos(queryClient: QueryClient, tenant: string) {
  void queryClient.invalidateQueries({ queryKey: queryKeys.turnos.all(tenant) })
  void queryClient.invalidateQueries({ queryKey: queryKeys.disponibilidad(tenant) })
  void queryClient.invalidateQueries({ queryKey: queryKeys.estadisticas.panel(tenant) })
}

export type { TurnoRequest, TurnoResponse, FiltrosTurno }
