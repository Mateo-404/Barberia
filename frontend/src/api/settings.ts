/**
 * API de la marca del negocio (SPEC §8).
 *
 * Todas las rutas cuelgan de `/api/admin/*` y requieren el Bearer del admin:
 * el `api-client` lo agrega automáticamente.
 */

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { api, ApiError } from "@/lib/api-client"
import { queryKeys } from "@/lib/query-keys"
import { useTenant } from "@/context/TenantContext"
import type {
  Settings,
  SettingsAdminState,
  SettingsDraftResponse,
  SettingsPublishResponse,
  SettingsValidationDetail,
  SettingsVersion,
} from "@/types/settings"

/** Opciones de la query del estado de marca (compartidas con `fetchQuery`). */
export function adminSettingsQueryOptions(tenant: string) {
  return {
    queryKey: queryKeys.settings.admin(tenant),
    queryFn: ({ signal }: { signal: AbortSignal }) =>
      api<SettingsAdminState>("/api/admin/settings", { signal }),
  }
}

/**
 * Estado de la marca del negocio.
 *
 * `enabled: false` evita pedirla sin sesión (p. ej. en la pantalla de login,
 * donde sólo se consulta una vez autenticado).
 */
export function useAdminSettings(options: { enabled?: boolean } = {}) {
  const tenant = useTenant()
  return useQuery({
    ...adminSettingsQueryOptions(tenant),
    enabled: options.enabled ?? true,
  })
}

/** Guarda el borrador. No cambia el sitio público. */
export function useSaveDraft() {
  const tenant = useTenant()
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (settings: Settings) =>
      api<SettingsDraftResponse>("/api/admin/settings/draft", {
        method: "PUT",
        body: JSON.stringify(settings),
      }),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: queryKeys.settings.admin(tenant) })
    },
  })
}

/** Publica el borrador: crea la versión N+1 y marca `onboarded = true`. */
export function usePublishSettings() {
  const tenant = useTenant()
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: () =>
      api<SettingsPublishResponse>("/api/admin/settings/publish", { method: "POST" }),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: queryKeys.settings.admin(tenant) })
      void queryClient.invalidateQueries({ queryKey: queryKeys.settings.versiones(tenant) })
    },
  })
}

/** Últimas versiones publicadas. */
export function useSettingsVersions() {
  const tenant = useTenant()
  return useQuery({
    queryKey: queryKeys.settings.versiones(tenant),
    queryFn: ({ signal }: { signal: AbortSignal }) =>
      api<SettingsVersion[]>("/api/admin/settings/versions", { signal }),
  })
}

/** Restaura una versión: el backend crea una versión nueva con ese contenido. */
export function useRollbackSettings() {
  const tenant = useTenant()
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (version: number) =>
      api<SettingsPublishResponse>("/api/admin/settings/rollback", {
        method: "POST",
        body: JSON.stringify({ version }),
      }),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: queryKeys.settings.admin(tenant) })
      void queryClient.invalidateQueries({ queryKey: queryKeys.settings.versiones(tenant) })
    },
  })
}

/**
 * Extrae los errores por campo de un 422 del backend
 * (`{error, code:"validation_failed", details:[{field,message}]}`).
 */
export function validationDetails(error: unknown): SettingsValidationDetail[] {
  if (!(error instanceof ApiError) || error.status !== 422) return []
  const body = error.body
  if (!body || typeof body !== "object") return []
  const details = (body as { details?: unknown }).details
  if (!Array.isArray(details)) return []
  return details.flatMap((item) => {
    if (!item || typeof item !== "object") return []
    const { field, message } = item as { field?: unknown; message?: unknown }
    if (typeof field !== "string" || typeof message !== "string") return []
    return [{ field, message }]
  })
}
