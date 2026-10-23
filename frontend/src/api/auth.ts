import { useMutation } from "@tanstack/react-query"
import { api } from "@/lib/api-client"
import type { components } from "@/types/api"

type LoginRequest = components["schemas"]["LoginRequest"]
export type LoginResponse = components["schemas"]["LoginResponse"]

export function useLogin() {
  return useMutation({
    mutationFn: (data: LoginRequest) =>
      api<LoginResponse>("/api/auth/login", { method: "POST", body: JSON.stringify(data) }),
  })
}
