export type APIErrorPayload = {
  error?: {
    code?: string
    message?: string
    details?: unknown
  }
}

export type APIErrorDetails = {
  code?: string
  message: string
  details?: unknown
}

export function apiErrorFromUnknown(err: unknown): APIErrorDetails {
  const maybe = err as { data?: unknown; status?: number } | undefined
  const data = maybe?.data as APIErrorPayload | undefined
  const code = data?.error?.code
  const message =
    data?.error?.message ??
    (typeof err === 'string' ? err : err instanceof Error ? err.message : String(err))
  const details = data?.error?.details
  return { code, message, details }
}

export function apiErrorMessage(err: unknown): string {
  return apiErrorFromUnknown(err).message
}

