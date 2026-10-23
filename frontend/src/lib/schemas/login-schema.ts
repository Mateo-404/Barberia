import { z } from "zod"

export const loginSchema = z.object({
  email: z.string().email("Email inválido"),
  contrasenia: z.string().min(8, "Mínimo 8 caracteres"),
})

export type LoginFormData = z.infer<typeof loginSchema>
