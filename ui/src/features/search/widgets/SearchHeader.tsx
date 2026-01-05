import { timeAgo } from '../../../lib/time'
import type { WorkspaceStatus } from '../../../services/docmgrApi'

export function SearchHeader({
  wsStatus,
  wsError,
  refreshLoading,
  onRefresh,
}: {
  wsStatus?: WorkspaceStatus
  wsError: boolean
  refreshLoading: boolean
  onRefresh: () => void
}) {
  return (
    <>
      <div className="d-flex justify-content-between align-items-center mb-4">
        <h1 className="h3 mb-0">docmgr Search</h1>
        <button
          className="btn btn-outline-secondary refresh-btn"
          onClick={onRefresh}
          disabled={refreshLoading}
        >
          {refreshLoading ? 'Refreshing…' : `Refresh (${timeAgo(wsStatus?.indexedAt)})`}
        </button>
      </div>

      {wsError ? (
        <div className="alert alert-warning">
          Workspace status unavailable. Is the server running on <code>127.0.0.1:3001</code>?
        </div>
      ) : null}
    </>
  )
}

