import { lazy, Suspense } from "react"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { BrowserRouter, Routes, Route, useLocation } from "react-router-dom"
import { useTranslation } from "react-i18next"
import { AuthProvider } from "@/context/AuthContext"
import { TenantProvider } from "@/context/TenantContext"
import { ApiError } from "@/lib/api-client"
import Booking from "@/pages/Booking"
import Login from "@/pages/Login"
import Onboarding from "@/pages/Onboarding"
import ProtectedRoute from "@/components/ProtectedRoute"
import AdminLayout from "@/components/admin/AdminLayout"
import { ShortcutsHelp } from "@/components/ShortcutsHelp"

const Dashboard = lazy(() => import("@/pages/admin/Dashboard"))
const Appointments = lazy(() => import("@/pages/admin/Appointments"))
const Stats = lazy(() => import("@/pages/admin/Stats"))
const MarcaAdmin = lazy(() => import("@/pages/admin/MarcaAdmin"))
const SettingsPage = lazy(() => import("@/pages/admin/SettingsPage"))

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 60_000,
      // Reintentar sólo lo transitorio (red, timeout, 5xx). Un 4xx no se
      // recupera solo y reintentarlo enmascara errores reales.
      retry: (failureCount, error) => {
        if (failureCount >= 2) return false
        if (error instanceof ApiError) return error.isRetriable
        return true
      },
    },
  },
})

/** Suspense boundary for lazily loaded admin pages. */
function Lazy({ children }: { children: React.ReactNode }) {
  const { t } = useTranslation()
  return <Suspense fallback={<div className="p-6 text-sm text-muted-foreground">{t("loading")}</div>}>{children}</Suspense>
}

function Shell() {
  const { pathname } = useLocation()
  const scope: "customer" | "admin" = pathname.startsWith("/admin")
    ? "admin"
    : "customer"

  return <ShortcutsHelp scope={scope} />
}

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <TenantProvider slug={import.meta.env.VITE_TENANT_SLUG ?? "default"}>
        <BrowserRouter>
          <AuthProvider>
            <Routes>
              <Route path="/" element={<Booking />} />
              <Route path="/login" element={<Login />} />
              <Route element={<ProtectedRoute />}>
                <Route path="/onboarding" element={<Onboarding />} />
                <Route element={<AdminLayout />}>
                  <Route path="/admin" element={<Lazy><Dashboard /></Lazy>} />
                  <Route path="/admin/appointments" element={<Lazy><Appointments /></Lazy>} />
                  <Route path="/admin/theme" element={<Lazy><MarcaAdmin /></Lazy>} />
                  <Route path="/admin/settings" element={<Lazy><SettingsPage /></Lazy>} />
                  <Route path="/admin/stats" element={<Lazy><Stats /></Lazy>} />
                </Route>
              </Route>
            </Routes>
            <Shell />
          </AuthProvider>
        </BrowserRouter>
      </TenantProvider>
    </QueryClientProvider>
  )
}

export default App
