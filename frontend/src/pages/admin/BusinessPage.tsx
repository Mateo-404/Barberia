import { useEffect, useState, type FormEvent } from "react"
import { useTranslation } from "react-i18next"
import { Link } from "react-router-dom"
import { useBusiness, useSaveBusiness, type Business } from "@/api/business"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { errorMessage, fieldErrors } from "@/lib/error-message"

const CURRENCIES = ["ARS", "USD", "EUR", "UYU", "CLP"]
const TIME_ZONES = [
  "America/Argentina/Buenos_Aires", "America/Montevideo", "America/Santiago", "America/Bogota",
  "America/Lima", "America/Mexico_City", "America/Sao_Paulo", "Europe/Madrid", "UTC",
]

/** Operating data of the business: contact, currency and opening hours. */
export default function BusinessPage() {
  const { t } = useTranslation("admin")
  const { t: tc } = useTranslation()
  const { data, isLoading, isError } = useBusiness()
  const save = useSaveBusiness()
  const [form, setForm] = useState<Business | null>(null)
  const [saved, setSaved] = useState(false)

  useEffect(() => {
    if (data) setForm(data)
  }, [data])

  if (isLoading || !form) {
    return <p className="p-6 text-sm text-muted-foreground">{isError ? t("business.loadError") : tc("loading")}</p>
  }

  const errors = fieldErrors(tc, save.error)
  const set = (patch: Partial<Business>) => {
    setSaved(false)
    setForm({ ...form, ...patch })
  }

  async function onSubmit(event: FormEvent) {
    event.preventDefault()
    try {
      await save.mutateAsync(form as Business)
      setSaved(true)
    } catch {
      // Field errors and the general message render below.
    }
  }

  const field = (id: keyof Business, label: string, props: React.ComponentProps<typeof Input> = {}) => (
    <div className="space-y-1.5">
      <Label htmlFor={`business-${id}`}>{label}</Label>
      <Input
        id={`business-${id}`}
        value={String(form[id])}
        onChange={(e) => set({ [id]: props.type === "number" ? Number(e.target.value) : e.target.value } as Partial<Business>)}
        aria-invalid={errors[id] ? true : undefined}
        className="border border-border"
        {...props}
      />
      {errors[id] && <p role="alert" className="text-xs text-destructive">{errors[id]}</p>}
    </div>
  )

  const selectClass = "w-full rounded-xl border border-border bg-input px-4 py-2.5 text-sm text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-focus-ring"
  const zones = TIME_ZONES.includes(form.tz) ? TIME_ZONES : [form.tz, ...TIME_ZONES]

  return (
    <form onSubmit={(e) => void onSubmit(e)} className="mx-auto w-full max-w-2xl space-y-8 px-4 py-6 sm:px-6">
      <header className="space-y-1 border-b border-border/60 pb-6">
        <Link to="/admin/settings" className="text-sm text-muted-foreground hover:text-foreground">
          {t("business.back")}
        </Link>
        <h1 className="text-2xl font-semibold tracking-tight text-foreground sm:text-3xl">{t("business.title")}</h1>
        <p className="text-sm text-muted-foreground">{t("business.subtitle")}</p>
      </header>

      <fieldset className="space-y-4">
        <legend className="mb-2 text-base font-semibold text-foreground">{t("business.contact")}</legend>
        {field("whatsapp", t("business.whatsapp"), { placeholder: t("business.whatsappPlaceholder"), inputMode: "tel" })}
        {field("instagram", t("business.instagram"), { placeholder: t("business.instagramPlaceholder") })}
        {field("address", t("business.address"))}
        <div className="space-y-1.5">
          <Label htmlFor="business-currency">{t("business.currency")}</Label>
          <select id="business-currency" value={form.currency} onChange={(e) => set({ currency: e.target.value as Business["currency"] })} className={selectClass}>
            {CURRENCIES.map((c) => <option key={c} value={c}>{c}</option>)}
          </select>
        </div>
      </fieldset>

      <fieldset className="space-y-4">
        <legend className="mb-2 text-base font-semibold text-foreground">{t("business.hours")}</legend>
        <div className="grid grid-cols-2 gap-4">
          {field("open", t("business.open"), { type: "time" })}
          {field("close", t("business.close"), { type: "time" })}
          {field("slot", t("business.slot"), { type: "number", min: 10, max: 240, step: 5 })}
          {field("notice", t("business.notice"), { type: "number", min: 0, max: 168 })}
        </div>
        <div className="space-y-1.5">
          <Label htmlFor="business-tz">{t("business.tz")}</Label>
          <select id="business-tz" value={form.tz} onChange={(e) => set({ tz: e.target.value })} className={selectClass}>
            {zones.map((z) => <option key={z} value={z}>{z.replace(/_/g, " ")}</option>)}
          </select>
        </div>
      </fieldset>

      <div className="flex flex-wrap items-center gap-3">
        <Button type="submit" disabled={save.isPending}>
          {save.isPending ? t("business.saving") : t("business.save")}
        </Button>
        {saved && <p role="status" className="text-sm text-muted-foreground">{t("business.saved")}</p>}
        {save.isError && Object.keys(errors).length === 0 && (
          <p role="alert" className="text-sm text-destructive">{errorMessage(tc, save.error)}</p>
        )}
      </div>
    </form>
  )
}
