import type { SearchFilters, SearchMode } from '../searchSlice'
import { SearchFiltersFields, SearchReverseFields } from './SearchFiltersFields'

export function SearchFiltersDesktop({
  mode,
  filters,
  onFilterChange,
  onClear,
}: {
  mode: SearchMode
  filters: SearchFilters
  onFilterChange: <K extends keyof SearchFilters>(key: K, value: SearchFilters[K]) => void
  onClear: () => void
}) {
  return (
    <div className="filter-row mb-3">
      <div className="row g-2 align-items-end">
        <SearchFiltersFields variant="desktop" filters={filters} onFilterChange={onFilterChange} />
      </div>

      {mode === 'reverse' ? (
        <SearchReverseFields variant="desktop" filters={filters} onFilterChange={onFilterChange} />
      ) : null}

      <div className="d-flex flex-wrap gap-3 mt-3 align-items-center">
        <div className="form-check">
          <input
            className="form-check-input"
            type="checkbox"
            checked={filters.includeArchived}
            onChange={(e) => onFilterChange('includeArchived', e.target.checked)}
            id="includeArchived"
          />
          <label className="form-check-label" htmlFor="includeArchived">
            Include archived
          </label>
        </div>
        <div className="form-check">
          <input
            className="form-check-input"
            type="checkbox"
            checked={filters.includeScripts}
            onChange={(e) => onFilterChange('includeScripts', e.target.checked)}
            id="includeScripts"
          />
          <label className="form-check-label" htmlFor="includeScripts">
            Include scripts
          </label>
        </div>
        <div className="form-check">
          <input
            className="form-check-input"
            type="checkbox"
            checked={filters.includeControlDocs}
            onChange={(e) => onFilterChange('includeControlDocs', e.target.checked)}
            id="includeControlDocs"
          />
          <label className="form-check-label" htmlFor="includeControlDocs">
            Control docs
          </label>
        </div>
        <div className="ms-auto">
          <button type="button" className="btn btn-sm btn-outline-secondary" onClick={onClear}>
            Clear
          </button>
        </div>
      </div>
    </div>
  )
}
