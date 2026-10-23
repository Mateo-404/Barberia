import { z } from "zod"

export const turnoSchema = z.object({
  idServicio: z.number({ message: "Seleccioná un servicio" }),
  fechaHora: z
    .string()
    .min(1, "Seleccioná fecha y horario")
    .refine(
      (val) => {
        const [datePart, timePart] = val.split("T")
        if (!datePart || !timePart) return false
        const [y, m, d] = datePart.split("-").map(Number)
        const [h, min] = timePart.split(":").map(Number)
        return new Date(y, m - 1, d, h, min) > new Date()
      },
      { message: "La fecha debe ser futura" },
    )
    .refine(
      (val) => {
        const [, timePart] = val.split("T")
        if (!timePart) return false
        const [h, m] = timePart.split(":").map(Number)
        return h >= 13 && h < 20 && (m === 0 || m === 30)
      },
      { message: "El horario debe ser entre 13:00 y 20:00, en intervalos de 30 minutos" },
    ),
  nombreCliente: z.string().min(1, "El nombre es obligatorio"),
  apellidoCliente: z.string().min(1, "El apellido es obligatorio"),
  telefonoCliente: z.string().min(1, "El teléfono es obligatorio"),
  emailCliente: z.string().email("Email inválido").optional().or(z.literal("")),
})

export type TurnoFormData = z.infer<typeof turnoSchema>
