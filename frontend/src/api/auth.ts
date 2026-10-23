import { useMutation, useQuery } from "@tanstack/react-query"
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

export function useAuthProviders() {
  return useQuery({
    queryKey: ["auth-providers"],
    queryFn: () => api<{ google: boolean }>("/api/auth/providers"),
    staleTime: 5 * 60_000,
  })
}

export function useGoogleStart() {
  return useMutation({
    mutationFn: (redirect?: string) =>
      api<{ url: string; state: string }>("/api/auth/google/start", {
        method: "POST",
        body: JSON.stringify({ redirect: redirect ?? "/admin" }),
      }),
  })
}

export function useGoogleCallback() {
  return useMutation({
    mutationFn: (data: { code: string; state: string }) =>
      api<{ token?: string; requires2fa?: boolean; tempToken?: string; admin?: any }>(
        "/api/auth/google/callback",
        { method: "POST", body: JSON.stringify(data) },
      ),
  })
}

export function useTOTPSetup() {
  return useMutation({
    mutationFn: () => api<{ secret: string; otpauthUrl: string }>("/api/auth/totp/setup", { method: "POST" }),
  })
}

export function useTOTPVerify() {
  return useMutation({
    mutationFn: (code: string) => api<{ enabled: boolean }>("/api/auth/totp/verify", { method: "POST", body: JSON.stringify({ code }) }),
  })
}

export function useTOTPDisable() {
  return useMutation({
    mutationFn: (code: string) => api<{ enabled: boolean }>("/api/auth/totp/disable", { method: "POST", body: JSON.stringify({ code }) }),
  })
}

export function useTOTPChallenge() {
  return useMutation({
    mutationFn: (data: { tempToken: string; code: string }) =>
      api<{ token: string; admin: any }>("/api/auth/totp/challenge", { method: "POST", body: JSON.stringify(data) }),
  })
}
