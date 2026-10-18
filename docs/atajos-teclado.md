# Atajos de teclado

La app React (`frontend/`) tiene un sistema de atajos disponible en todo el sitio. Cada vista registra sus propios atajos y se muestran en un panel de ayuda con **`Shift + ?`** (o `Shift + /`).

## Comportamiento general

- Las **teclas simples** (ej. `r`, `l`, `1`) solo disparan si el foco NO está en un input, textarea, select o contenido editable.
- Los **combos con modificadores** (ej. `Shift+g`, `Ctrl+Enter`, `Alt+d`) disparan siempre, salvo que la vista lo indique.
- Cuando hay un **diálogo/modal abierto**, los atajos de la vista se desactivan (excepto los de navegación del panel admin, que ignoran el foco en diálogos).
- `Escape` cierra el panel de ayuda y los modales.

## Cliente — wizard de reserva

| Atajo | Acción |
|---|---|
| `1` … `6` | Seleccionar servicio / horario |
| `Enter` o `Shift+n` | Siguiente paso |
| `←`, `Backspace`, `Shift+p` o `Shift+l` | Paso anterior |
| `Esc` | Limpiar error / volver |
| `r` | Reservar otro turno (pantalla de confirmación) |

## Administrador — login

| Atajo | Acción |
|---|---|
| `Ctrl+Enter` | Iniciar sesión |
| `Alt+d` | Mostrar/ocultar contraseña |
| `Esc` | Limpiar error |

## Administrador — navegación global (panel)

| Atajo | Acción |
|---|---|
| `Shift+g` | Ir al Panel |
| `Shift+e` | Ir a Estadísticas |
| `Shift+t` | Ir a Turnos |
| `Shift+h` | Ir al sitio público |
| `Ctrl+Shift+q` | Cerrar sesión |
| `Shift+r` | Recargar datos del panel |

## Administrador — Estadísticas

| Atajo | Acción |
|---|---|
| `r` | Refrescar datos |

## Administrador — Turnos

| Atajo | Acción |
|---|---|
| `l` / `c` | Vista lista / calendario |
| `/` | Buscar |
| `1` / `2` / `3` / `4` | Filtro: todos / confirmados / pendientes / hoy |
| `x` | Limpiar filtros |
| `←` / `→` | Mes anterior / siguiente (calendario) |
| `h` | Ir a hoy (calendario) |
| `n` / `p` | Día siguiente / anterior (calendario) |
| `j` / `k` | Seleccionar día siguiente / anterior |
| `e` | Editar el primer turno del día |
| `Delete` | Eliminar el primer turno |
| `Enter` | Abrir el cliente del turno (WhatsApp) |
| `Esc` | Cerrar panel / modal |

## Cómo agregar un atajo (dev)

Dos piezas en `frontend/src/hooks/`:

```ts
import { useHotkeys } from "@/hooks/useHotkeys"
import { useRegisterShortcuts } from "@/hooks/useShortcutsRegistry"

// 1) Definir el comportamiento
useHotkeys(
  { "Shift+r": () => refetchAll() },
  enabled,                          // desactivar condicionalmente
  { ignoreWhenDialogOpen: true },   // opcional
)

// 2) Registrarlo en el panel de ayuda (scope: "global" | "cliente" | "admin")
useRegisterShortcuts("admin", [
  { key: "Shift+r", description: "Recargar datos del panel" },
])
```

- `useHotkeys` acepta specs tipo `"Ctrl+Shift+q"`, `"Shift+?"` o teclas simples (`"r"`, `"Delete"`, `"Escape"`, `"ArrowLeft"`).
- El registro por scope hace que el panel de ayuda (`ShortcutsHelp`) muestre solo los atajos vigentes en la vista actual. **Mantené ambos en sincronía**: si agregás un atajo, registrá su descripción.
