import { lazy, Suspense } from "react"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { BrowserRouter, Routes, Route } from "react-router-dom"
import { AuthProvider } from "@/context/AuthContext"
import Reserva from "@/pages/Reserva"
import Login from "@/pages/Login"
import ProtectedRoute from "@/components/ProtectedRoute"
import AdminLayout from "@/components/admin/AdminLayout"

const AdminDashboard = lazy(() => import("@/pages/admin/AdminDashboard"))

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 60_000,
    },
  },
})

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
              </Route>
            </Route>
          </Routes>
        </AuthProvider>
      </BrowserRouter>
    </QueryClientProvider>
  )
}

export default App
