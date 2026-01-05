export type SearchChip = {
  key: string
  label: string
  onRemove: () => void
}

export function SearchActiveChips({ chips }: { chips: SearchChip[] }) {
  if (chips.length === 0) return null

  return (
    <div className="mb-3 d-flex flex-wrap gap-2 align-items-center">
      <div className="text-muted small">Active:</div>
      {chips.map((c) => (
        <button
          key={c.key}
          type="button"
          className="btn btn-sm btn-outline-secondary"
          onClick={c.onRemove}
        >
          {c.label} ×
        </button>
      ))}
    </div>
  )
}

