import { NavLink, Outlet, useLocation } from 'react-router-dom'

import { PageHeader } from '../../components/PageHeader'
import { useToast } from '../toast/useToast'
import { timeAgo } from '../../lib/time'
import { useGetWorkspaceStatusQuery, useRefreshIndexMutation } from '../../services/docmgrApi'

function ShellNav() {
  const location = useLocation()
  const inWorkspace = location.pathname.startsWith('/workspace')

  const navLink = (to: string, label: string) => (
    <NavLink
      to={to}
      className={({ isActive }) =>
        ['list-group-item list-group-item-action', isActive ? 'active' : ''].filter(Boolean).join(' ')
      }
      end
    >
      {label}
    </NavLink>
  )

  return (
    <div className="list-group">
      {navLink('/workspace', 'Home')}
      {navLink('/workspace/tickets', 'Tickets')}
      <NavLink
        to="/"
        className={['list-group-item list-group-item-action', !inWorkspace ? 'active' : ''].filter(Boolean).join(' ')}
        end
      >
        Search
      </NavLink>
      {navLink('/workspace/topics', 'Topics')}
      {navLink('/workspace/recent', 'Recent')}
    </div>
  )
}

export function WorkspaceLayout() {
  const toast = useToast()
  const { data: wsStatus, isError: wsError, refetch } = useGetWorkspaceStatusQuery()
  const [refreshIndex, refreshState] = useRefreshIndexMutation()

  async function onRefresh() {
    try {
      await refreshIndex().unwrap()
      await refetch()
      toast.success('Index refreshed successfully', { timeoutMs: 2000 })
    } catch (e) {
      toast.error(`Index refresh failed: ${String(e)}`, { timeoutMs: 2500 })
    }
  }

  return (
    <div className="container py-4">
      <PageHeader
        title="docmgr"
        titleClassName="h3"
        actions={
          <>
            <NavLink className="btn btn-outline-primary" to="/" end>
              Search
            </NavLink>
            <button className="btn btn-outline-secondary" onClick={() => void onRefresh()} disabled={refreshState.isLoading}>
              {refreshState.isLoading
                ? 'Refreshing…'
                : `Refresh (${timeAgo(wsStatus?.indexedAt)})`}
            </button>
          </>
        }
      />

      {wsError ? (
        <div className="alert alert-warning">
          Workspace status unavailable. Is the server running on <code>127.0.0.1:3001</code>?
        </div>
      ) : null}

      <div className="row g-3">
        <div className="col-12 col-lg-3">
          <ShellNav />
          <div className="card mt-3">
            <div className="card-header fw-semibold">Quick stats</div>
            <div className="card-body">
              {wsStatus ? (
                <div className="d-flex flex-wrap gap-2">
                  <span className="badge text-bg-light text-dark">
                    Docs: <span className="fw-semibold">{wsStatus.docsIndexed}</span>
                  </span>
                </div>
              ) : (
                <div className="text-muted small">Loading…</div>
              )}
              <div className="text-muted small mt-2">
                Ticket/topic/activity stats need workspace summary endpoints (see design doc).
              </div>
            </div>
          </div>
        </div>

        <div className="col-12 col-lg-9">
          <Outlet />
        </div>
      </div>
    </div>
  )
}

