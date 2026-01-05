import { timeAgo } from '../../../lib/time'
import type { WorkspaceStatus } from '../../../services/docmgrApi'
import { PageHeader } from '../../../components/PageHeader'
import { Link } from 'react-router-dom'

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
      <PageHeader
        title="docmgr Search"
        titleClassName="h3"
        mb={4}
        actions={
          <>
            <Link className="btn btn-outline-primary" to="/workspace">
              Workspace
            </Link>
            <button
              className="btn btn-outline-secondary refresh-btn"
              onClick={onRefresh}
              disabled={refreshLoading}
            >
              {refreshLoading ? 'Refreshing…' : `Refresh (${timeAgo(wsStatus?.indexedAt)})`}
            </button>
          </>
        }
      />

      {wsError ? (
        <div className="alert alert-warning">
          Workspace status unavailable. Is the server running on <code>127.0.0.1:3001</code>?
        </div>
      ) : null}
    </>
  )
}
