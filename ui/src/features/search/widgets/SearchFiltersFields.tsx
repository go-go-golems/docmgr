import { TopicMultiSelect } from '../components/TopicMultiSelect'
import type { SearchFilters } from '../searchSlice'

type Variant = 'desktop' | 'mobile'

export function SearchReverseFields({
  variant,
  filters,
  onFilterChange,
}: {
  variant: Variant
  filters: SearchFilters
  onFilterChange: <K extends keyof SearchFilters>(key: K, value: SearchFilters[K]) => void
}) {
  return (
    <div className={`row g-2 ${variant === 'desktop' ? 'mt-2' : ''}`}>
      <div className={variant === 'desktop' ? 'col-md-6' : 'col-12'}>
        <label className="form-label small mb-1">File</label>
        <input
          className="form-control form-control-sm"
          placeholder="backend/api/register.go or register.go"
          value={filters.file}
          onChange={(e) => onFilterChange('file', e.target.value)}
        />
      </div>
      <div className={variant === 'desktop' ? 'col-md-6' : 'col-12'}>
        <label className="form-label small mb-1">Dir</label>
        <input
          className="form-control form-control-sm"
          placeholder="backend/chat/ws/"
          value={filters.dir}
          onChange={(e) => onFilterChange('dir', e.target.value)}
        />
      </div>
    </div>
  )
}

export function SearchFiltersFields({
  variant,
  filters,
  onFilterChange,
}: {
  variant: Variant
  filters: SearchFilters
  onFilterChange: <K extends keyof SearchFilters>(key: K, value: SearchFilters[K]) => void
}) {
  const cols =
    variant === 'desktop'
      ? { ticket: 'col-md-3', topics: 'col-md-3', docType: 'col-md-2', status: 'col-md-2', orderBy: 'col-md-2' }
      : { ticket: 'col-12', topics: 'col-12', docType: 'col-12', status: 'col-12', orderBy: 'col-12' }

  return (
    <>
      <div className={cols.ticket}>
        <label className="form-label small mb-1">Ticket</label>
        <input
          className="form-control form-control-sm"
          placeholder="e.g. MEN-4242"
          value={filters.ticket}
          onChange={(e) => onFilterChange('ticket', e.target.value)}
        />
      </div>
      <div className={cols.topics}>
        <label className="form-label small mb-1">Topics</label>
        <TopicMultiSelect topics={filters.topics} onChange={(topics) => onFilterChange('topics', topics)} />
      </div>
      <div className={cols.docType}>
        <label className="form-label small mb-1">Type</label>
        <input
          className="form-control form-control-sm"
          placeholder="e.g. reference"
          value={filters.docType}
          onChange={(e) => onFilterChange('docType', e.target.value)}
        />
      </div>
      <div className={cols.status}>
        <label className="form-label small mb-1">Status</label>
        <select
          className="form-select form-select-sm"
          value={filters.status}
          onChange={(e) => onFilterChange('status', e.target.value)}
        >
          <option value="">All</option>
          <option value="active">active</option>
          <option value="review">review</option>
          <option value="complete">complete</option>
          <option value="draft">draft</option>
        </select>
      </div>
      <div className={cols.orderBy}>
        <label className="form-label small mb-1">Sort</label>
        <select
          className="form-select form-select-sm"
          value={filters.orderBy}
          onChange={(e) => onFilterChange('orderBy', e.target.value)}
        >
          <option value="rank">Relevance</option>
          <option value="path">Path</option>
          <option value="last_updated">Last updated</option>
        </select>
      </div>
    </>
  )
}
