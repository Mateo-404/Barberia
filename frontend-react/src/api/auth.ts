import { useMutation } from "@tanstack/react-query"
import { api } from "@/lib/api-client"
import type { components } from "@/types/api"

type LoginRequest = components["schemas"]["LoginRequestDTO"]
export type LoginResponse = Required<components["schemas"]["LoginResponseDTO"]>

export function useLogin() {
  return useMutation({
    mutationFn: (data: LoginRequest) =>
      api<LoginResponse>("/administradores/login", {
        method: "POST",
        body: JSON.stringify(data),
      }),
  })
}
