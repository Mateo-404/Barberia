import { z } from "zod"
import { isFutureDateTime } from "@/lib/datetime"

// Messages are field error codes, translated where the form shows them
// (see lib/error-message.ts). Commercial rules (hours, intervals) live in the
// backend; the client only validates shape.
export const appointmentSchema = z.object({
  serviceId: z.number({ message: "required" }),
  startsAt: z
    .string()
    .min(1, "required")
    .refine((value) => isFutureDateTime(value), { message: "must_be_future" }),
  customerFirstName: z.string().trim().min(1, "required").max(120, "too_long"),
  customerLastName: z.string().trim().min(1, "required").max(120, "too_long"),
  customerPhone: z
    .string()
    .trim()
    .min(1, "required")
    .max(32, "too_long")
    .regex(/^[+\d][\d\s()-]{5,19}$/, "invalid_phone"),
  customerEmail: z.string().email("invalid_email").max(254, "too_long").optional().or(z.literal("")),
})

export type AppointmentFormData = z.infer<typeof appointmentSchema>
