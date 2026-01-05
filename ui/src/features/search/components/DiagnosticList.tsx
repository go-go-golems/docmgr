import type { DiagnosticTaxonomy } from '../../../services/docmgrApi'

export function DiagnosticList({ diagnostics }: { diagnostics: DiagnosticTaxonomy[] }) {
  const max = 20
  const shown = diagnostics.slice(0, max)

  return (
    <div className="vstack gap-2">
      {shown.map((d, idx) => {
        const severity = (d.Severity || 'info').toLowerCase()
        const badge =
          severity === 'warning'
            ? 'warning'
            : severity === 'error'
              ? 'danger'
              : severity === 'info'
                ? 'info'
                : 'secondary'

        const reason = typeof d.Context === 'object' && d.Context != null ? (d.Context['Reason'] as unknown) : undefined

        return (
          <div key={`${d.Stage ?? 'stage'}:${d.Symptom ?? 'symptom'}:${idx}`} className="card">
            <div className="card-body py-2">
              <div className="d-flex justify-content-between align-items-start gap-2">
                <div>
                  <span className={`badge text-bg-${badge} me-2`}>{d.Severity ?? 'info'}</span>
                  <span className="fw-semibold">
                    {(d.Stage ?? 'unknown') + (d.Symptom ? ` • ${d.Symptom}` : '')}
                  </span>
                </div>
                {d.Tool ? <span className="text-muted small">{d.Tool}</span> : null}
              </div>
              {d.Path ? (
                <div className="small mt-1">
                  <span className="text-muted">Path: </span>
                  <span className="font-monospace">{d.Path}</span>
                </div>
              ) : null}
              {typeof reason === 'string' && reason.trim() !== '' ? (
                <div className="small mt-1 text-muted">{reason}</div>
              ) : null}
              <details className="mt-2">
                <summary className="small text-muted">Details</summary>
                <pre className="small mb-0">{JSON.stringify(d, null, 2)}</pre>
              </details>
            </div>
          </div>
        )
      })}
      {diagnostics.length > max ? (
        <div className="text-muted small">… {diagnostics.length - max} more diagnostics</div>
      ) : null}
    </div>
  )
}

