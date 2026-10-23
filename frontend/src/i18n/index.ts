import i18n from "i18next"
import { initReactI18next } from "react-i18next"
import common from "@/locales/es/common.json"
import errors from "@/locales/es/errors.json"
import booking from "@/locales/es/booking.json"
import admin from "@/locales/es/admin.json"
import theme from "@/locales/es/theme.json"
import shortcuts from "@/locales/es/shortcuts.json"

/** Spanish is the only shipped language; add a folder under locales/ and an
 *  entry here to ship another. Keys are typed from the Spanish catalog. */
export const resources = {
  es: { common, errors, booking, admin, theme, shortcuts },
} as const

void i18n.use(initReactI18next).init({
  resources,
  lng: "es",
  fallbackLng: "es",
  defaultNS: "common",
  ns: Object.keys(resources.es),
  interpolation: { escapeValue: false },
  returnNull: false,
})

export default i18n
