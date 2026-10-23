import { lazy, Suspense } from "react"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { BrowserRouter, Routes, Route, useLocation } from "react-router-dom"
import { useTranslation } from "react-i18next"
import { AuthProvider } from "@/context/AuthContext"
import { ThemeProvider } from "@/context/ThemeContext"
import { TenantProvider } from "@/context/TenantContext"
import { ApiError } from "@/lib/api-client"
import Booking from "@/pages/Booking"
import Login from "@/pages/Login"
import AuthCallback from "@/pages/AuthCallback"
import ProtectedRoute from "@/components/ProtectedRoute"
import AdminLayout from "@/components/admin/AdminLayout"
import { ShortcutsHelp } from "@/components/ShortcutsHelp"
import { PreviewBridge } from "@/components/PreviewBridge"

const Dashboard = lazy(() => import("@/pages/admin/Dashboard"))
const Appointments = lazy(() => import("@/pages/admin/Appointments"))
const Stats = lazy(() => import("@/pages/admin/Stats"))
const Welcome = lazy(() => import("@/pages/theme/Welcome"))
const ThemeEditorPage = lazy(() => import("@/pages/theme/ThemeEditorPage"))
const BusinessPage = lazy(() => import("@/pages/admin/BusinessPage"))
const SettingsPage = lazy(() => import("@/pages/admin/SettingsPage"))
const SecurityPage = lazy(() => import("@/pages/admin/SecurityPage"))

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 60_000,
      // Retry only transient failures (network, timeout, 5xx). A 4xx does not
      // fix itself, and retrying it would hide real errors.
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
          <ThemeProvider>
          <AuthProvider>
            <Routes>
              <Route path="/" element={<Booking />} />
              <Route path="/login" element={<Login />} />
              <Route path="/auth/google/callback" element={<AuthCallback />} />
              <Route element={<ProtectedRoute />}>
                <Route path="/welcome" element={<Lazy><Welcome /></Lazy>} />
                <Route path="/admin/theme" element={<Lazy><ThemeEditorPage /></Lazy>} />
                <Route path="/admin/security" element={<Lazy><SecurityPage /></Lazy>} />
                <Route element={<AdminLayout />}>
                  <Route path="/admin" element={<Lazy><Dashboard /></Lazy>} />
                  <Route path="/admin/appointments" element={<Lazy><Appointments /></Lazy>} />
                  <Route path="/admin/settings" element={<Lazy><SettingsPage /></Lazy>} />
                  <Route path="/admin/business" element={<Lazy><BusinessPage /></Lazy>} />
                  <Route path="/admin/stats" element={<Lazy><Stats /></Lazy>} />
                </Route>
              </Route>
            </Routes>
            <Shell />
            <PreviewBridge />
          </AuthProvider>
          </ThemeProvider>
        </BrowserRouter>
      </TenantProvider>
    </QueryClientProvider>
  )
}

export default App
