import { useState, useMemo } from "react"
import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { useServicios } from "@/api/servicios"
import { useFechasOcupadas, useCrearTurno } from "@/api/turnos"
import { turnoSchema, type TurnoFormData } from "@/lib/schemas/turno-schema"
import { ApiError } from "@/lib/api-client"

const STEPS = ["Servicio", "Fecha y hora", "Tus datos"]

const TIME_SLOTS = [
  "13:00", "13:30", "14:00", "14:30",
  "15:00", "15:30", "16:00", "16:30",
  "17:00", "17:30", "18:00", "18:30",
  "19:00", "19:30",
]

function todayStr() {
  const d = new Date()
  const mm = String(d.getMonth() + 1).padStart(2, "0")
  const dd = String(d.getDate()).padStart(2, "0")
  return `${d.getFullYear()}-${mm}-${dd}`
}

export default function Reserva() {
  const [step, setStep] = useState(0)
  const [success, setSuccess] = useState(false)
  const [selectedDate, setSelectedDate] = useState("")
  const [selectedTime, setSelectedTime] = useState("")

  const { data: servicios, isLoading: loadingServicios } = useServicios()
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

  function isSlotOccupied(time: string) {
    if (!selectedDate || !fechasOcupadas) return false
    return fechasOcupadas.includes(`${selectedDate}T${time}:00`)
  }

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

  if (success) {
    return (
      <main className="min-h-screen flex items-center justify-center p-4">
        <div className="w-full max-w-md text-center space-y-4">
          <div className="text-6xl">✅</div>
          <h1 className="text-2xl font-bold">Turno reservado</h1>
          <p className="text-muted-foreground">Te esperamos en la barbería.</p>
          <button
            type="button"
            onClick={() => {
              setSuccess(false)
              setStep(0)
              setSelectedDate("")
              setSelectedTime("")
              reset()
            }}
            className="underline text-sm"
          >
            Reservar otro turno
          </button>
        </div>
      </main>
    )
  }

  return (
    <main className="min-h-screen flex items-start justify-center p-4 pt-12">
      <div className="w-full max-w-lg space-y-8">
        <div className="text-center space-y-1">
          <h1 className="text-2xl font-bold">Reservá tu turno</h1>
          <p className="text-sm text-muted-foreground">Completá los pasos para agendar</p>
        </div>

        <div className="flex items-center justify-center gap-2">
          {STEPS.map((label, i) => (
            <div key={label} className="flex items-center gap-2">
              <div
                className={`w-8 h-8 rounded-full flex items-center justify-center text-sm font-medium ${
                  i === step
                    ? "bg-primary text-primary-foreground"
                    : i < step
                      ? "bg-primary/20 text-primary"
                      : "bg-muted text-muted-foreground"
                }`}
              >
                {i < step ? "✓" : i + 1}
              </div>
              <span
                className={`text-sm hidden sm:inline ${i === step ? "font-medium" : "text-muted-foreground"}`}
              >
                {label}
              </span>
              {i < STEPS.length - 1 && <div className="w-8 h-px bg-border" />}
            </div>
          ))}
        </div>

        {errors.root && (
          <div className="bg-destructive/10 text-destructive text-sm p-3 rounded-lg">
            {errors.root.message}
          </div>
        )}

        <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
          {step === 0 && (
            <div className="space-y-3">
              <h2 className="text-lg font-semibold">Elegí un servicio</h2>
              {loadingServicios ? (
                <p className="text-muted-foreground">Cargando servicios...</p>
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
                      className={`text-left p-4 rounded-xl border-2 transition-colors ${
                        selectedServicioId === s.id
                          ? "border-primary bg-primary/5"
                          : "border-border hover:border-primary/50"
                      }`}
                    >
                      <div className="flex justify-between items-center">
                        <span className="font-medium">{s.tipo}</span>
                        <span className="text-muted-foreground">${s.precio}</span>
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
            <div className="space-y-4">
              <h2 className="text-lg font-semibold">Elegí fecha y horario</h2>
              <div>
                <label className="block text-sm font-medium mb-1">Fecha</label>
                <input
                  type="date"
                  value={selectedDate}
                  min={todayStr()}
                  onChange={(e) => {
                    setSelectedDate(e.target.value)
                    setSelectedTime("")
                  }}
                  className="w-full px-3 py-2 rounded-lg border border-input bg-background"
                />
              </div>
              {selectedDate && (
                <div>
                  <label className="block text-sm font-medium mb-2">Horario</label>
                  <div className="grid grid-cols-4 gap-2">
                    {TIME_SLOTS.map((time) => {
                      const occupied = isSlotOccupied(time)
                      const selected = selectedTime === time
                      return (
                        <button
                          key={time}
                          type="button"
                          disabled={occupied}
                          onClick={() => {
                            setSelectedTime(time)
                          }}
                          className={`py-2 px-3 rounded-lg text-sm font-medium border transition-colors ${
                            occupied
                              ? "bg-muted text-muted-foreground/50 line-through cursor-not-allowed"
                              : selected
                                ? "border-primary bg-primary/5 text-primary"
                                : "border-border hover:border-primary/50"
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
            <div className="space-y-4">
              <h2 className="text-lg font-semibold">Tus datos</h2>
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-sm font-medium mb-1">Nombre</label>
                  <input
                    {...register("nombreCliente")}
                    className="w-full px-3 py-2 rounded-lg border border-input bg-background"
                  />
                  {errors.nombreCliente && (
                    <p className="text-xs text-destructive mt-1">{errors.nombreCliente.message}</p>
                  )}
                </div>
                <div>
                  <label className="block text-sm font-medium mb-1">Apellido</label>
                  <input
                    {...register("apellidoCliente")}
                    className="w-full px-3 py-2 rounded-lg border border-input bg-background"
                  />
                  {errors.apellidoCliente && (
                    <p className="text-xs text-destructive mt-1">{errors.apellidoCliente.message}</p>
                  )}
                </div>
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">Teléfono</label>
                <input
                  {...register("telefonoCliente")}
                  className="w-full px-3 py-2 rounded-lg border border-input bg-background"
                />
                {errors.telefonoCliente && (
                  <p className="text-xs text-destructive mt-1">{errors.telefonoCliente.message}</p>
                )}
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">Email (opcional)</label>
                <input
                  type="email"
                  {...register("emailCliente")}
                  className="w-full px-3 py-2 rounded-lg border border-input bg-background"
                />
                {errors.emailCliente && (
                  <p className="text-xs text-destructive mt-1">{errors.emailCliente.message}</p>
                )}
              </div>

              {selectedServicio && (
                <div className="bg-muted p-4 rounded-xl space-y-1 text-sm">
                  <p>
                    <span className="font-medium">Servicio:</span> {selectedServicio.tipo} — $
                    {selectedServicio.precio}
                  </p>
                  {selectedDate && selectedTime && (
                    <p>
                      <span className="font-medium">Turno:</span> {selectedDate} a las{" "}
                      {selectedTime}
                    </p>
                  )}
                </div>
              )}
            </div>
          )}

          <div className="flex gap-3">
            {step > 0 && (
              <button
                type="button"
                onClick={handleBack}
                className="flex-1 py-2.5 rounded-lg border border-input hover:bg-muted transition-colors"
              >
                Anterior
              </button>
            )}
            {step < 2 ? (
              <button
                type="button"
                onClick={handleNext}
                className="flex-1 py-2.5 rounded-lg bg-primary text-primary-foreground hover:opacity-90 transition-opacity font-medium"
              >
                Siguiente
              </button>
            ) : (
              <button
                type="submit"
                disabled={crearTurno.isPending}
                className="flex-1 py-2.5 rounded-lg bg-primary text-primary-foreground hover:opacity-90 transition-opacity font-medium disabled:opacity-50"
              >
                {crearTurno.isPending ? "Reservando..." : "Reservar turno"}
              </button>
            )}
          </div>
        </form>
      </div>
    </main>
  )
}
