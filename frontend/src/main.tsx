import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './index.css'
import './i18n'
import App from './App.tsx'
import { ThemeProvider } from './context/ThemeContext'

async function bootstrap() {
  try {
    const base = import.meta.env.VITE_API_BASE_URL ?? "http://localhost:8080"
    const res = await fetch(`${base}/api/public/theme.css`, { cache: "no-cache" })
    if (res.ok) {
      const css = await res.text()
      const style = document.createElement("style")
      style.id = "tenant-theme"
      style.textContent = css
      document.head.appendChild(style)
    }
  } catch {
  }
  createRoot(document.getElementById('root')!).render(
    <StrictMode>
      <ThemeProvider>
        <App />
      </ThemeProvider>
    </StrictMode>,
  )
}

bootstrap()
