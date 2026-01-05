import { EmptyState } from '../../components/EmptyState'

export function WorkspaceTopicsPage() {
  return (
    <div className="card">
      <div className="card-header fw-semibold">Topics</div>
      <div className="card-body">
        <EmptyState title="Not implemented yet">
          <p className="mb-0">
            This page needs the workspace topics endpoints from `design/03-workspace-rest-api.md`.
          </p>
        </EmptyState>
      </div>
    </div>
  )
}

