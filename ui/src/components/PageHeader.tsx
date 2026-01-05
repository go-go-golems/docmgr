import type { ReactNode } from 'react'

export function PageHeader({
  title,
  titleClassName,
  subtitle,
  subtitleClassName,
  actions,
  mb,
}: {
  title: ReactNode
  titleClassName?: string
  subtitle?: ReactNode
  subtitleClassName?: string
  actions?: ReactNode
  mb?: 0 | 1 | 2 | 3 | 4 | 5
}) {
  return (
    <div className={`d-flex justify-content-between align-items-center mb-${mb ?? 3} gap-2`}>
      <div className="flex-grow-1">
        <div className={`${titleClassName ?? 'h4'} mb-0`.trim()}>{title}</div>
        {subtitle ? <div className={`text-muted small ${subtitleClassName ?? ''}`.trim()}>{subtitle}</div> : null}
      </div>
      {actions ? <div className="d-flex gap-2 flex-wrap">{actions}</div> : null}
    </div>
  )
}
