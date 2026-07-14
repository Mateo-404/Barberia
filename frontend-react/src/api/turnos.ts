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
