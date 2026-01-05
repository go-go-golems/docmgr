import type { DiagnosticTaxonomy } from '../services/docmgrApi'

export function DiagnosticCard({ diag }: { diag: DiagnosticTaxonomy }) {
  const severity = (diag.Severity || 'info').toLowerCase()
  const badge =
    severity === 'warning'
      ? 'warning'
      : severity === 'error'
        ? 'danger'
        : severity === 'info'
          ? 'info'
          : 'secondary'

  const reason =
    typeof diag.Context === 'object' && diag.Context != null ? (diag.Context['Reason'] as unknown) : undefined

  return (
    <div className="alert alert-warning">
      <div className="fw-semibold mb-1">
        <span className={`badge text-bg-${badge} me-2`}>{diag.Severity ?? 'info'}</span>
        Parse diagnostics
      </div>
      <div className="small">
        {(diag.Stage ?? 'unknown') + (diag.Symptom ? ` • ${diag.Symptom}` : '')}
      </div>
      {diag.Path ? (
        <div className="small mt-1">
          <span className="text-muted">Path: </span>
          <span className="font-monospace">{diag.Path}</span>
        </div>
      ) : null}
      {typeof reason === 'string' && reason.trim() !== '' ? (
        <div className="small mt-1 text-muted">{reason}</div>
      ) : null}
    </div>
  )
}
