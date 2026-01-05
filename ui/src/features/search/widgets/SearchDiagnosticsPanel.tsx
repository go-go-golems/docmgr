import { timeAgo } from '../../../lib/time'
import type { DiagnosticTaxonomy, WorkspaceStatus } from '../../../services/docmgrApi'
import { DiagnosticList } from '../components/DiagnosticList'

export function SearchDiagnosticsPanel({
  hasSearched,
  docsTotal,
  diagnostics,
  showDiagnostics,
  onToggleDiagnostics,
  wsStatus,
}: {
  hasSearched: boolean
  docsTotal: number
  diagnostics: DiagnosticTaxonomy[]
  showDiagnostics: boolean
  onToggleDiagnostics: () => void
  wsStatus?: WorkspaceStatus
}) {
  return (
    <>
      <div className="d-flex align-items-center mb-3">
        <div>
          {hasSearched ? (
            <>
              <strong>{docsTotal}</strong> results
            </>
          ) : (
            <span className="text-muted">No search performed yet</span>
          )}
        </div>
        {diagnostics.length > 0 ? (
          <button type="button" className="btn btn-sm btn-outline-warning ms-3" onClick={onToggleDiagnostics}>
            {diagnostics.length} diagnostics {showDiagnostics ? '▲' : '▼'}
          </button>
        ) : null}
        <div className="ms-auto">
          {wsStatus ? (
            <span className="text-muted small">
              Indexed {timeAgo(wsStatus.indexedAt)} • {wsStatus.docsIndexed} docs
            </span>
          ) : null}
        </div>
      </div>

      {showDiagnostics && diagnostics.length > 0 ? (
        <div className="alert alert-warning">
          <div className="fw-semibold mb-2">Diagnostics</div>
          <DiagnosticList diagnostics={diagnostics} />
        </div>
      ) : null}
    </>
  )
}

