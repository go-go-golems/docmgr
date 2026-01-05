import type { ReactNode } from 'react'

export function PathHeader({
  path,
  label = 'Path',
  actions,
}: {
  path: string
  label?: string
  actions?: ReactNode
}) {
  return (
    <div className="mb-2">
      <span className="text-muted small">{label}</span>
      <div className="dm-path-pill">{path}</div>
      {actions ? <div className="mt-2 d-flex gap-2 flex-wrap">{actions}</div> : null}
    </div>
  )
}
