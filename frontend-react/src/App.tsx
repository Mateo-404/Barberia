import { lazy, Suspense } from "react"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { BrowserRouter, Routes, Route, Navigate, useLocation } from "react-router-dom"
import { AuthProvider } from "@/context/AuthContext"
import Reserva from "@/pages/Reserva"
import Login from "@/pages/Login"
import ProtectedRoute from "@/components/ProtectedRoute"
import AdminLayout from "@/components/admin/AdminLayout"
import { ShortcutsHelp } from "@/components/ShortcutsHelp"

const AdminDashboard = lazy(() => import("@/pages/admin/AdminDashboard"))
const TurnosAdmin = lazy(() => import("@/pages/admin/TurnosAdmin"))
const EstadisticasAdmin = lazy(() => import("@/pages/admin/EstadisticasAdmin"))

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 60_000,
    },
  },
})

function Shell() {
  const { pathname } = useLocation()
  const scope: "cliente" | "admin" = pathname.startsWith("/admin")
    ? "admin"
    : "cliente"

  return <ShortcutsHelp scope={scope} />
}

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <AuthProvider>
          <Routes>
            <Route path="/" element={<Reserva />} />
            <Route path="/login" element={<Login />} />
            <Route element={<ProtectedRoute />}>
              <Route element={<AdminLayout />}>
                <Route
                  path="/admin"
                  element={
                    <Suspense
                      fallback={
                        <div className="p-6 text-sm text-muted-foreground">
                          Cargando panel…
                        </div>
                      }
                    >
                      <AdminDashboard />
                    </Suspense>
                  }
                />
                <Route
                  path="/admin/turnos"
                  element={
                    <Suspense
                      fallback={
                        <div className="p-6 text-sm text-muted-foreground">
                          Cargando turnos…
                        </div>
                      }
                    >
                      <TurnosAdmin />
                    </Suspense>
                  }
                />
                <Route
                  path="/admin/estadisticas"
                  element={
                    <Suspense
                      fallback={
                        <div className="p-6 text-sm text-muted-foreground">
                          Cargando estadísticas…
                        </div>
                      }
                    >
                      <EstadisticasAdmin />
                    </Suspense>
                  }
                />
              </Route>
            </Route>
            <Route
              path="/estadisticas/panel"
              element={<Navigate to="/admin/estadisticas" replace />}
            />
          </Routes>
          <Shell />
        </AuthProvider>
      </BrowserRouter>
    </QueryClientProvider>
  )
}

export default App
