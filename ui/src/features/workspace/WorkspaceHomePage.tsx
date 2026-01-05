import { Link } from 'react-router-dom'

import { timeAgo } from '../../lib/time'
import { useGetWorkspaceSummaryQuery } from '../../services/docmgrApi'
import { LoadingSpinner } from '../../components/LoadingSpinner'
import { ApiErrorAlert } from '../../components/ApiErrorAlert'

function StatusBadge({ status }: { status: string }) {
  const variant =
    status === 'active'
      ? 'primary'
      : status === 'review'
        ? 'warning'
        : status === 'complete'
          ? 'success'
          : status === 'draft'
            ? 'secondary'
            : 'secondary'
  return (
    <span className={`badge text-bg-${variant}`} style={{ fontWeight: 600 }}>
      {status || 'unknown'}
    </span>
  )
}

export function WorkspaceHomePage() {
  const { data, isLoading, error } = useGetWorkspaceSummaryQuery()

  if (isLoading) return <LoadingSpinner />
  if (error) return <ApiErrorAlert title="Failed to load workspace summary" error={error} />

  const stats = data?.stats
  const recentTickets = data?.recent?.tickets ?? []
  const recentDocs = data?.recent?.docs ?? []

  return (
    <div className="vstack gap-3">
      <div className="card">
        <div className="card-header fw-semibold">Workspace overview</div>
        <div className="card-body">
          <div className="d-flex flex-wrap gap-2">
            <span className="badge text-bg-light text-dark">
              Indexed: <span className="fw-semibold">{timeAgo(data?.indexedAt)}</span>
            </span>
            <span className="badge text-bg-light text-dark">
              Documents: <span className="fw-semibold">{data?.docsIndexed ?? 0}</span>
            </span>
          </div>
          <div className="mt-2 text-muted small">
            Root: <span className="font-monospace">{data?.root ?? '—'}</span>
          </div>
          <div className="text-muted small">
            Repo: <span className="font-monospace">{data?.repoRoot ?? '—'}</span>
          </div>
        </div>
      </div>

      <div className="card">
        <div className="card-header fw-semibold">Ticket stats</div>
        <div className="card-body">
          {stats ? (
            <>
              <div className="d-flex flex-wrap gap-2 mb-3">
                <span className="badge text-bg-light text-dark">
                  Tickets: <span className="fw-semibold">{stats.ticketsTotal}</span>
                </span>
                <span className="badge text-bg-primary">
                  Active: <span className="fw-semibold">{stats.ticketsActive}</span>
                </span>
                <span className="badge text-bg-warning text-dark">
                  Review: <span className="fw-semibold">{stats.ticketsReview}</span>
                </span>
                <span className="badge text-bg-success">
                  Complete: <span className="fw-semibold">{stats.ticketsComplete}</span>
                </span>
                <span className="badge text-bg-secondary">
                  Draft: <span className="fw-semibold">{stats.ticketsDraft}</span>
                </span>
              </div>

              <div className="progress" role="progressbar" aria-label="Tickets by status">
                {stats.ticketsTotal > 0 ? (
                  <>
                    <div
                      className="progress-bar bg-primary"
                      style={{ width: `${(100 * stats.ticketsActive) / stats.ticketsTotal}%` }}
                    />
                    <div
                      className="progress-bar bg-warning"
                      style={{ width: `${(100 * stats.ticketsReview) / stats.ticketsTotal}%` }}
                    />
                    <div
                      className="progress-bar bg-success"
                      style={{ width: `${(100 * stats.ticketsComplete) / stats.ticketsTotal}%` }}
                    />
                    <div
                      className="progress-bar bg-secondary"
                      style={{ width: `${(100 * stats.ticketsDraft) / stats.ticketsTotal}%` }}
                    />
                  </>
                ) : null}
              </div>

              <div className="mt-3">
                <Link to="/workspace/tickets" className="btn btn-outline-primary">
                  View all tickets →
                </Link>
              </div>
            </>
          ) : (
            <div className="text-muted small">No stats available.</div>
          )}
        </div>
      </div>

      <div className="card">
        <div className="card-header fw-semibold d-flex justify-content-between align-items-center">
          <span>Recent activity</span>
          <Link to="/workspace/recent" className="btn btn-sm btn-outline-secondary">
            View all →
          </Link>
        </div>
        <div className="card-body">
          <div className="row g-3">
            <div className="col-12 col-lg-6">
              <div className="d-flex justify-content-between align-items-center mb-2">
                <div className="fw-semibold">Recently updated tickets</div>
                <Link to="/workspace/tickets" className="small text-decoration-none">
                  Tickets →
                </Link>
              </div>
              {recentTickets.length === 0 ? (
                <div className="text-muted small">No tickets found.</div>
              ) : (
                <div className="list-group">
                  {recentTickets.map((t) => (
                    <Link
                      key={t.ticket}
                      to={`/ticket/${encodeURIComponent(t.ticket)}`}
                      className="list-group-item list-group-item-action"
                    >
                      <div className="d-flex justify-content-between gap-2">
                        <div className="fw-semibold">
                          <span className="font-monospace">{t.ticket}</span>
                          {t.title ? <span>: {t.title}</span> : null}
                        </div>
                        <div className="d-flex gap-2 align-items-center">
                          <StatusBadge status={t.status} />
                          <span className="text-muted small">{timeAgo(t.updatedAt)}</span>
                        </div>
                      </div>
                      {t.topics?.length ? (
                        <div className="mt-1">
                          {t.topics.slice(0, 6).map((topic) => (
                            <span key={topic} className="badge text-bg-secondary dm-topic-badge">
                              {topic}
                            </span>
                          ))}
                        </div>
                      ) : null}
                    </Link>
                  ))}
                </div>
              )}
            </div>

            <div className="col-12 col-lg-6">
              <div className="d-flex justify-content-between align-items-center mb-2">
                <div className="fw-semibold">Recently updated documents</div>
                <Link to="/search" className="small text-decoration-none">
                  Search →
                </Link>
              </div>
              {recentDocs.length === 0 ? (
                <div className="text-muted small">No documents found.</div>
              ) : (
                <div className="list-group">
                  {recentDocs.map((d) => (
                    <Link
                      key={d.path}
                      to={`/doc?path=${encodeURIComponent(d.path)}`}
                      className="list-group-item list-group-item-action"
                    >
                      <div className="fw-semibold">{d.title || d.path}</div>
                      <div className="small text-muted">
                        <span className="font-monospace">{d.ticket}</span> • {d.docType} • {timeAgo(d.updatedAt)}
                      </div>
                    </Link>
                  ))}
                </div>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
