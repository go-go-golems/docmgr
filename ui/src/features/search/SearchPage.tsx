import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import type { FormEvent } from 'react'
import { Link, useNavigate } from 'react-router-dom'

import { useAppDispatch, useAppSelector } from '../../app/hooks'
import { DocCard } from '../../components/DocCard'
import { DiagnosticList } from './components/DiagnosticList'
import { MarkdownSnippet } from './components/MarkdownSnippet'
import { TopicMultiSelect } from './components/TopicMultiSelect'
import { useIsMobile } from './hooks/useIsMobile'
import { clearFilters, setFilter, setMode, setQuery } from './searchSlice'
import {
  useGetWorkspaceStatusQuery,
  useLazySearchDocsQuery,
  useLazySearchFilesQuery,
  useRefreshIndexMutation,
} from '../../services/docmgrApi'
import type { DiagnosticTaxonomy, SearchDocResult } from '../../services/docmgrApi'

type ToastState = { kind: 'success' | 'error'; message: string } | null

type APIErrorPayload = {
  error?: {
    code?: string
    message?: string
    details?: unknown
  }
}

type ErrorBanner = {
  title: string
  code?: string
  message: string
  details?: unknown
}

function isEditableTarget(target: EventTarget | null): boolean {
  const el = target as HTMLElement | null
  if (!el) return false
  const tag = el.tagName.toLowerCase()
  if (tag === 'input' || tag === 'textarea' || tag === 'select') return true
  if (el.isContentEditable) return true
  return false
}

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

function parseBoolParam(v: string | null, def: boolean): boolean {
  if (v == null || v.trim() === '') return def
  const s = v.trim().toLowerCase()
  if (['1', 'true', 't', 'yes', 'y', 'on'].includes(s)) return true
  if (['0', 'false', 'f', 'no', 'n', 'off'].includes(s)) return false
  return def
}

function parseCSV(v: string | null): string[] {
  if (v == null) return []
  const s = v.trim()
  if (s === '') return []
  return s
    .split(',')
    .map((p) => p.trim())
    .filter((p) => p !== '')
}

function formatCSV(values: string[]): string {
  return values.map((v) => v.trim()).filter((v) => v !== '').join(',')
}

function toErrorBanner(err: unknown, title: string): ErrorBanner {
  const maybe = err as { data?: unknown; status?: number } | undefined
  const data = maybe?.data as APIErrorPayload | undefined
  const code = data?.error?.code
  const message =
    data?.error?.message ??
    (typeof err === 'string' ? err : err instanceof Error ? err.message : String(err))
  const details = data?.error?.details
  return { title, code, message, details }
}

export function SearchPage() {
  const navigate = useNavigate()
  const dispatch = useAppDispatch()
  const { mode, query, filters } = useAppSelector((s) => s.search)
  const searchInputRef = useRef<HTMLInputElement | null>(null)
  const isMobile = useIsMobile(992)

  const { data: wsStatus, isError: wsError, refetch: refetchWs } = useGetWorkspaceStatusQuery()
  const [refreshIndex, refreshState] = useRefreshIndexMutation()

  const [triggerSearchDocs, searchDocsState] = useLazySearchDocsQuery()
  const [triggerSearchFiles, searchFilesState] = useLazySearchFilesQuery()

  const [toast, setToast] = useState<ToastState>(null)
  const [errorBanner, setErrorBanner] = useState<ErrorBanner | null>(null)
  const [showFilters, setShowFilters] = useState(true)
  const [showDiagnostics, setShowDiagnostics] = useState(false)
  const [urlSyncReady, setURLSyncReady] = useState(false)
  const [showShortcuts, setShowShortcuts] = useState(false)
  const [showFilterDrawer, setShowFilterDrawer] = useState(false)
  const [showPreviewModal, setShowPreviewModal] = useState(false)

  const [hasSearched, setHasSearched] = useState(false)

  const [docsResults, setDocsResults] = useState<SearchDocResult[]>([])
  const [docsTotal, setDocsTotal] = useState<number>(0)
  const [docsNextCursor, setDocsNextCursor] = useState<string>('')
  const [docsDiagnostics, setDocsDiagnostics] = useState<DiagnosticTaxonomy[]>([])

  const [selected, setSelected] = useState<SearchDocResult | null>(null)
  const [selectedIndex, setSelectedIndex] = useState<number | null>(null)

  const onCopyPath = useCallback(
    async (path: string) => {
      try {
        await navigator.clipboard.writeText(path)
        setToast({ kind: 'success', message: `Copied path: ${path}` })
      } catch {
        setToast({ kind: 'error', message: 'Failed to copy path (clipboard not available)' })
      }
    },
    [setToast],
  )

  const onRefresh = useCallback(async () => {
    try {
      await refreshIndex().unwrap()
      await refetchWs()
      setToast({ kind: 'success', message: 'Index refreshed successfully' })
    } catch (e) {
      setToast({ kind: 'error', message: `Index refresh failed: ${String(e)}` })
    }
  }, [refreshIndex, refetchWs])

  const selectDocByIndex = useCallback(
    (idx: number, opts?: { openModal?: boolean }) => {
      if (idx < 0 || idx >= docsResults.length) return
      setSelected(docsResults[idx])
      setSelectedIndex(idx)
      if (isMobile && (opts?.openModal ?? true)) {
        setShowPreviewModal(true)
      }
    },
    [docsResults, isMobile],
  )

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
      if (e.key === 'Escape') {
        if (showShortcuts) {
          e.preventDefault()
          setShowShortcuts(false)
          return
        }
        if (showPreviewModal) {
          e.preventDefault()
          setShowPreviewModal(false)
          return
        }
      }

      if (e.key === '/' && document.activeElement !== searchInputRef.current && !isEditableTarget(e.target)) {
        e.preventDefault()
        searchInputRef.current?.focus()
      }
      if (e.key === '?' && !e.ctrlKey && !e.metaKey && !e.altKey && !isEditableTarget(e.target)) {
        e.preventDefault()
        setShowShortcuts(true)
      }

      if (e.altKey && !e.ctrlKey && !e.metaKey && !isEditableTarget(e.target)) {
        if (e.key === '1') {
          e.preventDefault()
          dispatch(setMode('docs'))
          setSelected(null)
          setSelectedIndex(null)
          setShowPreviewModal(false)
          searchInputRef.current?.focus()
          return
        }
        if (e.key === '2') {
          e.preventDefault()
          dispatch(setMode('reverse'))
          setSelected(null)
          setSelectedIndex(null)
          setShowPreviewModal(false)
          searchInputRef.current?.focus()
          return
        }
        if (e.key === '3') {
          e.preventDefault()
          dispatch(setMode('files'))
          setSelected(null)
          setSelectedIndex(null)
          setShowPreviewModal(false)
          searchInputRef.current?.focus()
          return
        }
      }

      if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 'k' && !isEditableTarget(e.target)) {
        if (!selected) return
        e.preventDefault()
        void onCopyPath(selected.path)
        return
      }

      if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 'r') {
        e.preventDefault()
        void onRefresh()
        return
      }

      if (mode !== 'files' && docsResults.length > 0 && !isEditableTarget(e.target)) {
        if (e.key === 'ArrowDown') {
          e.preventDefault()
          const idx =
            selectedIndex != null
              ? selectedIndex
              : selected
                ? docsResults.findIndex((d) => d.path === selected.path && d.ticket === selected.ticket)
                : -1
          const next = Math.min(docsResults.length - 1, idx + 1)
          selectDocByIndex(next)
          return
        }
        if (e.key === 'ArrowUp') {
          e.preventDefault()
          const idx =
            selectedIndex != null
              ? selectedIndex
              : selected
                ? docsResults.findIndex((d) => d.path === selected.path && d.ticket === selected.ticket)
                : docsResults.length
          const prev = Math.max(0, idx - 1)
          selectDocByIndex(prev)
          return
        }
        if (e.key === 'Enter') {
          if (!selected) {
            e.preventDefault()
            selectDocByIndex(0)
            return
          }
          e.preventDefault()
          navigate(`/doc?path=${encodeURIComponent(selected.path)}`)
          return
        }
        if (e.key === 'Escape') {
          if (selected) {
            e.preventDefault()
            setSelected(null)
            setSelectedIndex(null)
            setShowPreviewModal(false)
            return
          }
        }
      }
    }
    window.addEventListener('keydown', onKeyDown)
    return () => window.removeEventListener('keydown', onKeyDown)
  }, [
    dispatch,
    docsResults,
    isMobile,
    mode,
    navigate,
    onCopyPath,
    onRefresh,
    selectDocByIndex,
    selected,
    selectedIndex,
    showPreviewModal,
    showShortcuts,
  ])

  useEffect(() => {
    // If we exit mobile sizing, close the modal and keep the selection in the desktop preview panel.
    if (!isMobile) {
      setShowPreviewModal(false)
      return
    }
  }, [isMobile])

  useEffect(() => {
    if (!toast) return
    const t = window.setTimeout(() => setToast(null), 2000)
    return () => window.clearTimeout(t)
  }, [toast])

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
    setDocsDiagnostics(resp.diagnostics ?? [])
    setHasSearched(true)

    if (append) setDocsResults((prev) => [...prev, ...resp.results])
    else setDocsResults(resp.results)
  }

  const onSubmit = async (e: FormEvent) => {
    e.preventDefault()
    setSelected(null)
    setShowDiagnostics(false)
    setErrorBanner(null)

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
        setErrorBanner(toErrorBanner(err, 'Files search failed'))
      }
      return
    }

    try {
      await doSearchDocs('', false)
    } catch (err) {
      setErrorBanner(toErrorBanner(err, 'Search failed'))
    }
  }

  const onLoadMore = async () => {
    if (!docsNextCursor) return
    try {
      await doSearchDocs(docsNextCursor, true)
    } catch (err) {
      setErrorBanner(toErrorBanner(err, 'Load more failed'))
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

  const [desiredSelectedPath, setDesiredSelectedPath] = useState<string>('')
  const [desiredPreviewOpen, setDesiredPreviewOpen] = useState<boolean>(false)

  // URL sync (mode/q/filters) – no cursor/pagination state in URL.
  const urlWriteTimerRef = useRef<number | null>(null)
  const autoSearchedRef = useRef(false)
  const selectionAppliedRef = useRef(false)

  useEffect(() => {
    const params = new URLSearchParams(window.location.search)
    const modeParam = (params.get('mode') || '').trim()
    const nextMode = modeParam === 'reverse' || modeParam === 'files' ? modeParam : 'docs'

    dispatch(setMode(nextMode))
    dispatch(setQuery((params.get('q') || '').trim()))
    dispatch(setFilter({ key: 'ticket', value: (params.get('ticket') || '').trim() }))
    dispatch(setFilter({ key: 'topics', value: parseCSV(params.get('topics')) }))
    dispatch(setFilter({ key: 'docType', value: (params.get('docType') || '').trim() }))
    dispatch(setFilter({ key: 'status', value: (params.get('status') || '').trim() }))
    dispatch(setFilter({ key: 'file', value: (params.get('file') || '').trim() }))
    dispatch(setFilter({ key: 'dir', value: (params.get('dir') || '').trim() }))
    dispatch(setFilter({ key: 'orderBy', value: (params.get('orderBy') || '').trim() || 'rank' }))
    dispatch(
      setFilter({
        key: 'includeArchived',
        value: parseBoolParam(params.get('includeArchived'), true),
      }),
    )
    dispatch(
      setFilter({
        key: 'includeScripts',
        value: parseBoolParam(params.get('includeScripts'), true),
      }),
    )
    dispatch(
      setFilter({
        key: 'includeControlDocs',
        value: parseBoolParam(params.get('includeControlDocs'), true),
      }),
    )

    setDesiredSelectedPath((params.get('sel') || '').trim())
    setDesiredPreviewOpen(parseBoolParam(params.get('preview'), false))

    setURLSyncReady(true)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  useEffect(() => {
    if (!urlSyncReady) return

    if (urlWriteTimerRef.current != null) window.clearTimeout(urlWriteTimerRef.current)
    urlWriteTimerRef.current = window.setTimeout(() => {
      const params = new URLSearchParams()
      if (mode !== 'docs') params.set('mode', mode)

      if (query.trim() !== '') params.set('q', query.trim())
      if (filters.ticket.trim() !== '') params.set('ticket', filters.ticket.trim())
      if (filters.topics.length > 0) params.set('topics', formatCSV(filters.topics))
      if (filters.docType.trim() !== '') params.set('docType', filters.docType.trim())
      if (filters.status.trim() !== '') params.set('status', filters.status.trim())
      if (filters.file.trim() !== '') params.set('file', filters.file.trim())
      if (filters.dir.trim() !== '') params.set('dir', filters.dir.trim())
      if (filters.orderBy.trim() !== '' && filters.orderBy.trim() !== 'rank')
        params.set('orderBy', filters.orderBy.trim())
      if (filters.includeArchived !== true) params.set('includeArchived', String(filters.includeArchived))
      if (filters.includeScripts !== true) params.set('includeScripts', String(filters.includeScripts))
      if (filters.includeControlDocs !== true)
        params.set('includeControlDocs', String(filters.includeControlDocs))

      if (selected && selected.path.trim() !== '') params.set('sel', selected.path.trim())
      if (isMobile && showPreviewModal) params.set('preview', 'true')

      const next = params.toString()
      const url = next === '' ? window.location.pathname : `${window.location.pathname}?${next}`
      window.history.replaceState({}, '', url)
    }, 250)

    return () => {
      if (urlWriteTimerRef.current != null) window.clearTimeout(urlWriteTimerRef.current)
    }
  }, [filters, isMobile, mode, query, selected, showPreviewModal, urlSyncReady])

  useEffect(() => {
    if (!urlSyncReady) return
    if (selectionAppliedRef.current) return
    if (desiredSelectedPath.trim() === '') return
    if (docsResults.length === 0) return

    const idx = docsResults.findIndex((d) => d.path === desiredSelectedPath.trim())
    if (idx < 0) {
      selectionAppliedRef.current = true
      return
    }

    selectionAppliedRef.current = true
    setSelected(docsResults[idx])
    setSelectedIndex(idx)
    if (isMobile && desiredPreviewOpen) setShowPreviewModal(true)
  }, [desiredPreviewOpen, desiredSelectedPath, docsResults, isMobile, urlSyncReady])

  useEffect(() => {
    if (!urlSyncReady) return
    if (autoSearchedRef.current) return

    const hasIntent =
      query.trim() !== '' ||
      filters.ticket.trim() !== '' ||
      filters.topics.length > 0 ||
      filters.docType.trim() !== '' ||
      filters.status.trim() !== '' ||
      filters.file.trim() !== '' ||
      filters.dir.trim() !== ''

    if (!hasIntent) return

    autoSearchedRef.current = true
    // Auto-run search once after URL restore so shared links "just work".
    void (async () => {
      try {
        setErrorBanner(null)
        setSelected(null)
        setShowDiagnostics(false)
        if (mode === 'files') {
          setHasSearched(true)
          await triggerSearchFiles({
            query,
            ticket: filters.ticket,
            topics: filters.topics,
            limit: 200,
          }).unwrap()
          return
        }
        await doSearchDocs('', false)
      } catch (err) {
        setErrorBanner(toErrorBanner(err, 'Search failed'))
      }
    })()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [urlSyncReady])

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

      {errorBanner ? (
        <div className="alert alert-danger">
          <div className="fw-semibold">{errorBanner.title}</div>
          <div className="small">
            {errorBanner.code ? <span className="me-2">({errorBanner.code})</span> : null}
            {errorBanner.message}
          </div>
          {errorBanner.details ? (
            <details className="mt-2">
              <summary className="small">Details</summary>
              <pre className="small mb-0">{JSON.stringify(errorBanner.details, null, 2)}</pre>
            </details>
          ) : null}
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
            Press <kbd>/</kbd> focus • <kbd>?</kbd> shortcuts • <kbd>Ctrl/Cmd</kbd>+<kbd>R</kbd> refresh •{' '}
            <kbd>Esc</kbd> close
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
            {isMobile ? (
              <button
                type="button"
                className="btn btn-sm btn-outline-secondary"
                onClick={() => setShowFilterDrawer(true)}
              >
                Filters
              </button>
            ) : (
              <button
                type="button"
                className="btn btn-sm btn-outline-secondary"
                onClick={() => setShowFilters((v) => !v)}
              >
                {showFilters ? 'Hide filters' : 'Show filters'}
              </button>
            )}
          </div>
        </div>

        {!isMobile && showFilters ? (
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

      {isMobile && showFilterDrawer ? (
        <>
          <div className="modal-backdrop show" />
          <div className="modal show d-block" tabIndex={-1} role="dialog" aria-modal="true">
            <div className="modal-dialog modal-fullscreen-sm-down modal-dialog-scrollable" role="document">
              <div className="modal-content">
                <div className="modal-header">
                  <h5 className="modal-title">Filters</h5>
                  <button type="button" className="btn-close" onClick={() => setShowFilterDrawer(false)} />
                </div>
                <div className="modal-body">
                  <div className="filter-row mb-0 border-0 p-0">
                    <div className="row g-2 align-items-end">
                      <div className="col-12">
                        <label className="form-label small mb-1">Ticket</label>
                        <input
                          className="form-control form-control-sm"
                          placeholder="e.g. MEN-4242"
                          value={filters.ticket}
                          onChange={(e) => dispatch(setFilter({ key: 'ticket', value: e.target.value }))}
                        />
                      </div>
                      <div className="col-12">
                        <label className="form-label small mb-1">Topics</label>
                        <TopicMultiSelect
                          topics={filters.topics}
                          onChange={(topics) => dispatch(setFilter({ key: 'topics', value: topics }))}
                        />
                      </div>
                      <div className="col-12">
                        <label className="form-label small mb-1">Type</label>
                        <input
                          className="form-control form-control-sm"
                          placeholder="e.g. reference"
                          value={filters.docType}
                          onChange={(e) => dispatch(setFilter({ key: 'docType', value: e.target.value }))}
                        />
                      </div>
                      <div className="col-12">
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
                      <div className="col-12">
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
                        <div className="col-12">
                          <label className="form-label small mb-1">File</label>
                          <input
                            className="form-control form-control-sm"
                            placeholder="backend/api/register.go or register.go"
                            value={filters.file}
                            onChange={(e) => dispatch(setFilter({ key: 'file', value: e.target.value }))}
                          />
                        </div>
                        <div className="col-12">
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
                          id="includeArchivedMobile"
                        />
                        <label className="form-check-label" htmlFor="includeArchivedMobile">
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
                          id="includeScriptsMobile"
                        />
                        <label className="form-check-label" htmlFor="includeScriptsMobile">
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
                          id="includeControlDocsMobile"
                        />
                        <label className="form-check-label" htmlFor="includeControlDocsMobile">
                          Control docs
                        </label>
                      </div>
                    </div>
                  </div>
                </div>
                <div className="modal-footer">
                  <button
                    type="button"
                    className="btn btn-outline-secondary"
                    onClick={() => {
                      dispatch(clearFilters())
                      setSelected(null)
                      setSelectedIndex(null)
                      setHasSearched(false)
                      setDocsResults([])
                      setDocsTotal(0)
                      setDocsNextCursor('')
                      setDocsDiagnostics([])
                      setShowFilterDrawer(false)
                    }}
                  >
                    Clear
                  </button>
                  <button type="button" className="btn btn-primary" onClick={() => setShowFilterDrawer(false)}>
                    Done
                  </button>
                </div>
              </div>
            </div>
          </div>
        </>
      ) : null}

      {showShortcuts ? (
        <>
          <div className="modal-backdrop show" />
          <div className="modal show d-block" tabIndex={-1} role="dialog" aria-modal="true">
            <div className="modal-dialog" role="document">
              <div className="modal-content">
                <div className="modal-header">
                  <h5 className="modal-title">Keyboard shortcuts</h5>
                  <button type="button" className="btn-close" onClick={() => setShowShortcuts(false)} />
                </div>
                <div className="modal-body">
                  <ul className="mb-0">
                    <li>
                      <kbd>/</kbd> focus search
                    </li>
                    <li>
                      <kbd>↑</kbd>/<kbd>↓</kbd> select result
                    </li>
                    <li>
                      <kbd>Enter</kbd> open selected doc
                    </li>
                    <li>
                      <kbd>Esc</kbd> close modal/preview
                    </li>
                    <li>
                      <kbd>Alt</kbd>+<kbd>1</kbd>/<kbd>2</kbd>/<kbd>3</kbd> switch modes
                    </li>
                    <li>
                      <kbd>Ctrl/Cmd</kbd>+<kbd>R</kbd> refresh index
                    </li>
                    <li>
                      <kbd>Ctrl/Cmd</kbd>+<kbd>K</kbd> copy selected doc path
                    </li>
                  </ul>
                </div>
                <div className="modal-footer">
                  <button type="button" className="btn btn-primary" onClick={() => setShowShortcuts(false)}>
                    Close
                  </button>
                </div>
              </div>
            </div>
          </div>
        </>
      ) : null}

      {isMobile && selected && showPreviewModal ? (
        <>
          <div className="modal-backdrop show" />
          <div className="modal show d-block" tabIndex={-1} role="dialog" aria-modal="true">
            <div className="modal-dialog modal-fullscreen-sm-down modal-dialog-scrollable" role="document">
              <div className="modal-content">
                <div className="modal-header">
                  <h5 className="modal-title">{selected.title}</h5>
                  <button type="button" className="btn-close" onClick={() => setShowPreviewModal(false)} />
                </div>
                <div className="modal-body">
                  <div className="text-muted small mb-2">
                    <Link to={`/ticket/${encodeURIComponent(selected.ticket)}`} className="text-decoration-none">
                      {selected.ticket}
                    </Link>{' '}
                    • {selected.docType} • {selected.status}
                    {selected.lastUpdated ? (
                      <span className="ms-2">Updated {timeAgo(selected.lastUpdated)}</span>
                    ) : null}
                  </div>

                  <div className="mb-2">
                    <span className="text-muted small">Path</span>
                    <div className="result-path">{selected.path}</div>
	                  <div className="mt-2 d-flex gap-2 flex-wrap">
                      <button className="btn btn-sm btn-outline-primary" onClick={() => void onCopyPath(selected.path)}>
                        Copy path
                      </button>
                      <Link className="btn btn-sm btn-primary" to={`/doc?path=${encodeURIComponent(selected.path)}`}>
                        Open doc
                      </Link>
	                  </div>
	                </div>

	                <div className="mb-3">
	                  <div className="text-muted small mb-1">Snippet</div>
                    <div className="small">
                      <MarkdownSnippet markdown={selected.snippet} query={mode === 'docs' ? query : ''} />
                    </div>
	                </div>

                  {selected.relatedFiles && selected.relatedFiles.length > 0 ? (
                    <div>
                      <div className="text-muted small mb-1">Related files</div>
                      <ul className="small mb-0 list-unstyled vstack gap-2">
                        {selected.relatedFiles.map((rf) => (
                          <li key={`${rf.path}:${rf.note ?? ''}`} className="d-flex gap-2 align-items-start">
                            <div className="flex-grow-1">
                              <div className="font-monospace">{rf.path}</div>
                              {rf.note ? <div className="text-muted">{rf.note}</div> : null}
                            </div>
                            <div className="d-flex gap-2">
                              <button
                                className="btn btn-sm btn-outline-secondary"
                                onClick={() => void onCopyPath(rf.path)}
                              >
                                Copy
                              </button>
                              <Link className="btn btn-sm btn-outline-primary" to={`/file?root=repo&path=${encodeURIComponent(rf.path)}`}>
                                Open
                              </Link>
	                          </div>
	                        </li>
	                      ))}
	                    </ul>
                    </div>
                  ) : null}
                </div>
                <div className="modal-footer">
                  <button
                    type="button"
                    className="btn btn-outline-secondary"
                    onClick={() => {
                      setShowPreviewModal(false)
                      setSelected(null)
                      setSelectedIndex(null)
                    }}
                  >
                    Close preview
                  </button>
                </div>
              </div>
            </div>
          </div>
        </>
      ) : null}

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
          <DiagnosticList diagnostics={docsDiagnostics} />
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
                  <DocCard
                    key={`${r.path}:${r.ticket}`}
                    title={r.title}
                    ticket={r.ticket}
                    docType={r.docType}
                    status={r.status}
                    topics={r.topics}
                    path={r.path}
                    lastUpdated={r.lastUpdated}
                    relatedFiles={r.relatedFiles}
                    snippet={<MarkdownSnippet markdown={r.snippet} query={mode === 'docs' ? query : ''} />}
                    selected={selected?.path === r.path && selected?.ticket === r.ticket}
                    onCopyPath={(p) => void onCopyPath(p)}
                    onSelect={() => {
                      const idx = docsResults.findIndex(
                        (d) => d.path === r.path && d.ticket === r.ticket,
                      )
                      setSelected(r)
                      setSelectedIndex(idx >= 0 ? idx : null)
                      if (isMobile) setShowPreviewModal(true)
                    }}
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

              {selected && !isMobile ? (
                <div className="preview-panel">
                  <div className="d-flex justify-content-between align-items-start mb-2">
                    <div>
                      <div className="h5 mb-1">{selected.title}</div>
                      <div className="text-muted small">
                        <Link to={`/ticket/${encodeURIComponent(selected.ticket)}`} className="text-decoration-none">
                          {selected.ticket}
                        </Link>{' '}
                        • {selected.docType} • {selected.status}
                        {selected.lastUpdated ? (
                          <span className="ms-2">Updated {timeAgo(selected.lastUpdated)}</span>
                        ) : null}
                      </div>
                    </div>
                    <button
                      className="btn btn-sm btn-outline-secondary"
                      onClick={() => {
                        setSelected(null)
                        setSelectedIndex(null)
                      }}
                    >
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
	                    <Link
	                      className="btn btn-sm btn-primary"
	                      to={`/doc?path=${encodeURIComponent(selected.path)}`}
	                    >
	                      Open doc
	                    </Link>
	                  </div>
	                </div>
	                <div className="mb-3">
	                  <div className="text-muted small mb-1">Snippet</div>
	                  <div className="small">
	                    <MarkdownSnippet markdown={selected.snippet} query={mode === 'docs' ? query : ''} />
	                  </div>
	                </div>
                  {selected.relatedFiles && selected.relatedFiles.length > 0 ? (
                    <div>
                      <div className="text-muted small mb-1">Related files</div>
                      <ul className="small mb-0 list-unstyled vstack gap-2">
                        {selected.relatedFiles.map((rf) => (
                          <li key={`${rf.path}:${rf.note ?? ''}`} className="d-flex gap-2 align-items-start">
                            <div className="flex-grow-1">
                              <div className="font-monospace">{rf.path}</div>
                              {rf.note ? <div className="text-muted">{rf.note}</div> : null}
                            </div>
                            <div className="d-flex gap-2">
                              <button
                                className="btn btn-sm btn-outline-secondary"
                                onClick={() => void onCopyPath(rf.path)}
                              >
                                Copy
                              </button>
	                              <Link
	                                className="btn btn-sm btn-outline-primary"
	                                to={`/file?root=repo&path=${encodeURIComponent(rf.path)}`}
	                              >
	                                Open
	                              </Link>
	                            </div>
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

      {/* Errors are rendered via the main error banner near the top. */}
    </div>
  )
}
