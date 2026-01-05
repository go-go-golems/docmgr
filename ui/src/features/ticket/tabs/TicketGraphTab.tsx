import { ApiErrorAlert } from '../../../components/ApiErrorAlert'
import { LoadingSpinner } from '../../../components/LoadingSpinner'
import { MermaidDiagram } from '../../../components/MermaidDiagram'
import type { TicketGraphResponse } from '../../../services/docmgrApi'

export function TicketGraphTab({
  graphData,
  graphError,
  graphLoading,
}: {
  graphData?: TicketGraphResponse
  graphError?: unknown
  graphLoading: boolean
}) {
  return (
    <div className="card">
      <div className="card-header fw-semibold d-flex justify-content-between align-items-center">
        <span>Graph</span>
        <span className="text-muted small">
          {graphData ? `${graphData.stats.nodes} nodes • ${graphData.stats.edges} edges` : ''}
        </span>
      </div>
      <div className="card-body">
        {graphError ? <ApiErrorAlert title="Failed to load graph" error={graphError} /> : null}
        {graphLoading ? <LoadingSpinner /> : null}
        {graphData ? (
          <div>
            <div className="overflow-auto border rounded p-2" style={{ maxHeight: 650 }}>
              <MermaidDiagram code={graphData.mermaid} />
            </div>
            <details className="mt-3">
              <summary className="text-muted small">Mermaid DSL</summary>
              <pre className="bg-light p-2 rounded small overflow-auto" style={{ maxHeight: 500 }}>
                {graphData.mermaid}
              </pre>
            </details>
          </div>
        ) : null}
      </div>
    </div>
  )
}
