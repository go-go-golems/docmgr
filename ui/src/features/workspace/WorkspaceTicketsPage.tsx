import { useMemo, useState } from 'react'
import { Link, useSearchParams } from 'react-router-dom'

import { ApiErrorAlert } from '../../components/ApiErrorAlert'
import { EmptyState } from '../../components/EmptyState'
import { LoadingSpinner } from '../../components/LoadingSpinner'
import { timeAgo } from '../../lib/time'
import {
  type WorkspaceTicketListItem,
  useGetWorkspaceFacetsQuery,
  useGetWorkspaceTicketsQuery,
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

function setParam(sp: URLSearchParams, key: string, value: string) {
  if (value.trim()) sp.set(key, value.trim())
  else sp.delete(key)
}

function setParamBool(sp: URLSearchParams, key: string, value: boolean) {
  if (value) sp.set(key, '1')
  else sp.delete(key)
}

function getParamBool(sp: URLSearchParams, key: string, def: boolean): boolean {
  const v = (sp.get(key) ?? '').trim().toLowerCase()
  if (!v) return def
  return v === '1' || v === 'true' || v === 't' || v === 'yes' || v === 'y' || v === 'on'
}

function TicketRow({ t }: { t: WorkspaceTicketListItem }) {
  return (
    <tr>
      <td className="font-monospace">
        <Link to={`/ticket/${encodeURIComponent(t.ticket)}`} className="text-decoration-none">
          {t.ticket}
        </Link>
      </td>
      <td>{t.title}</td>
      <td>
        <StatusBadge status={t.status} />
      </td>
      <td>
        <div className="d-flex flex-wrap gap-1">
          {(t.topics ?? []).slice(0, 5).map((topic) => (
            <span key={topic} className="badge text-bg-secondary dm-topic-badge">
              {topic}
            </span>
          ))}
        </div>
      </td>
      <td>
        <div className="d-flex flex-wrap gap-1">
          {(t.owners ?? []).slice(0, 4).map((o) => (
            <span key={o} className="badge text-bg-light text-dark">
              {o}
            </span>
          ))}
        </div>
      </td>
      <td className="text-muted">{t.intent}</td>
      <td className="text-muted">{t.updatedAt ? timeAgo(t.updatedAt) : '—'}</td>
    </tr>
  )
}

export function WorkspaceTicketsPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const [cursor, setCursor] = useState('')

  const q = (searchParams.get('q') ?? '').trim()
  const status = (searchParams.get('status') ?? '').trim()
  const topic = (searchParams.get('topic') ?? '').trim()
  const owner = (searchParams.get('owner') ?? '').trim()
  const intent = (searchParams.get('intent') ?? '').trim()
  const includeArchived = getParamBool(searchParams, 'archived', true)
  const includeStats = getParamBool(searchParams, 'stats', false)

  const { data: facets, error: facetsError, isLoading: facetsLoading } = useGetWorkspaceFacetsQuery({
    includeArchived,
  })

  const ticketQueryKey = useMemo(
    () => ({
      q,
      status,
      topics: topic ? [topic] : [],
      owners: owner ? [owner] : [],
      intent,
      includeArchived,
      includeStats,
    }),
    [q, status, topic, owner, intent, includeArchived, includeStats],
  )

  const {
    data: ticketsData,
    error: ticketsError,
    isLoading: ticketsLoading,
    isFetching: ticketsFetching,
  } = useGetWorkspaceTicketsQuery({
    ...ticketQueryKey,
    orderBy: 'last_updated',
    reverse: true,
    pageSize: 200,
    cursor,
  })

  const tickets = ticketsData?.results ?? []
  const nextCursor = ticketsData?.nextCursor ?? ''

  function onClear() {
    setCursor('')
    setSearchParams(new URLSearchParams(), { replace: true })
  }

  return (
    <div className="vstack gap-3">
      <div className="card">
        <div className="card-header fw-semibold d-flex justify-content-between align-items-center">
          <span>Filters</span>
          <button className="btn btn-sm btn-outline-secondary" onClick={onClear}>
            Clear
          </button>
        </div>
        <div className="card-body">
          <div className="row g-2">
            <div className="col-12 col-md-6 col-lg-4">
              <label className="form-label">Query</label>
              <input
                className="form-control"
                value={q}
                onChange={(e) => {
                  setCursor('')
                  const sp = new URLSearchParams(searchParams)
                  setParam(sp, 'q', e.target.value)
                  setSearchParams(sp, { replace: true })
                }}
                placeholder="FTS query (optional)"
              />
              <div className="form-text">Uses SQLite FTS5 `MATCH` syntax (when available).</div>
            </div>

            <div className="col-12 col-md-6 col-lg-2">
              <label className="form-label">Status</label>
              <select
                className="form-select"
                value={status}
                onChange={(e) => {
                  setCursor('')
                  const sp = new URLSearchParams(searchParams)
                  setParam(sp, 'status', e.target.value)
                  setSearchParams(sp, { replace: true })
                }}
                disabled={facetsLoading}
              >
                <option value="">All</option>
                {(facets?.statuses ?? []).map((s) => (
                  <option key={s} value={s}>
                    {s}
                  </option>
                ))}
              </select>
            </div>

            <div className="col-12 col-md-6 col-lg-2">
              <label className="form-label">Topic</label>
              <select
                className="form-select"
                value={topic}
                onChange={(e) => {
                  setCursor('')
                  const sp = new URLSearchParams(searchParams)
                  setParam(sp, 'topic', e.target.value)
                  setSearchParams(sp, { replace: true })
                }}
                disabled={facetsLoading}
              >
                <option value="">All</option>
                {(facets?.topics ?? []).map((t) => (
                  <option key={t} value={t}>
                    {t}
                  </option>
                ))}
              </select>
            </div>

            <div className="col-12 col-md-6 col-lg-2">
              <label className="form-label">Owner</label>
              <select
                className="form-select"
                value={owner}
                onChange={(e) => {
                  setCursor('')
                  const sp = new URLSearchParams(searchParams)
                  setParam(sp, 'owner', e.target.value)
                  setSearchParams(sp, { replace: true })
                }}
                disabled={facetsLoading}
              >
                <option value="">All</option>
                {(facets?.owners ?? []).map((o) => (
                  <option key={o} value={o}>
                    {o}
                  </option>
                ))}
              </select>
            </div>

            <div className="col-12 col-md-6 col-lg-2">
              <label className="form-label">Intent</label>
              <select
                className="form-select"
                value={intent}
                onChange={(e) => {
                  setCursor('')
                  const sp = new URLSearchParams(searchParams)
                  setParam(sp, 'intent', e.target.value)
                  setSearchParams(sp, { replace: true })
                }}
                disabled={facetsLoading}
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

          <div className="d-flex flex-wrap gap-3 align-items-center mt-3">
            <div className="form-check">
              <input
                className="form-check-input"
                type="checkbox"
                checked={includeArchived}
                onChange={(e) => {
                  setCursor('')
                  const sp = new URLSearchParams(searchParams)
                  setParamBool(sp, 'archived', e.target.checked)
                  setSearchParams(sp, { replace: true })
                }}
                id="tickets-include-archived"
              />
              <label className="form-check-label" htmlFor="tickets-include-archived">
                Include archived
              </label>
            </div>

            <div className="form-check">
              <input
                className="form-check-input"
                type="checkbox"
                checked={includeStats}
                onChange={(e) => {
                  setCursor('')
                  const sp = new URLSearchParams(searchParams)
                  setParamBool(sp, 'stats', e.target.checked)
                  setSearchParams(sp, { replace: true })
                }}
                id="tickets-include-stats"
              />
              <label className="form-check-label" htmlFor="tickets-include-stats">
                Include stats (slower)
              </label>
            </div>

            {facetsError ? <span className="text-danger small">Facets failed to load.</span> : null}
          </div>
        </div>
      </div>

      <div className="card">
        <div className="card-header fw-semibold d-flex justify-content-between align-items-center">
          <span>Tickets</span>
          <span className="text-muted small">
            {ticketsData ? (
              <>
                {tickets.length} / {ticketsData.total}
              </>
            ) : (
              '—'
            )}
          </span>
        </div>
        <div className="card-body">
          {ticketsError ? <ApiErrorAlert title="Failed to load tickets" error={ticketsError} /> : null}
          {ticketsLoading ? <LoadingSpinner /> : null}

          {!ticketsLoading && !ticketsError ? (
            tickets.length === 0 ? (
              <EmptyState title="No tickets found">
                <p className="mb-0">Try clearing filters or changing the query.</p>
              </EmptyState>
            ) : (
              <div className="table-responsive">
                <table className="table table-sm align-middle">
                  <thead>
                    <tr>
                      <th>Ticket</th>
                      <th>Title</th>
                      <th>Status</th>
                      <th>Topics</th>
                      <th>Owners</th>
                      <th>Intent</th>
                      <th>Updated</th>
                    </tr>
                  </thead>
                  <tbody>
                    {tickets.map((t) => (
                      <TicketRow key={t.ticket} t={t} />
                    ))}
                  </tbody>
                </table>
              </div>
            )
          ) : null}

          {ticketsFetching && !ticketsLoading ? <div className="text-muted small">Loading…</div> : null}

          {nextCursor && !ticketsLoading && !ticketsError ? (
            <button
              className="btn btn-outline-primary"
              onClick={() => setCursor(nextCursor)}
              disabled={ticketsFetching}
            >
              {ticketsFetching ? 'Loading…' : 'Load more'}
            </button>
          ) : null}

          {!nextCursor && tickets.length > 0 && !ticketsLoading && !ticketsError ? (
            <div className="text-muted small">End of results.</div>
          ) : null}

          {/* Helpful: show the active query in a copy/paste friendly way. */}
          {ticketsData?.query ? (
            <details className="mt-3">
              <summary className="small">Query debug</summary>
              <pre className="small mb-0">{JSON.stringify(ticketsData.query, null, 2)}</pre>
            </details>
          ) : null}
        </div>
      </div>
    </div>
  )
}
