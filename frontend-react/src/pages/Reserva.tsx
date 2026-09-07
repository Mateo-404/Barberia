import { useState, useMemo, useCallback } from "react"
import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { useServicios } from "@/api/servicios"
import { useFechasOcupadas, useCrearTurno } from "@/api/turnos"
import { turnoSchema, type TurnoFormData } from "@/lib/schemas/turno-schema"
import { ApiError } from "@/lib/api-client"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { todayInputValue } from "@/lib/format"

const STEPS = ["Servicio", "Fecha y hora", "Tus datos"]

const TIME_SLOTS = [
  "13:00", "13:30", "14:00", "14:30",
  "15:00", "15:30", "16:00", "16:30",
  "17:00", "17:30", "18:00", "18:30",
  "19:00", "19:30",
]

const SuccessView = ({
  onReset,
}: {
  onReset: () => void
}) => (
  <main className="min-h-screen flex flex-col items-center justify-center p-6 bg-background">
    <div className="w-full max-w-md text-center space-y-6 page-enter">
      <div className="w-20 h-20 rounded-full bg-primary flex items-center justify-center mx-auto" style={{ animation: "pop 0.7s ease-out" }}>
        <svg className="w-10 h-10 text-primary-foreground" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={3} d="M5 13l4 4L19 7" />
        </svg>
      </div>
      <h1 className="text-2xl font-bold text-foreground">Turno reservado</h1>
      <p className="text-muted-foreground">Te esperamos en la barbería.</p>
      <Button variant="secondary" onClick={onReset}>
        Reservar otro turno
      </Button>
    </div>
  </main>
)

export default function Reserva() {
  const [step, setStep] = useState(0)
  const [success, setSuccess] = useState(false)
  const [selectedDate, setSelectedDate] = useState("")
  const [selectedTime, setSelectedTime] = useState("")

  const { data: servicios, isLoading: loadingServicios, isError: errorServicios } = useServicios()
  const { data: fechasOcupadas } = useFechasOcupadas()
  const crearTurno = useCrearTurno()

  const {
    register,
    handleSubmit,
    formState: { errors },
    setValue,
    trigger,
    setError,
    watch,
    reset,
  } = useForm<TurnoFormData>({
    resolver: zodResolver(turnoSchema),
    defaultValues: {
      idServicio: undefined as unknown as number,
      fechaHora: "",
      nombreCliente: "",
      apellidoCliente: "",
      telefonoCliente: "",
      emailCliente: "",
    },
    mode: "onChange",
  })

  const selectedServicioId = watch("idServicio")

  const selectedServicio = useMemo(() => {
    if (!servicios || !selectedServicioId) return null
    return servicios.find((s) => s.id === selectedServicioId) ?? null
  }, [servicios, selectedServicioId])

  const isSlotOccupied = useCallback(
    (time: string) => {
      if (!selectedDate || !fechasOcupadas) return false
      return fechasOcupadas.includes(`${selectedDate}T${time}`)
    },
    [selectedDate, fechasOcupadas],
  )

  async function handleNext() {
    if (step === 0) {
      const valid = await trigger("idServicio")
      if (valid) setStep(1)
    } else if (step === 1) {
      if (!selectedDate || !selectedTime) {
        setError("fechaHora", { message: "Seleccioná fecha y horario" })
        return
      }
      const isoStr = `${selectedDate}T${selectedTime}:00`
      setValue("fechaHora", isoStr, { shouldValidate: true })
      const valid = await trigger("fechaHora")
      if (valid) setStep(2)
    }
  }

  function handleBack() {
    setStep((s) => Math.max(s - 1, 0))
  }

  function handleReset() {
    setSuccess(false)
    setStep(0)
    setSelectedDate("")
    setSelectedTime("")
    reset()
  }

  async function onSubmit(data: TurnoFormData) {
    crearTurno.mutate(
      {
        fechaHora: data.fechaHora,
        idServicio: data.idServicio,
        telefonoCliente: data.telefonoCliente,
        nombreCliente: data.nombreCliente,
        apellidoCliente: data.apellidoCliente,
        emailCliente: data.emailCliente || undefined,
      },
      {
        onSuccess: () => setSuccess(true),
        onError: (err: Error) => {
          const detail = err instanceof ApiError ? err.detail : err.message
          setError("root", { message: detail })
        },
      },
    )
  }

  if (success) return <SuccessView onReset={handleReset} />

  return (
    <div className="min-h-screen bg-background flex flex-col">
      <nav className="flex items-center justify-between px-4" style={{ height: "var(--nav-height)" }}>
        {step > 0 ? (
          <button onClick={handleBack} className="bg-transparent border-0 p-0 cursor-pointer rounded focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-focus-ring">
            <img src="/img/flecha.svg" alt="Volver" className="h-8" />
          </button>
        ) : (
          <div className="w-8" />
        )}
        <a href="/" className="inline-flex">
          <img src="/logo.svg" alt="TH Barber Club" className="h-12" />
        </a>
        <a href="https://instagram.com" target="_blank" rel="noopener noreferrer" className="inline-flex">
          <img src="/img/icono-instagram.svg" alt="Instagram" className="h-8" />
        </a>
      </nav>

      <main className="flex-1 px-4 pb-8 max-w-lg mx-auto w-full">
        <div className="text-center mb-6 page-enter">
          <h1 className="text-xl font-bold text-foreground">¡Reservá tu turno!</h1>
          <p className="text-sm text-muted-foreground mt-1">
            Completá los pasos para agendar tu cita
          </p>
        </div>

        <div className="flex items-center justify-center gap-2 mb-8">
          {STEPS.map((label, i) => (
            <div key={label} className="flex items-center gap-1.5 sm:gap-2">
              <div
                className={`w-8 h-8 rounded-full flex items-center justify-center text-sm font-bold transition-colors shrink-0 ${
                  i === step
                    ? "bg-primary text-primary-foreground"
                    : i < step
                      ? "bg-accent text-primary"
                      : "bg-card text-muted-foreground"
                }`}
              >
                {i < step ? (
                  <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={3} d="M5 13l4 4L19 7" />
                  </svg>
                ) : (
                  i + 1
                )}
              </div>
              <span
                className={`text-xs sm:text-sm hidden sm:inline ${i === step ? "text-foreground font-medium" : "text-muted-foreground"}`}
              >
                {label}
              </span>
              {i < STEPS.length - 1 && (
                <div className={`w-6 sm:w-8 h-px ${i < step ? "bg-primary" : "bg-border"}`} />
              )}
            </div>
          ))}
        </div>

        {errors.root && (
          <div className="bg-destructive/10 text-destructive text-sm p-3 rounded-lg text-center mb-4 page-enter">
            {errors.root.message}
          </div>
        )}

        <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
          {step === 0 && (
            <div className="page-enter space-y-4">
              <h2 className="text-lg font-semibold text-foreground">Elegí un servicio</h2>
              {loadingServicios ? (
                <p className="text-muted-foreground">Cargando servicios...</p>
              ) : errorServicios ? (
                <div className="bg-destructive/10 text-destructive text-sm p-4 rounded-lg">
                  No se pudieron cargar los servicios. Revisá tu conexión e intentá de nuevo.
                </div>
              ) : !servicios?.length ? (
                <p className="text-muted-foreground">No hay servicios disponibles.</p>
              ) : (
                <div className="grid gap-3">
                  {servicios.map((s) => (
                    <button
                      key={s.id}
                      type="button"
                      onClick={() => {
                        setValue("idServicio", s.id)
                        setSelectedDate("")
                        setSelectedTime("")
                      }}
                      className={`text-left p-4 rounded-2xl border-2 transition-all duration-200 card-hover focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-focus-ring ${
                        selectedServicioId === s.id
                          ? "border-primary bg-accent"
                          : "border-border bg-card hover:border-primary/30"
                      }`}
                    >
                      <div className="flex justify-between items-center">
                        <span className="font-medium text-foreground">{s.tipo}</span>
                        <span className="text-primary font-bold">${s.precio}</span>
                      </div>
                    </button>
                  ))}
                </div>
              )}
              {errors.idServicio && (
                <p className="text-sm text-destructive">{errors.idServicio.message}</p>
              )}
            </div>
          )}

          {step === 1 && (
            <div className="page-enter space-y-6">
              <h2 className="text-lg font-semibold text-foreground">Elegí fecha y horario</h2>
              <div>
                <label className="block text-sm font-medium mb-2 text-foreground">
                  Seleccionar fecha
                </label>
                <Input
                  type="date"
                  value={selectedDate}
                  min={todayInputValue()}
                  onChange={(e) => {
                    setSelectedDate(e.target.value)
                    setSelectedTime("")
                  }}
                />
              </div>
              {selectedDate && (
                <div>
                  <label className="block text-sm font-medium mb-3 text-foreground">
                    Horarios disponibles
                  </label>
                  <div className="flex flex-wrap justify-center gap-2">
                    {TIME_SLOTS.map((time) => {
                      const occupied = isSlotOccupied(time)
                      const selected = selectedTime === time
                      return (
                        <button
                          key={time}
                          type="button"
                          disabled={occupied}
                          onClick={() => setSelectedTime(time)}
                          className={`px-4 sm:px-5 py-2.5 sm:py-3 rounded-[20px] border-2 text-sm font-medium transition-all duration-200 ${
                            occupied
                              ? "border-border bg-card text-muted-foreground line-through cursor-not-allowed opacity-50"
                              : selected
                                ? "border-primary bg-primary text-primary-foreground"
                                : "border-primary bg-transparent text-foreground hover:bg-accent"
                          }`}
                        >
                          {time}
                        </button>
                      )
                    })}
                  </div>
                </div>
              )}
              {errors.fechaHora && (
                <p className="text-sm text-destructive">{errors.fechaHora.message}</p>
              )}
            </div>
          )}

          {step === 2 && (
            <div className="page-enter space-y-5">
              <h2 className="text-lg font-semibold text-foreground">Tus datos</h2>
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <div className="space-y-2">
                  <label className="text-sm font-medium text-foreground">Nombre</label>
                  <Input {...register("nombreCliente")} placeholder="Juan" />
                  {errors.nombreCliente && (
                    <p className="text-xs text-destructive">{errors.nombreCliente.message}</p>
                  )}
                </div>
                <div className="space-y-2">
                  <label className="text-sm font-medium text-foreground">Apellido</label>
                  <Input {...register("apellidoCliente")} placeholder="Pérez" />
                  {errors.apellidoCliente && (
                    <p className="text-xs text-destructive">{errors.apellidoCliente.message}</p>
                  )}
                </div>
              </div>
              <div className="space-y-2">
                <label className="text-sm font-medium text-foreground">Teléfono</label>
                <Input {...register("telefonoCliente")} placeholder="1155551234" type="tel" />
                {errors.telefonoCliente && (
                  <p className="text-xs text-destructive">{errors.telefonoCliente.message}</p>
                )}
              </div>
              <div className="space-y-2">
                <label className="text-sm font-medium text-foreground">Email (opcional)</label>
                <Input {...register("emailCliente")} placeholder="juan@email.com" type="email" />
                {errors.emailCliente && (
                  <p className="text-xs text-destructive">{errors.emailCliente.message}</p>
                )}
              </div>

              {selectedServicio && (
                <div className="bg-card p-4 rounded-2xl space-y-1 text-sm border border-border">
                  <p className="text-foreground">
                    <span className="font-semibold text-primary">Servicio:</span>{" "}
                    {selectedServicio.tipo} — ${selectedServicio.precio}
                  </p>
                  {selectedDate && selectedTime && (
                    <p className="text-foreground">
                      <span className="font-semibold text-primary">Turno:</span>{" "}
                      {selectedDate} a las {selectedTime}
                    </p>
                  )}
                </div>
              )}
            </div>
          )}

          <div className="flex gap-3 pt-2">
            {step > 0 && step < 2 && (
              <Button type="button" variant="secondary" onClick={handleBack} className="flex-1">
                Atrás
              </Button>
            )}
            {step < 2 ? (
              <Button type="button" variant="default" onClick={handleNext} className="flex-1">
                Siguiente
              </Button>
            ) : (
              <Button
                type="submit"
                variant="default"
                disabled={crearTurno.isPending}
                className="flex-1"
              >
                {crearTurno.isPending ? "Reservando..." : "Reservar turno"}
              </Button>
            )}
          </div>
        </form>
      </main>
    </div>
  )
}
