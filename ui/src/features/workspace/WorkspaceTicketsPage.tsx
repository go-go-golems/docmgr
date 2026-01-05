import { EmptyState } from '../../components/EmptyState'

export function WorkspaceTicketsPage() {
  return (
    <div className="vstack gap-3">
      <div className="card">
        <div className="card-header fw-semibold">Filters</div>
        <div className="card-body">
          <div className="row g-2">
            <div className="col-12 col-md-4">
              <label className="form-label">Status</label>
              <select className="form-select" disabled>
                <option>All</option>
              </select>
            </div>
            <div className="col-12 col-md-4">
              <label className="form-label">Owner</label>
              <select className="form-select" disabled>
                <option>All</option>
              </select>
            </div>
            <div className="col-12 col-md-4">
              <label className="form-label">Intent</label>
              <select className="form-select" disabled>
                <option>All</option>
              </select>
            </div>
          </div>
          <div className="text-muted small mt-2">
            Filters will be enabled once the Workspace Tickets list endpoint is implemented.
          </div>
        </div>
      </div>

      <div className="card">
        <div className="card-header fw-semibold">Tickets</div>
        <div className="card-body">
          <EmptyState title="Not implemented yet">
            <p className="mb-0">
              This page needs the workspace tickets list endpoint from `design/03-workspace-rest-api.md`.
            </p>
          </EmptyState>
        </div>
      </div>
    </div>
  )
}

