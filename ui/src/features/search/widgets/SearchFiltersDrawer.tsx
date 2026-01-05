import type { SearchFilters, SearchMode } from '../searchSlice'
import { SearchFiltersFields, SearchReverseFields } from './SearchFiltersFields'

export function SearchFiltersDrawer({
  open,
  mode,
  filters,
  onFilterChange,
  onClose,
  onClear,
}: {
  open: boolean
  mode: SearchMode
  filters: SearchFilters
  onFilterChange: <K extends keyof SearchFilters>(key: K, value: SearchFilters[K]) => void
  onClose: () => void
  onClear: () => void
}) {
  if (!open) return null

  return (
    <>
      <div className="modal-backdrop show" />
      <div className="modal show d-block" tabIndex={-1} role="dialog" aria-modal="true">
        <div className="modal-dialog modal-fullscreen-sm-down modal-dialog-scrollable" role="document">
          <div className="modal-content">
            <div className="modal-header">
              <h5 className="modal-title">Filters</h5>
              <button type="button" className="btn-close" onClick={onClose} />
            </div>
            <div className="modal-body">
              <div className="filter-row mb-0 border-0 p-0">
                <div className="row g-2 align-items-end">
                  <SearchFiltersFields variant="mobile" filters={filters} onFilterChange={onFilterChange} />
                </div>

                {mode === 'reverse' ? (
                  <SearchReverseFields variant="mobile" filters={filters} onFilterChange={onFilterChange} />
                ) : null}

                <div className="d-flex flex-column gap-2 mt-3">
                  <div className="form-check">
                    <input
                      className="form-check-input"
                      type="checkbox"
                      checked={filters.includeArchived}
                      onChange={(e) => onFilterChange('includeArchived', e.target.checked)}
                      id="includeArchivedMobile"
                    />
                    <label className="form-check-label" htmlFor="includeArchivedMobile">
                      Include archived
                    </label>
                  </div>
                  <div className="form-check">
                    <input
                      className="form-check-input"
                      type="checkbox"
                      checked={filters.includeScripts}
                      onChange={(e) => onFilterChange('includeScripts', e.target.checked)}
                      id="includeScriptsMobile"
                    />
                    <label className="form-check-label" htmlFor="includeScriptsMobile">
                      Include scripts
                    </label>
                  </div>
                  <div className="form-check">
                    <input
                      className="form-check-input"
                      type="checkbox"
                      checked={filters.includeControlDocs}
                      onChange={(e) => onFilterChange('includeControlDocs', e.target.checked)}
                      id="includeControlDocsMobile"
                    />
                    <label className="form-check-label" htmlFor="includeControlDocsMobile">
                      Control docs
                    </label>
                  </div>
                </div>
              </div>
            </div>
            <div className="modal-footer">
              <button type="button" className="btn btn-outline-secondary" onClick={onClear}>
                Clear
              </button>
              <button type="button" className="btn btn-primary" onClick={onClose}>
                Done
              </button>
            </div>
          </div>
        </div>
      </div>
    </>
  )
}
