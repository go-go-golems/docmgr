import { useMemo, useState } from 'react'
import { Link, useParams, useSearchParams } from 'react-router-dom'

import { ApiErrorAlert } from '../../components/ApiErrorAlert'
import { EmptyState } from '../../components/EmptyState'
import { LoadingSpinner } from '../../components/LoadingSpinner'
import { timeAgo } from '../../lib/time'
import {
  type WorkspaceTicketListItem,
  useGetWorkspaceFacetsQuery,
  useGetWorkspaceTopicQuery,
} from '../../services/docmgrApi'

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

function setParam(sp: URLSearchParams, key: string, value: string) {
  if (value.trim()) sp.set(key, value.trim())
  else sp.delete(key)
}

function setParamBool(sp: URLSearchParams, key: string, value: boolean) {
  if (value) sp.set(key, '1')
  else sp.delete(key)
}

function groupTickets(items: WorkspaceTicketListItem[]): Record<string, WorkspaceTicketListItem[]> {
  const out: Record<string, WorkspaceTicketListItem[]> = {}
  for (const t of items) {
    const k = (t.status ?? '').trim() || 'unknown'
    out[k] ||= []
    out[k].push(t)
  }
  return out
}

export function WorkspaceTopicDetailPage() {
  const params = useParams()
  const topic = (params.topic ?? '').trim()

  const [searchParams, setSearchParams] = useSearchParams()
  const includeArchived = getParamBool(searchParams, 'archived', true)
  const statusFilter = (searchParams.get('status') ?? '').trim()
  const ownerFilter = (searchParams.get('owner') ?? '').trim()
  const intentFilter = (searchParams.get('intent') ?? '').trim()

  const { data, isLoading, error } = useGetWorkspaceTopicQuery(
    { topic, includeArchived, docsLimit: 50 },
    { skip: topic === '' },
  )
  const { data: facets } = useGetWorkspaceFacetsQuery({ includeArchived })

  const [expanded, setExpanded] = useState<Record<string, boolean>>({
    active: true,
    review: true,
    complete: false,
    draft: false,
    unknown: false,
  })

  const filteredTickets = useMemo(() => {
    const list = data?.tickets ?? []
    return list.filter((t) => {
      if (statusFilter && t.status !== statusFilter) return false
      if (ownerFilter && !(t.owners ?? []).includes(ownerFilter)) return false
      if (intentFilter && t.intent !== intentFilter) return false
      return true
    })
  }, [data, statusFilter, ownerFilter, intentFilter])

  const ticketsByStatus = useMemo(() => groupTickets(filteredTickets), [filteredTickets])
  const statusKeys = useMemo(() => Object.keys(ticketsByStatus).sort(), [ticketsByStatus])

  if (topic === '') return <div className="alert alert-info">Missing topic.</div>
  if (isLoading) return <LoadingSpinner />
  if (error) return <ApiErrorAlert title="Failed to load topic detail" error={error} />

  const stats = data?.stats
  const docs = data?.docs ?? []

  return (
    <div className="vstack gap-3">
      <div>
        <Link to="/workspace/topics" className="text-decoration-none">
          ← Back to Topics
        </Link>
      </div>

      <div className="card">
        <div className="card-header fw-semibold">Topic: {topic}</div>
        <div className="card-body">
          {stats ? (
            <div className="d-flex flex-wrap gap-2">
              <span className="badge text-bg-light text-dark">
                Tickets: <span className="fw-semibold">{stats.ticketsTotal}</span>
              </span>
              <span className="badge text-bg-light text-dark">
                Recent docs: <span className="fw-semibold">{docs.length}</span>
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
          ) : (
            <div className="text-muted small">No stats available.</div>
          )}
        </div>
      </div>

      <div className="card">
        <div className="card-header fw-semibold d-flex justify-content-between align-items-center">
          <span>Filters</span>
          <div className="d-flex gap-2 align-items-center">
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
                id="topic-include-archived"
              />
              <label className="form-check-label" htmlFor="topic-include-archived">
                Include archived
              </label>
            </div>

            <button
              className="btn btn-sm btn-outline-secondary"
              onClick={() => {
                const sp = new URLSearchParams(searchParams)
                sp.delete('status')
                sp.delete('owner')
                sp.delete('intent')
                setSearchParams(sp, { replace: true })
              }}
            >
              Clear
            </button>
          </div>
        </div>
        <div className="card-body">
          <div className="row g-2">
            <div className="col-12 col-md-4">
              <label className="form-label">Status</label>
              <select
                className="form-select"
                value={statusFilter}
                onChange={(e) => {
                  const sp = new URLSearchParams(searchParams)
                  setParam(sp, 'status', e.target.value)
                  setSearchParams(sp, { replace: true })
                }}
              >
                <option value="">All</option>
                {(facets?.statuses ?? []).map((s) => (
                  <option key={s} value={s}>
                    {s}
                  </option>
                ))}
              </select>
            </div>

            <div className="col-12 col-md-4">
              <label className="form-label">Owner</label>
              <select
                className="form-select"
                value={ownerFilter}
                onChange={(e) => {
                  const sp = new URLSearchParams(searchParams)
                  setParam(sp, 'owner', e.target.value)
                  setSearchParams(sp, { replace: true })
                }}
              >
                <option value="">All</option>
                {(facets?.owners ?? []).map((o) => (
                  <option key={o} value={o}>
                    {o}
                  </option>
                ))}
              </select>
            </div>

            <div className="col-12 col-md-4">
              <label className="form-label">Intent</label>
              <select
                className="form-select"
                value={intentFilter}
                onChange={(e) => {
                  const sp = new URLSearchParams(searchParams)
                  setParam(sp, 'intent', e.target.value)
                  setSearchParams(sp, { replace: true })
                }}
              >
                <option value="">All</option>
                {(facets?.intents ?? []).map((i) => (
                  <option key={i} value={i}>
                    {i}
                  </option>
                ))}
              </select>
            </div>
          </div>

          <div className="text-muted small mt-2">{filteredTickets.length} tickets match current filters.</div>
        </div>
      </div>

      <div className="vstack gap-3">
        {statusKeys.length === 0 ? (
          <EmptyState title="No tickets found" />
        ) : (
          statusKeys.map((k) => (
            <div key={k} className="card">
              <div className="card-header fw-semibold d-flex justify-content-between align-items-center">
                <span>
                  {k.toUpperCase()} ({ticketsByStatus[k]?.length ?? 0})
                </span>
                <button
                  className="btn btn-sm btn-outline-secondary"
                  onClick={() => setExpanded((prev) => ({ ...prev, [k]: !(prev[k] ?? false) }))}
                >
                  {expanded[k] ? 'Collapse' : 'Expand'}
                </button>
              </div>
              {expanded[k] ? (
                <div className="card-body">
                  <div className="list-group">
                    {(ticketsByStatus[k] ?? []).slice(0, 50).map((t) => (
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
                            <span className="text-muted small">{t.updatedAt ? timeAgo(t.updatedAt) : '—'}</span>
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
                  {(ticketsByStatus[k]?.length ?? 0) > 50 ? (
                    <div className="text-muted small mt-2">
                      Showing first 50. Use the Tickets page for full browsing.
                    </div>
                  ) : null}
                </div>
              ) : null}
            </div>
          ))
        )}
      </div>

      <div className="card">
        <div className="card-header fw-semibold">Recent documents</div>
        <div className="card-body">
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
  )
}
