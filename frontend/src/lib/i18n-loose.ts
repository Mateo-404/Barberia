/**
 * `t` typed for keys built at run time (a preset id from the server, an
 * error code). The catalog test and the fallbacks cover what the compiler
 * cannot, so use this only where the key really is dynamic.
 */
export type LooseT = (key: string, options?: Record<string, unknown>) => string

export const loose = (t: unknown): LooseT => t as LooseT
