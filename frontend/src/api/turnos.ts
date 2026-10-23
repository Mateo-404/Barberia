import { useMutation, useQuery } from "@tanstack/react-query"
import { api } from "@/lib/api-client"
import type { components } from "@/types/api"

type TurnoRequest = components["schemas"]["TurnoRequestDTO"]
type TurnoResponse = components["schemas"]["TurnoResponseDTO"]

export function useFechasOcupadas() {
  return useQuery({
    queryKey: ["fechasOcupadas"],
    queryFn: () => api<string[]>("/turnos/findDateTimes"),
  })
}

export function useCrearTurno() {
  return useMutation({
    mutationFn: (data: TurnoRequest) =>
      api<TurnoResponse>("/turnos", {
        method: "POST",
        body: JSON.stringify(data),
      }),
  })
}

export function useTodosTurnos() {
  return useQuery({
    queryKey: ["turnos", "todos"],
    queryFn: () => api<TurnoResponse[]>("/turnos"),
  })
}

type TurnoPatch = {
  fechaHora?: string
  servicio?: { id: number }
}

export function useActualizarTurno() {
  return useMutation({
    mutationFn: ({ id, patch }: { id: number; patch: TurnoPatch }) =>
      api<TurnoResponse>(`/turnos/${id}`, {
        method: "PATCH",
        body: JSON.stringify(patch),
      }),
  })
}

export function useEliminarTurno() {
  return useMutation({
    mutationFn: (id: number) =>
      api<void>(`/turnos/${id}`, { method: "DELETE" }),
  })
}
