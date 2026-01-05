import { useEffect, useRef } from 'react'

import { useAppDispatch, useAppSelector } from '../app/hooks'
import { removeToast } from '../features/toast/toastSlice'

export function ToastHost() {
  const dispatch = useAppDispatch()
  const toasts = useAppSelector((s) => s.toast.toasts)
  const timersRef = useRef<Map<string, number>>(new Map())

  useEffect(() => {
    for (const t of toasts) {
      if (timersRef.current.has(t.id)) continue
      const handle = window.setTimeout(() => {
        dispatch(removeToast({ id: t.id }))
      }, t.timeoutMs)
      timersRef.current.set(t.id, handle)
    }

    for (const [id, handle] of timersRef.current) {
      if (toasts.some((t) => t.id === id)) continue
      window.clearTimeout(handle)
      timersRef.current.delete(id)
    }
  }, [dispatch, toasts])

  useEffect(() => {
    const timers = timersRef.current
    return () => {
      for (const handle of timers.values()) window.clearTimeout(handle)
      timers.clear()
    }
  }, [])

  if (toasts.length === 0) return null

  function alertClass(kind: string): string {
    if (kind === 'success') return 'alert-success'
    if (kind === 'error') return 'alert-danger'
    if (kind === 'warning') return 'alert-warning'
    return 'alert-info'
  }

  return (
    <div className="position-fixed top-0 end-0 p-3" style={{ zIndex: 1080, maxWidth: 480 }}>
      <div className="d-flex flex-column gap-2">
        {toasts.map((t) => (
          <div key={t.id} className={`alert ${alertClass(t.kind)} py-2 px-3 mb-0 d-flex align-items-start gap-2`}>
            <div className="flex-grow-1">{t.message}</div>
            <button
              type="button"
              className="btn-close"
              aria-label="Close"
              onClick={() => dispatch(removeToast({ id: t.id }))}
            />
          </div>
        ))}
      </div>
    </div>
  )
}
