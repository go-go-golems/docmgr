import { EmptyState } from '../../components/EmptyState'

export function WorkspaceRecentPage() {
  return (
    <div className="card">
      <div className="card-header fw-semibold">Recent activity</div>
      <div className="card-body">
        <EmptyState title="Not implemented yet">
          <p className="mb-0">
            This page needs the workspace activity endpoint from `design/03-workspace-rest-api.md`.
          </p>
        </EmptyState>
      </div>
    </div>
  )
}

