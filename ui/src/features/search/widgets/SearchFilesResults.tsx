import { EmptyState } from '../../../components/EmptyState'
import { LoadingSpinner } from '../../../components/LoadingSpinner'

import type { FileSuggestion } from '../../../services/docmgrApi'

export function SearchFilesResults({
  loading,
  hasSearched,
  total,
  results,
  onCopyPath,
}: {
  loading: boolean
  hasSearched: boolean
  total: number
  results: FileSuggestion[]
  onCopyPath: (path: string) => void
}) {
  if (loading) return <LoadingSpinner />

  if (!hasSearched) {
    return (
      <EmptyState title="Find related files">
        <p className="text-muted mb-0">Use query + ticket/topics context.</p>
      </EmptyState>
    )
  }

  return (
    <>
      <div className="mb-3">
        <strong>{total}</strong> files
      </div>

      {results.length > 0 ? (
        <div className="vstack gap-2">
          {results.map((s) => (
            <div key={`${s.file}:${s.source}:${s.reason}`} className="result-card">
              <div className="d-flex justify-content-between">
                <div className="flex-grow-1">
                  <div className="result-title font-monospace">{s.file}</div>
                  <div className="result-meta">
                    {s.source} • <span className="text-muted">{s.reason}</span>
                  </div>
                </div>
                <button className="btn btn-sm btn-outline-primary copy-btn ms-2" onClick={() => onCopyPath(s.file)}>
                  Copy
                </button>
              </div>
            </div>
          ))}
        </div>
      ) : (
        <EmptyState title="No files found">
          <p className="mb-0">Try adjusting your query or context filters.</p>
        </EmptyState>
      )}
    </>
  )
}

