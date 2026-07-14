import { BrowserRouter, Routes, Route } from "react-router-dom"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import Reserva from "@/pages/Reserva"

const queryClient = new QueryClient()

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <Routes>
          <Route path="/" element={<Reserva />} />
        </Routes>
      </BrowserRouter>
    </QueryClientProvider>
  )
}

export default App
