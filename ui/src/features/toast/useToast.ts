import { useCallback, useMemo } from 'react'

import { useAppDispatch } from '../../app/hooks'
import { pushToast, type ToastKind } from './toastSlice'

export function useToast(): {
  push: (args: { kind: ToastKind; message: string; timeoutMs?: number }) => void
  success: (message: string, opts?: { timeoutMs?: number }) => void
  error: (message: string, opts?: { timeoutMs?: number }) => void
  info: (message: string, opts?: { timeoutMs?: number }) => void
  warning: (message: string, opts?: { timeoutMs?: number }) => void
} {
  const dispatch = useAppDispatch()

  const push = useCallback(
    (args: { kind: ToastKind; message: string; timeoutMs?: number }) => {
      dispatch(pushToast(args))
    },
    [dispatch],
  )

  return useMemo(
    () => ({
      push,
      success: (message: string, opts?: { timeoutMs?: number }) => push({ kind: 'success', message, ...opts }),
      error: (message: string, opts?: { timeoutMs?: number }) => push({ kind: 'error', message, ...opts }),
      info: (message: string, opts?: { timeoutMs?: number }) => push({ kind: 'info', message, ...opts }),
      warning: (message: string, opts?: { timeoutMs?: number }) => push({ kind: 'warning', message, ...opts }),
    }),
    [push],
  )
}

