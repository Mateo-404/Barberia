import { z } from "zod"

export const loginSchema = z.object({
  email: z.string().email("Email inválido").max(254, "Email demasiado largo"),
  contrasenia: z.string().min(8, "Mínimo 8 caracteres").max(128, "Máximo 128 caracteres"),
})

export type LoginFormData = z.infer<typeof loginSchema>
