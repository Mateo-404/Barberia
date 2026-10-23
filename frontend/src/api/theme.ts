import { useQuery } from "@tanstack/react-query"
import { api } from "@/lib/api-client"
import { useTenant } from "@/context/TenantContext"
import type { Theme } from "@/api/settings"
import type { components } from "@/types/api"

export type PresetView = components["schemas"]["PresetView"]
export type PreviewResult = components["schemas"]["PreviewResult"]
export type SchemeView = components["schemas"]["SchemeView"]

/** Curated starting points with their generated palettes. */
export function usePresets() {
  const tenant = useTenant()
  return useQuery({
    queryKey: ["theme-presets", tenant] as const,
    queryFn: ({ signal }) => api<PresetView[]>("/api/admin/theme/presets", { signal }),
    staleTime: Infinity,
  })
}

/** Renders an unsaved theme on the server: CSS, palettes and contrast. */
export function previewTheme(theme: Theme, signal?: AbortSignal): Promise<PreviewResult> {
  return api<PreviewResult>("/api/admin/theme/preview", { method: "POST", body: JSON.stringify(theme), signal })
}
