import { z } from "zod"

// Messages are field error codes, translated where the form shows them
// (see lib/error-message.ts).
export const loginSchema = z.object({
  email: z.string().email("invalid_email").max(254, "too_long"),
  password: z.string().min(8, "too_short").max(128, "too_long"),
})

export type LoginFormData = z.infer<typeof loginSchema>
