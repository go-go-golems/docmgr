import { Link, useSearchParams } from 'react-router-dom'

import { ApiErrorAlert } from '../../components/ApiErrorAlert'
import { EmptyState } from '../../components/EmptyState'
import { LoadingSpinner } from '../../components/LoadingSpinner'
import { timeAgo } from '../../lib/time'
import { useGetWorkspaceRecentQuery } from '../../services/docmgrApi'

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

function getParamBool(sp: URLSearchParams, key: string, def: boolean): boolean {
  const v = (sp.get(key) ?? '').trim().toLowerCase()
  if (!v) return def
  return v === '1' || v === 'true' || v === 't' || v === 'yes' || v === 'y' || v === 'on'
}

function setParamBool(sp: URLSearchParams, key: string, value: boolean) {
  if (value) sp.set(key, '1')
  else sp.delete(key)
}

export function WorkspaceRecentPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const includeArchived = getParamBool(searchParams, 'archived', true)

  const { data, isLoading, error } = useGetWorkspaceRecentQuery({
    includeArchived,
    ticketsLimit: 50,
    docsLimit: 50,
  })

  if (isLoading) return <LoadingSpinner />
  if (error) return <ApiErrorAlert title="Failed to load recent activity" error={error} />

  const tickets = data?.tickets ?? []
  const docs = data?.docs ?? []

  return (
    <div className="vstack gap-3">
      <div className="card">
        <div className="card-header fw-semibold d-flex justify-content-between align-items-center">
          <span>Recent activity</span>
          <div className="form-check mb-0">
            <input
              className="form-check-input"
              type="checkbox"
              checked={includeArchived}
              onChange={(e) => {
                const sp = new URLSearchParams(searchParams)
                setParamBool(sp, 'archived', e.target.checked)
                setSearchParams(sp, { replace: true })
              }}
              id="recent-include-archived"
            />
            <label className="form-check-label" htmlFor="recent-include-archived">
              Include archived
            </label>
          </div>
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
              {tickets.length === 0 ? (
                <EmptyState title="No tickets found" />
              ) : (
                <div className="list-group">
                  {tickets.map((t) => (
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
                          {t.topics.slice(0, 8).map((topic) => (
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
              {docs.length === 0 ? (
                <EmptyState title="No documents found" />
              ) : (
                <div className="list-group">
                  {docs.map((d) => (
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
