import { Link, useSearchParams } from 'react-router-dom'

import { ApiErrorAlert } from '../../components/ApiErrorAlert'
import { EmptyState } from '../../components/EmptyState'
import { LoadingSpinner } from '../../components/LoadingSpinner'
import { timeAgo } from '../../lib/time'
import { useGetWorkspaceTopicsQuery } from '../../services/docmgrApi'

function getParamBool(sp: URLSearchParams, key: string, def: boolean): boolean {
  const v = (sp.get(key) ?? '').trim().toLowerCase()
  if (!v) return def
  return v === '1' || v === 'true' || v === 't' || v === 'yes' || v === 'y' || v === 'on'
}

function setParamBool(sp: URLSearchParams, key: string, value: boolean) {
  if (value) sp.set(key, '1')
  else sp.delete(key)
}

export function WorkspaceTopicsPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const includeArchived = getParamBool(searchParams, 'archived', true)

  const { data, isLoading, error } = useGetWorkspaceTopicsQuery({ includeArchived })

  if (isLoading) return <LoadingSpinner />
  if (error) return <ApiErrorAlert title="Failed to load topics" error={error} />

  const topics = data?.results ?? []

  return (
    <div className="vstack gap-3">
      <div className="card">
        <div className="card-header fw-semibold d-flex justify-content-between align-items-center">
          <span>Topics</span>
          <div className="form-check">
            <input
              className="form-check-input"
              type="checkbox"
              checked={includeArchived}
              onChange={(e) => {
                const sp = new URLSearchParams(searchParams)
                setParamBool(sp, 'archived', e.target.checked)
                setSearchParams(sp, { replace: true })
              }}
              id="topics-include-archived"
            />
            <label className="form-check-label" htmlFor="topics-include-archived">
              Include archived
            </label>
          </div>
        </div>
        <div className="card-body">
          {topics.length === 0 ? (
            <EmptyState title="No topics found" />
          ) : (
            <>
              <div className="text-muted small mb-3">{topics.length} topics</div>
              <div className="row g-3">
                {topics.map((t) => (
                  <div key={t.topic} className="col-12 col-md-6 col-lg-4">
                    <div className="card h-100">
                      <div className="card-body">
                        <div className="fw-semibold mb-2">🏷️ {t.topic}</div>
                        <div className="d-flex flex-wrap gap-2">
                          <span className="badge text-bg-light text-dark">
                            Tickets: <span className="fw-semibold">{t.ticketsTotal}</span>
                          </span>
                          <span className="badge text-bg-light text-dark">
                            Docs: <span className="fw-semibold">{t.docsTotal}</span>
                          </span>
                        </div>
                        {t.updatedAt ? (
                          <div className="text-muted small mt-2">Updated {timeAgo(t.updatedAt)}</div>
                        ) : null}
                      </div>
                      <div className="card-footer bg-transparent">
                        <Link to={`/workspace/topics/${encodeURIComponent(t.topic)}`} className="btn btn-outline-primary btn-sm">
                          Browse →
                        </Link>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </>
          )}
        </div>
      </div>
    </div>
  )
}
