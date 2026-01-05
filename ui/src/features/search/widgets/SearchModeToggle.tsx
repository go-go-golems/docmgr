import type { SearchMode } from '../searchSlice'

export function SearchModeToggle({
  mode,
  onModeChange,
  isMobile,
  showFilters,
  onToggleFilters,
  onOpenFilterDrawer,
}: {
  mode: SearchMode
  onModeChange: (mode: SearchMode) => void
  isMobile: boolean
  showFilters: boolean
  onToggleFilters: () => void
  onOpenFilterDrawer: () => void
}) {
  return (
    <div className="d-flex gap-2 mb-3">
      <button
        type="button"
        className={`btn btn-sm ${mode === 'docs' ? 'btn-primary' : 'btn-outline-primary'}`}
        onClick={() => onModeChange('docs')}
      >
        Docs
      </button>
      <button
        type="button"
        className={`btn btn-sm ${mode === 'reverse' ? 'btn-primary' : 'btn-outline-primary'}`}
        onClick={() => onModeChange('reverse')}
      >
        Reverse Lookup
      </button>
      <button
        type="button"
        className={`btn btn-sm ${mode === 'files' ? 'btn-primary' : 'btn-outline-primary'}`}
        onClick={() => onModeChange('files')}
      >
        Files
      </button>
      <div className="ms-auto">
        {isMobile ? (
          <button type="button" className="btn btn-sm btn-outline-secondary" onClick={onOpenFilterDrawer}>
            Filters
          </button>
        ) : (
          <button type="button" className="btn btn-sm btn-outline-secondary" onClick={onToggleFilters}>
            {showFilters ? 'Hide filters' : 'Show filters'}
          </button>
        )}
      </div>
    </div>
  )
}

