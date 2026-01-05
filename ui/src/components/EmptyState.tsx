import type { ReactNode } from 'react'

export function EmptyState({
  title,
  children,
}: {
  title: string
  children?: ReactNode
}) {
  return (
    <div className="empty-state">
      <h4>{title}</h4>
      {children ? <div>{children}</div> : null}
    </div>
  )
}

