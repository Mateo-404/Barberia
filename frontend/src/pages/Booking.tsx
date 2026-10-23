import { useState, useMemo, useCallback } from "react"
import { Link } from "react-router-dom"
import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { useTranslation } from "react-i18next"
import { useServices } from "@/api/services"
import { useBookedSlots, useCreateAppointment } from "@/api/appointments"
import { appointmentSchema, type AppointmentFormData } from "@/lib/schemas/appointment-schema"
import { errorMessage, fieldErrorMessage } from "@/lib/error-message"
import { formatCurrency, formatDate } from "@/lib/format"
import { businessConfig, timeSlots } from "@/lib/business-config"
import { ThemeToggle } from "@/components/ThemeToggle"
import { todayInBusinessTz } from "@/lib/datetime"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { useHotkeys } from "@/hooks/useHotkeys"
import { useRegisterShortcuts } from "@/hooks/useShortcutsRegistry"

const STEPS = ["service", "dateTime", "details"] as const

function SuccessView({ onReset }: { onReset: () => void }) {
  const { t } = useTranslation("booking")
  return (
  <main className="min-h-screen flex flex-col items-center justify-center p-6 bg-background">
    <div className="w-full max-w-md text-center space-y-6 page-enter">
      <div className="w-20 h-20 rounded-full bg-primary flex items-center justify-center mx-auto" style={{ animation: "pop 0.7s ease-out" }}>
        <svg className="w-10 h-10 text-primary-foreground" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={3} d="M5 13l4 4L19 7" />
        </svg>
      </div>
      <h1 className="text-2xl font-bold text-foreground">{t("success.title")}</h1>
      <p className="text-muted-foreground">{t("success.message")}</p>
      <Button variant="secondary" onClick={onReset}>
        {t("success.again")}
      </Button>
    </div>
  </main>
  )
}

export default function Booking() {
  const { t } = useTranslation("booking")
  const { t: tc } = useTranslation()
  const [step, setStep] = useState(0)
  const [success, setSuccess] = useState(false)
  const [selectedDate, setSelectedDate] = useState("")
  const [selectedTime, setSelectedTime] = useState("")

  const {
    data: services,
    isLoading: loadingServices,
    isError: servicesFailed,
    refetch: refetchServices,
  } = useServices()
  const { data: bookedSlots } = useBookedSlots()
  const createAppointment = useCreateAppointment()

  const {
    register,
    handleSubmit,
    formState: { errors },
    setValue,
    trigger,
    setError,
    watch,
    reset,
  } = useForm<AppointmentFormData>({
    resolver: zodResolver(appointmentSchema),
    defaultValues: {
      serviceId: undefined as unknown as number,
      startsAt: "",
      customerFirstName: "",
      customerLastName: "",
      customerPhone: "",
      customerEmail: "",
    },
    mode: "onChange",
  })

  const selectedServiceId = watch("serviceId")

  // Slot grid from the (interim) business configuration.
  const slots = useMemo(() => timeSlots(), [])

  const selectedService = useMemo(() => {
    if (!services || !selectedServiceId) return null
    return services.find((s) => s.id === selectedServiceId) ?? null
  }, [services, selectedServiceId])

  const isSlotOccupied = useCallback(
    (time: string) => {
      if (!selectedDate || !bookedSlots) return false
      return bookedSlots.includes(`${selectedDate}T${time}`)
    },
    [selectedDate, bookedSlots],
  )

  const handleNext = useCallback(async () => {
    if (step === 0) {
      const valid = await trigger("serviceId")
      if (valid) setStep(1)
    } else if (step === 1) {
      if (!selectedDate || !selectedTime) {
        setError("startsAt", { message: "required" })
        return
      }
      const isoStr = `${selectedDate}T${selectedTime}:00`
      setValue("startsAt", isoStr, { shouldValidate: true })
      const valid = await trigger("startsAt")
      if (valid) setStep(2)
    }
  }, [step, selectedDate, selectedTime, trigger, setValue, setError])

  const handleBack = useCallback(() => {
    setStep((s) => Math.max(s - 1, 0))
  }, [])

  const handleReset = useCallback(() => {
    setSuccess(false)
    setStep(0)
    setSelectedDate("")
    setSelectedTime("")
    reset()
  }, [reset])

  const hotkeyMap = useMemo(() => {
    const map: Record<string, () => void> = {
      Escape: () => {
        if (errors.root) {
          setError("root", { message: "" })
        } else if (step > 0) {
          handleBack()
        }
      },
      ArrowLeft: handleBack,
      Backspace: handleBack,
      Enter: handleNext,
      "Shift+n": handleNext,
      "Shift+p": handleBack,
      "Shift+l": handleBack,
      r: () => {
        if (success) handleReset()
      },
    }
    for (let i = 1; i <= 6; i++) {
      map[String(i)] = () => {
        if (step === 0) {
          const s = services?.[i - 1]
          if (s) {
            setValue("serviceId", s.id)
            setSelectedDate("")
            setSelectedTime("")
          }
        } else if (step === 1) {
          let slotIdx = 0
          for (const t of slots) {
            if (isSlotOccupied(t)) continue
            slotIdx++
            if (slotIdx === i) {
              if (selectedDate) setSelectedTime(t)
              break
            }
          }
        }
      }
    }
    return map
  }, [
    step,
    services,
    slots,
    selectedDate,
    isSlotOccupied,
    errors.root,
    success,
    handleBack,
    handleNext,
    setValue,
    setError,
    handleReset,
  ])

  useHotkeys(hotkeyMap)

  const { t: ts } = useTranslation("shortcuts")
  const shortcutDefs = useMemo(
    () => [
      { key: "1-6", description: ts("booking.pick") },
      { key: "Enter", description: ts("booking.next") },
      { key: "Shift+n", description: ts("booking.next") },
      { key: "←", description: ts("booking.previous") },
      { key: "Esc", description: ts("booking.escape") },
      { key: "r", description: ts("booking.again") },
    ],
    [ts],
  )
  useRegisterShortcuts("customer", shortcutDefs)

  function onSubmit(data: AppointmentFormData) {
    createAppointment.mutate(
      { ...data, customerEmail: data.customerEmail || undefined },
      {
        onSuccess: () => setSuccess(true),
        onError: (err) => setError("root", { message: errorMessage(tc, err) }),
      },
    )
  }

  const fieldMessage = (message?: string) => (message ? fieldErrorMessage(tc, message) : undefined)

  if (success) return <SuccessView onReset={handleReset} />

  return (
    <div className="min-h-screen bg-background flex flex-col">
      <nav className="flex items-center justify-between px-4" style={{ height: "var(--nav-height)" }}>
        {step > 0 ? (
          <button onClick={handleBack} className="bg-transparent border-0 p-0 cursor-pointer rounded focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-focus-ring">
            <img src="/img/flecha.svg" alt={tc("back")} className="h-8" />
          </button>
        ) : (
          <div className="w-8" />
        )}
        <Link to="/" className="inline-flex" aria-label={businessConfig.branding.name}>
          <img src={businessConfig.branding.logoSrc} alt={businessConfig.branding.name} className="h-16" />
        </Link>
        <ThemeToggle />
        <a
          href={businessConfig.branding.instagramUrl}
          target="_blank"
          rel="noopener noreferrer"
          className="inline-flex"
        >
          <img src="/img/icono-instagram.svg" alt="Instagram" className="h-8" />
        </a>
      </nav>

      <main className="flex-1 px-4 pb-8 max-w-lg mx-auto w-full">
        <div className="text-center mb-6 page-enter">
          <h1 className="text-xl font-bold text-foreground">{t("title")}</h1>
          <p className="text-sm text-muted-foreground mt-1">{t("subtitle")}</p>
        </div>

        <div className="flex items-center justify-center gap-2 mb-8">
          {STEPS.map((step_, i) => (
            <div key={step_} className="flex items-center gap-1.5 sm:gap-2">
              <div
                aria-current={i === step ? "step" : undefined}
                className={`w-8 h-8 rounded-full flex items-center justify-center text-sm font-bold transition-colors shrink-0 ${
                  i === step
                    ? "bg-primary text-primary-foreground"
                    : i < step
                      ? "bg-accent text-primary"
                      : "bg-card text-muted-foreground"
                }`}
              >
                {i < step ? (
                  <svg className="w-4 h-4 animate-in zoom-in duration-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={3} d="M5 13l4 4L19 7" />
                  </svg>
                ) : (
                  i + 1
                )}
              </div>
              <span
                className={`text-xs sm:text-sm hidden sm:inline ${i === step ? "text-foreground font-medium" : "text-muted-foreground"}`}
              >
                {t(`steps.${step_}`)}
              </span>
              {i < STEPS.length - 1 && (
                <div className={`w-6 sm:w-8 h-px ${i < step ? "bg-primary" : "bg-border"}`} />
              )}
            </div>
          ))}
        </div>

        {errors.root && (
          <div
            role="alert"
            className="bg-destructive/10 text-destructive text-sm p-3 rounded-lg text-center mb-4 page-enter"
          >
            {errors.root.message}
          </div>
        )}

        <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
          {step === 0 && (
            <div className="page-enter space-y-4">
              <h2 className="text-lg font-semibold text-foreground">{t("service.title")}</h2>
              {loadingServices ? (
                <p className="text-muted-foreground">{t("service.loading")}</p>
              ) : servicesFailed ? (
                <div className="space-y-3">
                  <div role="alert" className="bg-destructive/10 text-destructive text-sm p-4 rounded-lg">
                    {t("service.loadError")}
                  </div>
                  <Button type="button" variant="secondary" onClick={() => refetchServices()}>
                    {tc("retry")}
                  </Button>
                </div>
              ) : !services?.length ? (
                <p className="text-muted-foreground">{t("service.empty")}</p>
              ) : (
                <div className="grid gap-3">
                  {services.map((s) => (
                    <button
                      key={s.id}
                      type="button"
                      onClick={() => {
                        setValue("serviceId", s.id)
                        setSelectedDate("")
                        setSelectedTime("")
                      }}
                      className={`text-left p-4 rounded-2xl border-2 transition-all duration-200 card-hover focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-focus-ring ${
                        selectedServiceId === s.id
                          ? "border-primary bg-accent"
                          : "border-border bg-card hover:border-primary/30"
                      }`}
                    >
                      <div className="flex justify-between items-center">
                        <span className="font-medium text-foreground">{s.name}</span>
                        <span className="text-primary font-bold">{formatCurrency(s.price)}</span>
                      </div>
                    </button>
                  ))}
                </div>
              )}
              {errors.serviceId && (
                <p className="text-sm text-destructive">{fieldMessage(errors.serviceId.message)}</p>
              )}
            </div>
          )}

          {step === 1 && (
            <div className="page-enter space-y-6">
              <h2 className="text-lg font-semibold text-foreground">{t("dateTime.title")}</h2>
              <div>
                <label htmlFor="booking-date" className="block text-sm font-medium mb-2 text-foreground">
                  {t("dateTime.pickDate")}
                </label>
                <Input
                  id="booking-date"
                  type="date"
                  value={selectedDate}
                  min={todayInBusinessTz()}
                  onChange={(e) => {
                    setSelectedDate(e.target.value)
                    setSelectedTime("")
                  }}
                  className="text-base"
                />
              </div>
              {selectedDate && (
                <div>
                  <label className="block text-sm font-medium mb-3 text-foreground">
                    {t("dateTime.availableTimes")}
                  </label>
                  <div className="flex flex-wrap justify-center gap-2">
                    {slots.map((time) => {
                      const occupied = isSlotOccupied(time)
                      const selected = selectedTime === time
                      return (
                        <button
                          key={selected ? `sel-${time}` : time}
                          type="button"
                          disabled={occupied}
                          onClick={() => setSelectedTime(time)}
                          className={`px-4 sm:px-5 min-h-11 rounded-[20px] border-2 text-sm font-medium transition-all duration-200 ${
                            occupied
                              ? "border-border bg-card text-muted-foreground line-through cursor-not-allowed opacity-50"
                              : selected
                                ? "border-primary bg-primary text-primary-foreground animate-in zoom-in duration-150"
                                : "border-primary bg-transparent text-foreground hover:bg-accent hover:scale-105 active:scale-95"
                          }`}
                        >
                          {time}
                        </button>
                      )
                    })}
                  </div>
                </div>
              )}
              {errors.startsAt && (
                <p className="text-sm text-destructive">{fieldMessage(errors.startsAt.message)}</p>
              )}
            </div>
          )}

          {step === 2 && (
            <div className="page-enter space-y-5">
              <h2 className="text-lg font-semibold text-foreground">{t("details.title")}</h2>
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <div className="space-y-2">
                  <label htmlFor="customerFirstName" className="text-sm font-medium text-foreground">{t("details.firstName")}</label>
                  <Input
                    id="customerFirstName"
                    {...register("customerFirstName")}
                    placeholder={t("details.firstNamePlaceholder")}
                    className="text-base"
                    aria-invalid={errors.customerFirstName ? true : undefined}
                    aria-describedby={errors.customerFirstName ? "customerFirstName-error" : undefined}
                  />
                  {errors.customerFirstName && (
                    <p id="customerFirstName-error" className="text-xs text-destructive">{fieldMessage(errors.customerFirstName.message)}</p>
                  )}
                </div>
                <div className="space-y-2">
                  <label htmlFor="customerLastName" className="text-sm font-medium text-foreground">{t("details.lastName")}</label>
                  <Input
                    id="customerLastName"
                    {...register("customerLastName")}
                    placeholder={t("details.lastNamePlaceholder")}
                    className="text-base"
                    aria-invalid={errors.customerLastName ? true : undefined}
                    aria-describedby={errors.customerLastName ? "customerLastName-error" : undefined}
                  />
                  {errors.customerLastName && (
                    <p id="customerLastName-error" className="text-xs text-destructive">{fieldMessage(errors.customerLastName.message)}</p>
                  )}
                </div>
              </div>
              <div className="space-y-2">
                <label htmlFor="customerPhone" className="text-sm font-medium text-foreground">{t("details.phone")}</label>
                <Input
                  id="customerPhone"
                  {...register("customerPhone")}
                  placeholder={t("details.phonePlaceholder")}
                  type="tel"
                  className="text-base"
                  aria-invalid={errors.customerPhone ? true : undefined}
                  aria-describedby={errors.customerPhone ? "customerPhone-error" : undefined}
                />
                {errors.customerPhone && (
                  <p id="customerPhone-error" className="text-xs text-destructive">{fieldMessage(errors.customerPhone.message)}</p>
                )}
              </div>
              <div className="space-y-2">
                <label htmlFor="customerEmail" className="text-sm font-medium text-foreground">{t("details.email")}</label>
                <Input
                  id="customerEmail"
                  {...register("customerEmail")}
                  placeholder={t("details.emailPlaceholder")}
                  type="email"
                  className="text-base"
                  aria-invalid={errors.customerEmail ? true : undefined}
                  aria-describedby={errors.customerEmail ? "customerEmail-error" : undefined}
                />
                {errors.customerEmail && (
                  <p id="customerEmail-error" className="text-xs text-destructive">{fieldMessage(errors.customerEmail.message)}</p>
                )}
              </div>

              {selectedService && (
                <div className="bg-card p-4 rounded-2xl space-y-1 text-sm border border-border">
                  <p className="text-foreground">
                    <span className="font-semibold text-primary">{t("summary.service")}</span>{" "}
                    {selectedService.name} — {formatCurrency(selectedService.price)}
                  </p>
                  {selectedDate && selectedTime && (
                    <p className="text-foreground">
                      <span className="font-semibold text-primary">{t("summary.appointment")}</span>{" "}
                      {t("summary.when", { date: formatDate(`${selectedDate}T00:00`), time: selectedTime })}
                    </p>
                  )}
                </div>
              )}
            </div>
          )}

          <div className="flex gap-3 pt-2">
            {step > 0 && step < 2 && (
              <Button type="button" variant="secondary" onClick={handleBack} className="flex-1">
                {tc("back")}
              </Button>
            )}
            {step < 2 ? (
              <Button type="button" variant="default" onClick={handleNext} className="flex-1">
                {t("actions.next")}
              </Button>
            ) : (
              <Button
                type="submit"
                variant="default"
                disabled={createAppointment.isPending}
                className="flex-1"
              >
                {createAppointment.isPending ? t("actions.submitting") : t("actions.submit")}
              </Button>
            )}
          </div>
        </form>
      </main>
    </div>
  )
}
