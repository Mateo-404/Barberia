import { useEffect, useMemo, useRef, useState } from "react"
import { useTranslation } from "react-i18next"
import { Link, NavLink, Outlet, useNavigate } from "react-router-dom"
import { useAuth } from "@/context/AuthContext"
import { useAdminSettings } from "@/api/settings"
import { useSiteIdentity } from "@/api/site"
import { cn } from "@/lib/utils"
import { SiteMark } from "@/components/SiteMark"
import { ThemeToggle } from "@/components/ThemeToggle"
import { useHotkeys } from "@/hooks/useHotkeys"
import { useRegisterShortcuts } from "@/hooks/useShortcutsRegistry"
import { ArrowLeft, Settings, LogOut, Store, LayoutDashboard } from "lucide-react"

const NAV_ITEMS = [
  { to: "/admin", key: "dashboard", hotkey: "Shift+g" },
  { to: "/admin/stats", key: "stats", hotkey: "Shift+e" },
  { to: "/admin/appointments", key: "appointments", hotkey: "Shift+t" },
  { to: "/admin/theme", key: "theme", hotkey: "Shift+m" },
  { to: "/admin/settings", key: "settings", hotkey: "Shift+s" },
] as const

export default function AdminLayout() {
  const { t } = useTranslation("admin")
  const { t: tc } = useTranslation()
  const { t: ts } = useTranslation("shortcuts")
  const { admin, logout } = useAuth()
  const { data: settingsState } = useAdminSettings()
  const navigate = useNavigate()
  const logoSrc = useSiteIdentity().logo
  const themeEditorEnabled = settingsState?.preferences.themeEditorEnabled !== false
  const [menuOpen, setMenuOpen] = useState(false)
  const menuRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (!menuOpen) return
    function onClick(e: MouseEvent) {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) setMenuOpen(false)
    }
    function onKey(e: KeyboardEvent) {
      if (e.key === "Escape") setMenuOpen(false)
    }
    document.addEventListener("mousedown", onClick)
    document.addEventListener("keydown", onKey)
    return () => {
      document.removeEventListener("mousedown", onClick)
      document.removeEventListener("keydown", onKey)
    }
  }, [menuOpen])

  // The theme editor menu is optional (Settings → Theme).
  const navItems = useMemo(
    () => NAV_ITEMS.filter((item) => item.key !== "theme" || themeEditorEnabled),
    [themeEditorEnabled],
  )

  const hotkeyMap = useMemo(
    () => ({
      "Shift+g": () => navigate("/admin"),
      "Shift+e": () => navigate("/admin/stats"),
      "Shift+t": () => navigate("/admin/appointments"),
      ...(themeEditorEnabled ? { "Shift+m": () => navigate("/admin/theme") } : {}),
      "Shift+s": () => navigate("/admin/settings"),
      "Shift+h": () => navigate("/"),
      "Ctrl+Shift+q": () => {
        logout()
        navigate("/login")
      },
    }),
    [navigate, logout, themeEditorEnabled],
  )

  useHotkeys(hotkeyMap, true, { ignoreWhenDialogOpen: true })

  const navDefs = useMemo(
    () => [
      ...navItems.map((item) => ({ key: item.hotkey, description: ts(`admin.go.${item.key}`) })),
      { key: "Shift+h", description: ts("admin.go.site") },
      { key: "Ctrl+Shift+q", description: ts("admin.logout") },
    ],
    [ts, navItems],
  )
  useRegisterShortcuts("admin", navDefs)

  function handleLogout() {
    logout()
    navigate("/login")
  }

  const initial = admin?.firstName?.charAt(0)?.toUpperCase() ?? "U"

  return (
    <div className="min-h-screen bg-background flex flex-col">
      <nav
        className="flex items-center justify-between px-4"
        style={{ height: "var(--nav-height)" }}
      >
        <button onClick={() => navigate("/")} className="bg-transparent border-0 p-0 cursor-pointer rounded focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-focus-ring">
          <ArrowLeft className="h-6 w-6 text-foreground" aria-label={tc("back")} />
        </button>

        <Link to="/" className="inline-flex items-center">
          <SiteMark />
        </Link>

        <ThemeToggle />
        <div className="relative" ref={menuRef}>
          <button
            onClick={() => setMenuOpen((v) => !v)}
            className="inline-flex h-10 w-10 items-center justify-center overflow-hidden rounded-full bg-primary/15 text-primary ring-1 ring-primary/30 transition-transform hover:scale-105 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-focus-ring"
            aria-label={t("layout.accountMenu")}
            aria-expanded={menuOpen}
          >
            {logoSrc ? (
              <img src={logoSrc} alt="Logo" className="h-full w-full object-cover" />
            ) : (
              <span className="text-sm font-semibold">{initial}</span>
            )}
          </button>
          {menuOpen && (
            <div className="absolute right-0 top-12 z-50 w-64 origin-top-right rounded-lg border border-border/60 bg-card/95 p-1.5 shadow-2xl backdrop-blur-md animate-in fade-in zoom-in-95 slide-in-from-top-2 duration-200">
              <div className="px-3 py-2">
                <p className="text-sm font-medium text-foreground">{admin?.firstName} {admin?.lastName}</p>
                <p className="truncate text-xs text-muted-foreground">{admin?.email}</p>
              </div>
              <div className="my-1 h-px bg-border" />
              <button
                onClick={() => { setMenuOpen(false); navigate("/admin") }}
                className="flex w-full items-center gap-2 rounded-sm px-3 py-2 text-sm text-foreground transition-all duration-150 hover:translate-x-0.5 hover:bg-muted animate-in fade-in slide-in-from-left-1 delay-75"
              >
                <LayoutDashboard className="h-4 w-4" /> {t("nav.dashboard")}
              </button>
              <button
                onClick={() => { setMenuOpen(false); navigate("/admin/settings") }}
                className="flex w-full items-center gap-2 rounded-sm px-3 py-2 text-sm text-foreground transition-all duration-150 hover:translate-x-0.5 hover:bg-muted animate-in fade-in slide-in-from-left-1 delay-100"
              >
                <Settings className="h-4 w-4" /> {t("nav.settings")}
              </button>
              <button
                onClick={() => { setMenuOpen(false); navigate("/") }}
                className="flex w-full items-center gap-2 rounded-sm px-3 py-2 text-sm text-foreground transition-all duration-150 hover:translate-x-0.5 hover:bg-muted animate-in fade-in slide-in-from-left-1 delay-150"
              >
                <Store className="h-4 w-4" /> {t("layout.viewSite")}
              </button>
              <div className="my-1 h-px bg-border" />
              <button
                onClick={() => { setMenuOpen(false); handleLogout() }}
                className="flex w-full items-center gap-2 rounded-sm px-3 py-2 text-sm text-destructive transition-all duration-150 hover:translate-x-0.5 hover:bg-destructive/10 animate-in fade-in slide-in-from-left-1 delay-200"
              >
                <LogOut className="h-4 w-4" /> {t("layout.logout")}
              </button>
            </div>
          )}
        </div>
      </nav>

      <div className="flex flex-col gap-2 px-4 mb-4">
        <nav className="flex gap-1 bg-card rounded-xl p-1 border border-border overflow-x-auto">
          {navItems.map((item) => (
            <NavLink
              key={item.to}
              to={item.to}
              end={item.to === "/admin"}
              className={({ isActive }) =>
                cn(
                  "flex-1 text-center px-4 py-2 rounded-lg text-sm font-medium whitespace-nowrap transition-colors",
                  isActive
                    ? "bg-primary text-primary-foreground"
                    : "text-muted-foreground hover:text-foreground",
                )
              }
            >
              {t(`nav.${item.key}`)}
            </NavLink>
          ))}
        </nav>
        <div className="flex flex-wrap items-center justify-center gap-x-3 gap-y-1 text-[11px] text-muted-foreground mt-1">
          {navItems.map((item) => (
            <span key={item.to}>
              <kbd className="px-1.5 py-0.5 rounded bg-muted text-[10px] font-mono">{item.hotkey}</kbd> {t(`nav.${item.key}`)}
            </span>
          ))}
        </div>
      </div>

      <main className="flex-1 px-4 pb-8">
        <Outlet />
      </main>
    </div>
  )
}
