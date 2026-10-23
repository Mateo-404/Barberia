import { StrictMode } from "react"
import { createRoot } from "react-dom/client"
import "./index.css"
import "./i18n"
import App from "./App.tsx"
import { loadBusinessConfig } from "./lib/business-config"
import { loadPublicTheme } from "./lib/tenant-theme"

// The theme and business hours are applied before React mounts so the first paint already wears
// the business's colors (no flash of the default theme).
void Promise.all([loadPublicTheme(), loadBusinessConfig()]).finally(() => {
  createRoot(document.getElementById("root")!).render(
    <StrictMode>
      <App />
    </StrictMode>,
  )
})
