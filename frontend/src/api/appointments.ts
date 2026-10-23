import { useMutation, useQuery, useQueryClient, type QueryClient } from "@tanstack/react-query"
import { api } from "@/lib/api-client"
import { queryKeys, type AppointmentFilters } from "@/lib/query-keys"
import { useTenant } from "@/context/TenantContext"
import type { components } from "@/types/api"

type Appointment = components["schemas"]["Appointment"]
type AppointmentCreate = components["schemas"]["AppointmentCreate"]
type AppointmentPatch = components["schemas"]["AppointmentPatch"]
type AppointmentPage = components["schemas"]["AppointmentPage"]

/** Booked start times: the availability grid is rebuilt from them. */
export function useBookedSlots() {
  const tenant = useTenant()
  return useQuery({
    queryKey: queryKeys.availability(tenant),
    queryFn: ({ signal }) => api<string[]>("/api/appointments/booked-slots", { signal }),
  })
}

export function useCreateAppointment() {
  const tenant = useTenant()
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (data: AppointmentCreate) =>
      api<Appointment>("/api/appointments", { method: "POST", body: JSON.stringify(data) }),
    onSuccess: () => invalidateAppointments(queryClient, tenant),
  })
}

export function useAppointments() {
  const tenant = useTenant()
  return useQuery({
    queryKey: queryKeys.appointments.all(tenant),
    queryFn: ({ signal }) => api<Appointment[]>("/api/appointments", { signal }),
  })
}

export function useAppointmentPage(page: number, size: number) {
  const tenant = useTenant()
  return useQuery({
    queryKey: queryKeys.appointments.page(tenant, page, size),
    queryFn: ({ signal }) =>
      api<AppointmentPage>(`/api/appointments/paged?page=${page}&size=${size}`, { signal }),
  })
}

/** The most recent appointments (the API orders by id). */
export function useLatestAppointments(size = 10) {
  const tenant = useTenant()
  return useQuery({
    queryKey: queryKeys.appointments.latest(tenant, size),
    queryFn: async ({ signal }) =>
      (await api<AppointmentPage>(`/api/appointments/paged?page=0&size=${size}`, { signal })).items,
  })
}

export function useUpdateAppointment() {
  const tenant = useTenant()
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, patch }: { id: number; patch: AppointmentPatch }) =>
      api<Appointment>(`/api/appointments/${id}`, { method: "PATCH", body: JSON.stringify(patch) }),
    onSuccess: () => invalidateAppointments(queryClient, tenant),
  })
}

export function useDeleteAppointment() {
  const tenant = useTenant()
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: number) => api<void>(`/api/appointments/${id}`, { method: "DELETE" }),
    onSuccess: () => invalidateAppointments(queryClient, tenant),
  })
}

/** Refreshes everything that depends on this business's appointments. */
function invalidateAppointments(queryClient: QueryClient, tenant: string) {
  void queryClient.invalidateQueries({ queryKey: queryKeys.appointments.all(tenant) })
  void queryClient.invalidateQueries({ queryKey: queryKeys.availability(tenant) })
  void queryClient.invalidateQueries({ queryKey: queryKeys.stats.dashboard(tenant) })
}

export type { Appointment, AppointmentCreate, AppointmentPatch, AppointmentFilters }
