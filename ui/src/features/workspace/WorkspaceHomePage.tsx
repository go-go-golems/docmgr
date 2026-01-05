import { timeAgo } from '../../lib/time'
import { useGetWorkspaceStatusQuery } from '../../services/docmgrApi'
import { EmptyState } from '../../components/EmptyState'
import { LoadingSpinner } from '../../components/LoadingSpinner'
import { ApiErrorAlert } from '../../components/ApiErrorAlert'

export function WorkspaceHomePage() {
  const { data: wsStatus, isLoading, error } = useGetWorkspaceStatusQuery()

  if (isLoading) return <LoadingSpinner />
  if (error) return <ApiErrorAlert title="Failed to load workspace status" error={error} />

  return (
    <div className="vstack gap-3">
      <div className="card">
        <div className="card-header fw-semibold">Workspace overview</div>
        <div className="card-body">
          <div className="d-flex flex-wrap gap-2">
            <span className="badge text-bg-light text-dark">
              Indexed: <span className="fw-semibold">{timeAgo(wsStatus?.indexedAt)}</span>
            </span>
            <span className="badge text-bg-light text-dark">
              Documents: <span className="fw-semibold">{wsStatus?.docsIndexed ?? 0}</span>
            </span>
          </div>
          {wsStatus?.root ? (
            <div className="mt-2 text-muted small">
              Root: <span className="font-monospace">{wsStatus.root}</span>
            </div>
          ) : null}
        </div>
      </div>

      <div className="card">
        <div className="card-header fw-semibold d-flex justify-content-between align-items-center">
          <span>Recent activity</span>
        </div>
        <div className="card-body">
          <EmptyState title="Not implemented yet">
            <p className="mb-0">
              This widget needs the workspace activity endpoint from `design/03-workspace-rest-api.md`.
            </p>
          </EmptyState>
        </div>
      </div>
    </div>
  )
}
