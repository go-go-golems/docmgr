export function StatusBadge({ status, className }: { status: string; className?: string }) {
  const variant =
    status === 'active'
      ? 'primary'
      : status === 'review'
        ? 'warning'
        : status === 'complete'
          ? 'success'
          : status === 'draft'
            ? 'secondary'
            : 'secondary'
  return (
    <span className={`badge text-bg-${variant} ${className ?? ''}`.trim()} style={{ fontWeight: 600 }}>
      {status || 'unknown'}
    </span>
  )
}
