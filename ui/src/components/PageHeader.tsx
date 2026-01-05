import type { ReactNode } from 'react'

export function PageHeader({
  title,
  subtitle,
  subtitleClassName,
  actions,
}: {
  title: ReactNode
  subtitle?: ReactNode
  subtitleClassName?: string
  actions?: ReactNode
}) {
  return (
    <div className="d-flex justify-content-between align-items-center mb-3 gap-2">
      <div className="flex-grow-1">
        <div className="h4 mb-0">{title}</div>
        {subtitle ? <div className={`text-muted small ${subtitleClassName ?? ''}`.trim()}>{subtitle}</div> : null}
      </div>
      {actions ? <div className="d-flex gap-2 flex-wrap">{actions}</div> : null}
    </div>
  )
}

