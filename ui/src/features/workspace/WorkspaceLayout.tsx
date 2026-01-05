import { NavLink, Outlet } from 'react-router-dom'

import { PageHeader } from '../../components/PageHeader'
import { useToast } from '../toast/useToast'
import { timeAgo } from '../../lib/time'
import { useGetWorkspaceStatusQuery, useGetWorkspaceSummaryQuery, useRefreshIndexMutation } from '../../services/docmgrApi'

function ShellNav() {
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
      {navLink('/search', 'Search')}
      {navLink('/workspace/topics', 'Topics')}
      {navLink('/workspace/recent', 'Recent')}
    </div>
  )
}

export function WorkspaceLayout() {
  const toast = useToast()
  const { data: wsStatus, isError: wsError, refetch } = useGetWorkspaceStatusQuery()
  const { data: summary } = useGetWorkspaceSummaryQuery(undefined, { skip: wsError })
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
            <NavLink className="btn btn-outline-primary" to="/search" end>
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
                  {summary?.stats ? (
                    <>
                      <span className="badge text-bg-light text-dark">
                        Tickets: <span className="fw-semibold">{summary.stats.ticketsTotal}</span>
                      </span>
                      <span className="badge text-bg-primary">
                        Active: <span className="fw-semibold">{summary.stats.ticketsActive}</span>
                      </span>
                      <span className="badge text-bg-warning text-dark">
                        Review: <span className="fw-semibold">{summary.stats.ticketsReview}</span>
                      </span>
                      <span className="badge text-bg-success">
                        Complete: <span className="fw-semibold">{summary.stats.ticketsComplete}</span>
                      </span>
                      <span className="badge text-bg-secondary">
                        Draft: <span className="fw-semibold">{summary.stats.ticketsDraft}</span>
                      </span>
                    </>
                  ) : null}
                </div>
              ) : (
                <div className="text-muted small">Loading…</div>
              )}

              <div className="d-grid gap-2 mt-3">
                <NavLink className="btn btn-outline-primary btn-sm" to="/workspace/tickets">
                  All tickets →
                </NavLink>
                <NavLink className="btn btn-outline-primary btn-sm" to="/workspace/topics">
                  Topics →
                </NavLink>
                <NavLink className="btn btn-outline-secondary btn-sm" to="/workspace/recent">
                  Recent →
                </NavLink>
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
