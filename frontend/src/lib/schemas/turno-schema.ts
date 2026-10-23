import { z } from "zod"
import { isFutureFechaHora } from "@/lib/datetime"

/**
 * Validación de FORMA y datos del cliente. Las reglas comerciales (horarios de
 * atención, intervalos) ya no viven en el cliente: la grilla sale de
 * `lib/business-config.ts` y la autoridad final es el backend
 * (ver docs/frontend/04-contrato-configuracion-negocio.md).
 */
export const turnoSchema = z.object({
  idServicio: z.number({ message: "Seleccioná un servicio" }),
  fechaHora: z
    .string()
    .min(1, "Seleccioná fecha y horario")
    .refine((val) => isFutureFechaHora(val), {
      message: "La fecha debe ser futura",
    }),
  nombreCliente: z.string().trim().min(1, "El nombre es obligatorio").max(120, "Máximo 120 caracteres"),
  apellidoCliente: z.string().trim().min(1, "El apellido es obligatorio").max(120, "Máximo 120 caracteres"),
  telefonoCliente: z
    .string()
    .trim()
    .min(1, "El teléfono es obligatorio")
    .max(32, "Teléfono demasiado largo")
    .regex(/^[+\d][\d\s()-]{5,19}$/, "Ingresá un teléfono válido (sólo números, espacios o +)"),
  emailCliente: z
    .string()
    .email("Email inválido")
    .max(254, "Email demasiado largo")
    .optional()
    .or(z.literal("")),
})

export type TurnoFormData = z.infer<typeof turnoSchema>
