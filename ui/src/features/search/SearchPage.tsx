import { useEffect, useMemo, useRef, useState } from 'react'
import type { FormEvent } from 'react'

import { useAppDispatch, useAppSelector } from '../../app/hooks'
import { clearFilters, setFilter, setMode, setQuery } from './searchSlice'
import {
  useGetWorkspaceStatusQuery,
  useLazySearchDocsQuery,
  useLazySearchFilesQuery,
  useRefreshIndexMutation,
} from '../../services/docmgrApi'
import type { SearchDocResult } from '../../services/docmgrApi'

type ToastState = { kind: 'success' | 'error'; message: string } | null

function timeAgo(iso?: string): string {
  if (!iso) return 'unknown'
  const t = new Date(iso)
  const deltaMs = Date.now() - t.getTime()
  if (!Number.isFinite(deltaMs)) return 'unknown'
  const seconds = Math.floor(deltaMs / 1000)
  if (seconds < 10) return 'just now'
  if (seconds < 60) return `${seconds}s ago`
  const minutes = Math.floor(seconds / 60)
  if (minutes < 60) return `${minutes}m ago`
  const hours = Math.floor(minutes / 60)
  if (hours < 48) return `${hours}h ago`
  const days = Math.floor(hours / 24)
  return `${days}d ago`
}

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
    <span className={`badge text-bg-${variant} ms-2`} style={{ fontWeight: 600 }}>
      {status || 'unknown'}
    </span>
  )
}

function ResultCard({
  result,
  onCopyPath,
  onSelect,
}: {
  result: SearchDocResult
  onCopyPath: (path: string) => void
  onSelect: (r: SearchDocResult) => void
}) {
  return (
    <div className="result-card" onClick={() => onSelect(result)} role="button" tabIndex={0}>
      <div className="d-flex justify-content-between align-items-start">
        <div className="flex-grow-1">
          <div className="result-title">{result.title}</div>
          <div className="result-meta">
            {result.ticket} • {result.docType}
            <StatusBadge status={result.status} />
            {result.lastUpdated ? (
              <span className="ms-2 text-muted">Updated {timeAgo(result.lastUpdated)}</span>
            ) : null}
          </div>
          <div className="mb-2">
            {result.topics.map((topic) => (
              <span key={topic} className="badge text-bg-secondary topic-badge">
                {topic}
              </span>
            ))}
          </div>
          <div className="result-snippet">“{result.snippet}”</div>
          <div className="result-path">{result.path}</div>
          {result.relatedFiles && result.relatedFiles.length > 0 ? (
            <div className="mt-2">
              <div className="small text-muted mb-1">Related files</div>
              <ul className="mb-0 small">
                {result.relatedFiles.slice(0, 3).map((rf) => (
                  <li key={`${rf.path}:${rf.note ?? ''}`}>
                    <span className="font-monospace">{rf.path}</span>
                    {rf.note ? <span className="text-muted ms-2">{rf.note}</span> : null}
                  </li>
                ))}
                {result.relatedFiles.length > 3 ? (
                  <li className="text-muted">… {result.relatedFiles.length - 3} more</li>
                ) : null}
              </ul>
            </div>
          ) : null}
        </div>
        <button
          className="btn btn-sm btn-outline-primary copy-btn ms-2"
          onClick={(e) => {
            e.stopPropagation()
            onCopyPath(result.path)
          }}
        >
          Copy
        </button>
      </div>
    </div>
  )
}

function TopicMultiSelect({
  topics,
  onChange,
}: {
  topics: string[]
  onChange: (topics: string[]) => void
}) {
  const [value, setValue] = useState('')

  const add = () => {
    const next = value.trim()
    if (!next) return
    if (topics.includes(next)) {
      setValue('')
      return
    }
    onChange([...topics, next])
    setValue('')
  }

  return (
    <div>
      <div className="d-flex flex-wrap gap-1 mb-2">
        {topics.map((t) => (
          <span key={t} className="badge text-bg-secondary">
            {t}{' '}
            <button
              type="button"
              className="btn btn-sm btn-link p-0 ms-1 text-white"
              style={{ textDecoration: 'none' }}
              onClick={() => onChange(topics.filter((x) => x !== t))}
            >
              ×
            </button>
          </span>
        ))}
      </div>
      <div className="input-group input-group-sm">
        <input
          className="form-control"
          placeholder="Add topic and press Enter"
          value={value}
          onChange={(e) => setValue(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === 'Enter') {
              e.preventDefault()
              add()
            }
          }}
        />
        <button className="btn btn-outline-secondary" type="button" onClick={add}>
          Add
        </button>
      </div>
    </div>
  )
}

export function SearchPage() {
  const dispatch = useAppDispatch()
  const { mode, query, filters } = useAppSelector((s) => s.search)
  const searchInputRef = useRef<HTMLInputElement | null>(null)

  const { data: wsStatus, isError: wsError, refetch: refetchWs } = useGetWorkspaceStatusQuery()
  const [refreshIndex, refreshState] = useRefreshIndexMutation()

  const [triggerSearchDocs, searchDocsState] = useLazySearchDocsQuery()
  const [triggerSearchFiles, searchFilesState] = useLazySearchFilesQuery()

  const [toast, setToast] = useState<ToastState>(null)
  const [showFilters, setShowFilters] = useState(true)
  const [showDiagnostics, setShowDiagnostics] = useState(false)

  const [hasSearched, setHasSearched] = useState(false)

  const [docsResults, setDocsResults] = useState<SearchDocResult[]>([])
  const [docsTotal, setDocsTotal] = useState<number>(0)
  const [docsNextCursor, setDocsNextCursor] = useState<string>('')
  const [docsDiagnostics, setDocsDiagnostics] = useState<unknown[]>([])

  const [selected, setSelected] = useState<SearchDocResult | null>(null)

  const effectiveOrderBy = useMemo(() => {
    if (filters.orderBy) return filters.orderBy
    if (mode === 'reverse') return 'path'
    if (query.trim() !== '') return 'rank'
    return 'path'
  }, [filters.orderBy, mode, query])

  const effectiveOrderBySafe = useMemo(() => {
    if (effectiveOrderBy === 'rank' && query.trim() === '') return 'path'
    return effectiveOrderBy
  }, [effectiveOrderBy, query])

  useEffect(() => {
    const onKeyDown = (e: KeyboardEvent) => {
      if (e.key === '/' && document.activeElement !== searchInputRef.current) {
        e.preventDefault()
        searchInputRef.current?.focus()
      }
      if (e.key === '?' && !e.ctrlKey && !e.metaKey && !e.altKey) {
        e.preventDefault()
        setToast({
          kind: 'success',
          message:
            'Shortcuts: / focus search • Ctrl/Cmd+R refresh index • Esc clear selection',
        })
      }
      if (e.key === 'Escape') {
        setSelected(null)
      }
      if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 'r') {
        e.preventDefault()
        void onRefresh()
      }
    }
    window.addEventListener('keydown', onKeyDown)
    return () => window.removeEventListener('keydown', onKeyDown)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [refreshState.isLoading, wsStatus?.indexedAt])

  useEffect(() => {
    if (!toast) return
    const t = window.setTimeout(() => setToast(null), 2000)
    return () => window.clearTimeout(t)
  }, [toast])

  const onCopyPath = async (path: string) => {
    try {
      await navigator.clipboard.writeText(path)
      setToast({ kind: 'success', message: `Copied path: ${path}` })
    } catch {
      setToast({ kind: 'error', message: 'Failed to copy path (clipboard not available)' })
    }
  }

  const onRefresh = async () => {
    try {
      await refreshIndex().unwrap()
      await refetchWs()
      setToast({ kind: 'success', message: 'Index refreshed successfully' })
    } catch (e) {
      setToast({ kind: 'error', message: `Index refresh failed: ${String(e)}` })
    }
  }

  const doSearchDocs = async (cursor: string, append: boolean) => {
    const textQuery = mode === 'reverse' ? '' : query
    const resp = await triggerSearchDocs({
      query: textQuery,
      ticket: filters.ticket,
      topics: filters.topics,
      docType: filters.docType,
      status: filters.status,
      file: filters.file,
      dir: filters.dir,
      orderBy: effectiveOrderBySafe,
      reverse: mode === 'reverse',
      includeArchived: filters.includeArchived,
      includeScripts: filters.includeScripts,
      includeControlDocs: filters.includeControlDocs,
      includeDiagnostics: true,
      pageSize: 200,
      cursor,
    }).unwrap()

    setDocsTotal(resp.total)
    setDocsNextCursor(resp.nextCursor || '')
    setDocsDiagnostics(resp.diagnostics || [])
    setHasSearched(true)

    if (append) setDocsResults((prev) => [...prev, ...resp.results])
    else setDocsResults(resp.results)
  }

  const onSubmit = async (e: FormEvent) => {
    e.preventDefault()
    setSelected(null)
    setShowDiagnostics(false)

    if (mode === 'files') {
      try {
        setHasSearched(true)
        await triggerSearchFiles({
          query,
          ticket: filters.ticket,
          topics: filters.topics,
          limit: 200,
        }).unwrap()
      } catch (err) {
        setToast({ kind: 'error', message: `Search failed: ${String(err)}` })
      }
      return
    }

    try {
      await doSearchDocs('', false)
    } catch (err) {
      setToast({ kind: 'error', message: `Search failed: ${String(err)}` })
    }
  }

  const onLoadMore = async () => {
    if (!docsNextCursor) return
    try {
      await doSearchDocs(docsNextCursor, true)
    } catch (err) {
      setToast({ kind: 'error', message: `Load more failed: ${String(err)}` })
    }
  }

  const activeChips = useMemo(() => {
    const chips: Array<{ key: string; label: string; onRemove: () => void }> = []
    if (mode !== 'reverse' && query.trim() !== '') {
      chips.push({
        key: 'q',
        label: query.trim(),
        onRemove: () => dispatch(setQuery('')),
      })
    }
    if (filters.ticket.trim() !== '') {
      chips.push({
        key: 'ticket',
        label: `ticket:${filters.ticket.trim()}`,
        onRemove: () => dispatch(setFilter({ key: 'ticket', value: '' })),
      })
    }
    for (const t of filters.topics) {
      chips.push({
        key: `topic:${t}`,
        label: `topic:${t}`,
        onRemove: () =>
          dispatch(setFilter({ key: 'topics', value: filters.topics.filter((x) => x !== t) })),
      })
    }
    if (filters.docType.trim() !== '') {
      chips.push({
        key: 'docType',
        label: `type:${filters.docType.trim()}`,
        onRemove: () => dispatch(setFilter({ key: 'docType', value: '' })),
      })
    }
    if (filters.status.trim() !== '') {
      chips.push({
        key: 'status',
        label: `status:${filters.status.trim()}`,
        onRemove: () => dispatch(setFilter({ key: 'status', value: '' })),
      })
    }
    if (mode === 'reverse' && filters.file.trim() !== '') {
      chips.push({
        key: 'file',
        label: `file:${filters.file.trim()}`,
        onRemove: () => dispatch(setFilter({ key: 'file', value: '' })),
      })
    }
    if (mode === 'reverse' && filters.dir.trim() !== '') {
      chips.push({
        key: 'dir',
        label: `dir:${filters.dir.trim()}`,
        onRemove: () => dispatch(setFilter({ key: 'dir', value: '' })),
      })
    }
    return chips
  }, [dispatch, filters, mode, query])

  const docsLoading = searchDocsState.isFetching
  const filesLoading = searchFilesState.isFetching

  const searchBarValue = mode === 'reverse' ? filters.file : query
  const setSearchBarValue = (v: string) => {
    if (mode === 'reverse') dispatch(setFilter({ key: 'file', value: v }))
    else dispatch(setQuery(v))
  }

  return (
    <div className="search-container container">
      {toast ? (
        <div className="toast-container">
          <div
            className={`alert ${toast.kind === 'success' ? 'alert-success' : 'alert-danger'} mb-0`}
            role="alert"
          >
            {toast.message}
          </div>
        </div>
      ) : null}

      <div className="d-flex justify-content-between align-items-center mb-4">
        <h1 className="h3 mb-0">docmgr Search</h1>
        <button
          className="btn btn-outline-secondary refresh-btn"
          onClick={() => void onRefresh()}
          disabled={refreshState.isLoading}
        >
          {refreshState.isLoading
            ? 'Refreshing…'
            : `Refresh (${timeAgo(wsStatus?.indexedAt)})`}
        </button>
      </div>

      {wsError ? (
        <div className="alert alert-warning">
          Workspace status unavailable. Is the server running on <code>127.0.0.1:3001</code>?
        </div>
      ) : null}

      <form onSubmit={onSubmit}>
        <div className="mb-3">
          <div className="input-group input-group-lg">
            <span className="input-group-text">Search</span>
            <input
              ref={searchInputRef}
              type="text"
              className="form-control search-input"
              placeholder={
                mode === 'reverse'
                  ? 'Search by file path (e.g. backend/api/register.go)'
                  : mode === 'files'
                    ? 'Search for related files…'
                    : 'Search docs…'
              }
              value={searchBarValue}
              onChange={(e) => setSearchBarValue(e.target.value)}
            />
            <button className="btn btn-primary" type="submit">
              Search
            </button>
          </div>
          <div className="keyboard-hint">
            Press <kbd>/</kbd> to focus • <kbd>Ctrl/Cmd</kbd>+<kbd>R</kbd> refresh • <kbd>Esc</kbd>{' '}
            close preview
          </div>
        </div>

        <div className="d-flex gap-2 mb-3">
          <button
            type="button"
            className={`btn btn-sm ${mode === 'docs' ? 'btn-primary' : 'btn-outline-primary'}`}
            onClick={() => dispatch(setMode('docs'))}
          >
            Docs
          </button>
          <button
            type="button"
            className={`btn btn-sm ${mode === 'reverse' ? 'btn-primary' : 'btn-outline-primary'}`}
            onClick={() => dispatch(setMode('reverse'))}
          >
            Reverse Lookup
          </button>
          <button
            type="button"
            className={`btn btn-sm ${mode === 'files' ? 'btn-primary' : 'btn-outline-primary'}`}
            onClick={() => dispatch(setMode('files'))}
          >
            Files
          </button>
          <div className="ms-auto">
            <button
              type="button"
              className="btn btn-sm btn-outline-secondary"
              onClick={() => setShowFilters((v) => !v)}
            >
              {showFilters ? 'Hide filters' : 'Show filters'}
            </button>
          </div>
        </div>

        {showFilters ? (
          <div className="filter-row mb-3">
            <div className="row g-2 align-items-end">
              <div className="col-md-3">
                <label className="form-label small mb-1">Ticket</label>
                <input
                  className="form-control form-control-sm"
                  placeholder="e.g. MEN-4242"
                  value={filters.ticket}
                  onChange={(e) => dispatch(setFilter({ key: 'ticket', value: e.target.value }))}
                />
              </div>
              <div className="col-md-3">
                <label className="form-label small mb-1">Topics</label>
                <TopicMultiSelect
                  topics={filters.topics}
                  onChange={(topics) => dispatch(setFilter({ key: 'topics', value: topics }))}
                />
              </div>
              <div className="col-md-2">
                <label className="form-label small mb-1">Type</label>
                <input
                  className="form-control form-control-sm"
                  placeholder="e.g. reference"
                  value={filters.docType}
                  onChange={(e) => dispatch(setFilter({ key: 'docType', value: e.target.value }))}
                />
              </div>
              <div className="col-md-2">
                <label className="form-label small mb-1">Status</label>
                <select
                  className="form-select form-select-sm"
                  value={filters.status}
                  onChange={(e) => dispatch(setFilter({ key: 'status', value: e.target.value }))}
                >
                  <option value="">All</option>
                  <option value="active">active</option>
                  <option value="review">review</option>
                  <option value="complete">complete</option>
                  <option value="draft">draft</option>
                </select>
              </div>
              <div className="col-md-2">
                <label className="form-label small mb-1">Sort</label>
                <select
                  className="form-select form-select-sm"
                  value={filters.orderBy}
                  onChange={(e) => dispatch(setFilter({ key: 'orderBy', value: e.target.value }))}
                >
                  <option value="rank">Relevance</option>
                  <option value="path">Path</option>
                  <option value="last_updated">Last updated</option>
                </select>
              </div>
            </div>

            {mode === 'reverse' ? (
              <div className="row g-2 mt-2">
                <div className="col-md-6">
                  <label className="form-label small mb-1">File</label>
                  <input
                    className="form-control form-control-sm"
                    placeholder="backend/api/register.go or register.go"
                    value={filters.file}
                    onChange={(e) => dispatch(setFilter({ key: 'file', value: e.target.value }))}
                  />
                </div>
                <div className="col-md-6">
                  <label className="form-label small mb-1">Dir</label>
                  <input
                    className="form-control form-control-sm"
                    placeholder="backend/chat/ws/"
                    value={filters.dir}
                    onChange={(e) => dispatch(setFilter({ key: 'dir', value: e.target.value }))}
                  />
                </div>
              </div>
            ) : null}

            <div className="d-flex flex-wrap gap-3 mt-3 align-items-center">
              <div className="form-check">
                <input
                  className="form-check-input"
                  type="checkbox"
                  checked={filters.includeArchived}
                  onChange={(e) =>
                    dispatch(setFilter({ key: 'includeArchived', value: e.target.checked }))
                  }
                  id="includeArchived"
                />
                <label className="form-check-label" htmlFor="includeArchived">
                  Include archived
                </label>
              </div>
              <div className="form-check">
                <input
                  className="form-check-input"
                  type="checkbox"
                  checked={filters.includeScripts}
                  onChange={(e) =>
                    dispatch(setFilter({ key: 'includeScripts', value: e.target.checked }))
                  }
                  id="includeScripts"
                />
                <label className="form-check-label" htmlFor="includeScripts">
                  Include scripts
                </label>
              </div>
              <div className="form-check">
                <input
                  className="form-check-input"
                  type="checkbox"
                  checked={filters.includeControlDocs}
                  onChange={(e) =>
                    dispatch(setFilter({ key: 'includeControlDocs', value: e.target.checked }))
                  }
                  id="includeControlDocs"
                />
                <label className="form-check-label" htmlFor="includeControlDocs">
                  Control docs
                </label>
              </div>
              <div className="ms-auto">
                <button
                  type="button"
                  className="btn btn-sm btn-outline-secondary"
                  onClick={() => {
                    dispatch(clearFilters())
                    setSelected(null)
                    setHasSearched(false)
                    setDocsResults([])
                    setDocsTotal(0)
                    setDocsNextCursor('')
                    setDocsDiagnostics([])
                  }}
                >
                  Clear
                </button>
              </div>
            </div>
          </div>
        ) : null}
      </form>

      {activeChips.length > 0 ? (
        <div className="mb-3 d-flex flex-wrap gap-2 align-items-center">
          <div className="text-muted small">Active:</div>
          {activeChips.map((c) => (
            <button
              key={c.key}
              type="button"
              className="btn btn-sm btn-outline-secondary"
              onClick={c.onRemove}
            >
              {c.label} ×
            </button>
          ))}
        </div>
      ) : null}

      {mode !== 'files' ? (
        <div className="d-flex align-items-center mb-3">
          <div>
            {hasSearched ? (
              <>
                <strong>{docsTotal}</strong> results
              </>
            ) : (
              <span className="text-muted">No search performed yet</span>
            )}
          </div>
          {docsDiagnostics.length > 0 ? (
            <button
              type="button"
              className="btn btn-sm btn-outline-warning ms-3"
              onClick={() => setShowDiagnostics((v) => !v)}
            >
              {docsDiagnostics.length} diagnostics {showDiagnostics ? '▲' : '▼'}
            </button>
          ) : null}
          <div className="ms-auto">
            {wsStatus ? (
              <span className="text-muted small">
                Indexed {timeAgo(wsStatus.indexedAt)} • {wsStatus.docsIndexed} docs
              </span>
            ) : null}
          </div>
        </div>
      ) : null}

      {mode !== 'files' && showDiagnostics && docsDiagnostics.length > 0 ? (
        <div className="alert alert-warning">
          <div className="fw-semibold mb-2">Diagnostics</div>
          <pre className="mb-0 small">{JSON.stringify(docsDiagnostics, null, 2)}</pre>
        </div>
      ) : null}

      {mode === 'files' ? (
        <>
          {filesLoading ? (
            <div className="loading-spinner">
              <div className="spinner-border text-primary" role="status" />
            </div>
          ) : hasSearched ? (
            <>
              <div className="mb-3">
                <strong>{searchFilesState.data?.total ?? 0}</strong> files
              </div>
              {searchFilesState.data && searchFilesState.data.results.length > 0 ? (
                <div className="vstack gap-2">
                  {searchFilesState.data.results.map((s) => (
                    <div key={`${s.file}:${s.source}:${s.reason}`} className="result-card">
                      <div className="d-flex justify-content-between">
                        <div className="flex-grow-1">
                          <div className="result-title font-monospace">{s.file}</div>
                          <div className="result-meta">
                            {s.source} • <span className="text-muted">{s.reason}</span>
                          </div>
                        </div>
                        <button
                          className="btn btn-sm btn-outline-primary copy-btn ms-2"
                          onClick={() => void onCopyPath(s.file)}
                        >
                          Copy
                        </button>
                      </div>
                    </div>
                  ))}
                </div>
              ) : (
                <div className="empty-state">
                  <h4>No files found</h4>
                  <p>Try adjusting your query or context filters.</p>
                </div>
              )}
            </>
          ) : (
            <div className="empty-state">
              <h4>Find related files</h4>
              <p className="text-muted">Use query + ticket/topics context.</p>
            </div>
          )}
        </>
      ) : docsLoading ? (
        <div className="loading-spinner">
          <div className="spinner-border text-primary" role="status" />
        </div>
      ) : hasSearched ? (
        docsResults.length > 0 ? (
          <>
            <div className={`results-grid ${selected ? 'split' : ''}`}>
              <div>
                {docsResults.map((r) => (
                  <ResultCard
                    key={`${r.path}:${r.ticket}`}
                    result={r}
                    onCopyPath={(p) => void onCopyPath(p)}
                    onSelect={(res) => setSelected(res)}
                  />
                ))}
                {docsNextCursor ? (
                  <div className="text-center mt-3">
                    <button className="btn btn-outline-primary" onClick={() => void onLoadMore()}>
                      Load more
                    </button>
                  </div>
                ) : null}
              </div>

              {selected ? (
                <div className="preview-panel">
                  <div className="d-flex justify-content-between align-items-start mb-2">
                    <div>
                      <div className="h5 mb-1">{selected.title}</div>
                      <div className="text-muted small">
                        {selected.ticket} • {selected.docType} • {selected.status}
                        {selected.lastUpdated ? (
                          <span className="ms-2">Updated {timeAgo(selected.lastUpdated)}</span>
                        ) : null}
                      </div>
                    </div>
                    <button className="btn btn-sm btn-outline-secondary" onClick={() => setSelected(null)}>
                      Close
                    </button>
                  </div>
                  <div className="mb-2">
                    <span className="text-muted small">Path</span>
                    <div className="result-path">{selected.path}</div>
                    <div className="mt-2 d-flex gap-2">
                      <button
                        className="btn btn-sm btn-outline-primary"
                        onClick={() => void onCopyPath(selected.path)}
                      >
                        Copy path
                      </button>
                    </div>
                  </div>
                  <div className="mb-3">
                    <div className="text-muted small mb-1">Snippet</div>
                    <div className="small">{selected.snippet}</div>
                  </div>
                  {selected.relatedFiles && selected.relatedFiles.length > 0 ? (
                    <div>
                      <div className="text-muted small mb-1">Related files</div>
                      <ul className="small mb-0">
                        {selected.relatedFiles.map((rf) => (
                          <li key={`${rf.path}:${rf.note ?? ''}`}>
                            <span className="font-monospace">{rf.path}</span>
                            {rf.note ? <span className="text-muted ms-2">{rf.note}</span> : null}
                          </li>
                        ))}
                      </ul>
                    </div>
                  ) : null}
                </div>
              ) : null}
            </div>
          </>
        ) : (
          <div className="empty-state">
            <h4>No results found</h4>
            <p>Try adjusting your query or filters.</p>
          </div>
        )
      ) : (
        <div className="empty-state">
          <h4>Search docmgr documentation</h4>
          <p className="text-muted">Enter a query or use filters to browse documentation.</p>
        </div>
      )}

      {searchDocsState.isError || searchFilesState.isError ? (
        <div className="alert alert-danger mt-3">
          Search error. Ensure the backend is running and the index is built.
        </div>
      ) : null}
    </div>
  )
}
