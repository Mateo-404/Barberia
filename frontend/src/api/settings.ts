/**
 * Site settings of the business: the theme draft and its publishing, plus the
 * owner's preferences for the admin panel. Every route lives under
 * `/api/admin/*` and needs the admin's Bearer, which the api-client adds.
 */

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { api } from "@/lib/api-client"
import { queryKeys } from "@/lib/query-keys"
import { useTenant } from "@/context/TenantContext"
import type { components } from "@/types/api"

export type Theme = components["schemas"]["Theme"]
export type SettingsState = components["schemas"]["SettingsState"]
export type Preferences = components["schemas"]["Preferences"]
export type DraftResult = components["schemas"]["DraftResult"]
export type PublishResult = components["schemas"]["PublishResult"]
export type SettingsVersion = components["schemas"]["SettingsVersion"]

/** Query options of the settings state (shared with `fetchQuery`). */
export function settingsQueryOptions(tenant: string) {
  return {
    queryKey: queryKeys.settings.admin(tenant),
    queryFn: ({ signal }: { signal: AbortSignal }) => api<SettingsState>("/api/admin/settings", { signal }),
  }
}

/**
 * Draft, preferences and publish state. `enabled: false` avoids asking
 * without a session (e.g. on the login screen).
 */
export function useAdminSettings(options: { enabled?: boolean } = {}) {
  const tenant = useTenant()
  return useQuery({ ...settingsQueryOptions(tenant), enabled: options.enabled ?? true })
}

/** Saves the draft. The public site does not change until it is published. */
export function useSaveDraft() {
  const tenant = useTenant()
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (theme: Theme) =>
      api<DraftResult>("/api/admin/settings/draft", { method: "PUT", body: JSON.stringify(theme) }),
    onSuccess: () => void queryClient.invalidateQueries({ queryKey: queryKeys.settings.admin(tenant) }),
  })
}

/** Publishes the draft as version N+1 and marks the business as onboarded. */
export function usePublish() {
  const tenant = useTenant()
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: () => api<PublishResult>("/api/admin/settings/publish", { method: "POST" }),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: queryKeys.settings.admin(tenant) })
      void queryClient.invalidateQueries({ queryKey: queryKeys.settings.versions(tenant) })
      void queryClient.invalidateQueries({ queryKey: ["public-settings", tenant] })
    },
  })
}

/** Last published versions. */
export function useVersions(enabled = true) {
  const tenant = useTenant()
  return useQuery({
    queryKey: queryKeys.settings.versions(tenant),
    queryFn: ({ signal }) => api<SettingsVersion[]>("/api/admin/settings/versions", { signal }),
    enabled,
  })
}

/** Restores a version: the backend publishes a new one with that content. */
export function useRollback() {
  const tenant = useTenant()
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (version: number) =>
      api<PublishResult>("/api/admin/settings/rollback", { method: "POST", body: JSON.stringify({ version }) }),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: queryKeys.settings.admin(tenant) })
      void queryClient.invalidateQueries({ queryKey: queryKeys.settings.versions(tenant) })
      void queryClient.invalidateQueries({ queryKey: ["public-settings", tenant] })
    },
  })
}

/** Switches the theme editor menu and the welcome wizard on or off. */
export function useUpdatePreferences() {
  const tenant = useTenant()
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (patch: Partial<Preferences>) =>
      api<Preferences>("/api/admin/settings/preferences", { method: "PATCH", body: JSON.stringify(patch) }),
    onSuccess: () => void queryClient.invalidateQueries({ queryKey: queryKeys.settings.admin(tenant) }),
  })
}
