import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import type { FormEvent } from 'react'
import { useNavigate } from 'react-router-dom'

import { useAppDispatch, useAppSelector } from '../../app/hooks'
import { ApiErrorAlert } from '../../components/ApiErrorAlert'
import { copyToClipboard } from '../../lib/clipboard'
import { useIsMobile } from './hooks/useIsMobile'
import { useSearchSelection } from './hooks/useSearchSelection'
import { useSearchUrlSync } from './hooks/useSearchUrlSync'
import type { SearchFilters } from './searchSlice'
import { clearFilters, setFilter, setMode, setQuery } from './searchSlice'
import { SearchActiveChips } from './widgets/SearchActiveChips'
import { SearchBar } from './widgets/SearchBar'
import { SearchDiagnosticsPanel } from './widgets/SearchDiagnosticsPanel'
import { SearchDocsResults } from './widgets/SearchDocsResults'
import { SearchFilesResults } from './widgets/SearchFilesResults'
import { SearchFiltersDesktop } from './widgets/SearchFiltersDesktop'
import { SearchFiltersDrawer } from './widgets/SearchFiltersDrawer'
import { SearchHeader } from './widgets/SearchHeader'
import { SearchModeToggle } from './widgets/SearchModeToggle'
import { SearchPreviewModal } from './widgets/SearchPreviewModal'
import { SearchPreviewPanel } from './widgets/SearchPreviewPanel'
import {
  useGetWorkspaceStatusQuery,
  useLazySearchDocsQuery,
  useLazySearchFilesQuery,
  useRefreshIndexMutation,
} from '../../services/docmgrApi'

type ToastState = { kind: 'success' | 'error'; message: string } | null
type ErrorState = { title: string; error: unknown } | null

function isEditableTarget(target: EventTarget | null): boolean {
  const el = target as HTMLElement | null
  if (!el) return false
  const tag = el.tagName.toLowerCase()
  if (tag === 'input' || tag === 'textarea' || tag === 'select') return true
  if (el.isContentEditable) return true
  return false
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
  const [errorState, setErrorState] = useState<ErrorState>(null)
  const [showFilters, setShowFilters] = useState(true)
  const [showDiagnostics, setShowDiagnostics] = useState(false)
  const [showShortcuts, setShowShortcuts] = useState(false)
  const [showFilterDrawer, setShowFilterDrawer] = useState(false)
  const [showPreviewModal, setShowPreviewModal] = useState(false)

  const [selectedPathForUrl, setSelectedPathForUrl] = useState<string>('')

  const { urlSyncReady, desiredSelectedPath, desiredPreviewOpen } = useSearchUrlSync({
    dispatch,
    mode,
    query,
    filters,
    selectedPath: selectedPathForUrl,
    previewOpen: isMobile && showPreviewModal,
  })

  const docsData = searchDocsState.data
  const docsResults = useMemo(() => docsData?.results ?? [], [docsData?.results])
  const docsTotal = docsData?.total ?? 0
  const docsNextCursor = docsData?.nextCursor ?? ''
  const docsDiagnostics = useMemo(() => docsData?.diagnostics ?? [], [docsData?.diagnostics])
  const docsHighlightQuery = mode === 'docs' ? (docsData?.query?.query ?? query) : ''

  const hasSearched = !searchDocsState.isUninitialized || !searchFilesState.isUninitialized

  const { selected, selectedIndex, setSelected, setSelectedIndex, clearSelection, selectDocByIndex } =
    useSearchSelection({
      docsResults,
      isMobile,
      urlSyncReady,
      desiredSelectedPath,
      desiredPreviewOpen,
      openPreviewModal: () => setShowPreviewModal(true),
      closePreviewModal: () => setShowPreviewModal(false),
    })

  useEffect(() => {
    setSelectedPathForUrl(selected?.path ?? '')
  }, [selected])

  const onCopyPath = useCallback(
    async (path: string) => {
      try {
        await copyToClipboard(path)
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
          clearSelection()
          searchInputRef.current?.focus()
          return
        }
        if (e.key === '2') {
          e.preventDefault()
          dispatch(setMode('reverse'))
          clearSelection()
          searchInputRef.current?.focus()
          return
        }
        if (e.key === '3') {
          e.preventDefault()
          dispatch(setMode('files'))
          clearSelection()
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
            clearSelection()
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
    clearSelection,
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

  const doSearchDocs = async (cursor: string) => {
    const textQuery = mode === 'reverse' ? '' : query
    await triggerSearchDocs(
      {
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
      },
      false,
    ).unwrap()
  }

  const onSubmit = async (e: FormEvent) => {
    e.preventDefault()
    setSelected(null)
    setShowDiagnostics(false)
    setErrorState(null)

    if (mode === 'files') {
      try {
        await triggerSearchFiles(
          {
            query,
            ticket: filters.ticket,
            topics: filters.topics,
            limit: 200,
          },
          false,
        ).unwrap()
      } catch (err) {
        setErrorState({ title: 'Files search failed', error: err })
      }
      return
    }

    try {
      await doSearchDocs('')
    } catch (err) {
      setErrorState({ title: 'Search failed', error: err })
    }
  }

  const onLoadMore = async () => {
    if (!docsNextCursor) return
    try {
      await doSearchDocs(docsNextCursor)
    } catch (err) {
      setErrorState({ title: 'Load more failed', error: err })
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

  const autoSearchedRef = useRef(false)

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
        setErrorState(null)
        setSelected(null)
        setShowDiagnostics(false)
        if (mode === 'files') {
          await triggerSearchFiles(
            {
              query,
              ticket: filters.ticket,
              topics: filters.topics,
              limit: 200,
            },
            false,
          ).unwrap()
          return
        }
        await doSearchDocs('')
      } catch (err) {
        setErrorState({ title: 'Search failed', error: err })
      }
    })()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [urlSyncReady])

  const onFilterChange = <K extends keyof SearchFilters,>(key: K, value: SearchFilters[K]) => {
    dispatch(setFilter({ key, value } as never))
  }

  const clearAll = (opts?: { closeFilterDrawer?: boolean }) => {
    dispatch(clearFilters())
    clearSelection()
    setErrorState(null)
    setShowDiagnostics(false)
    setShowPreviewModal(false)
    searchDocsState.reset?.()
    searchFilesState.reset?.()
    autoSearchedRef.current = false
    if (opts?.closeFilterDrawer) setShowFilterDrawer(false)
  }

  const filesTotal = searchFilesState.data?.total ?? 0
  const filesResults = searchFilesState.data?.results ?? []
  const showDocsGrid = hasSearched && docsResults.length > 0 && !docsLoading

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

      <SearchHeader
        wsStatus={wsStatus}
        wsError={wsError}
        refreshLoading={refreshState.isLoading}
        onRefresh={() => void onRefresh()}
      />

      {errorState ? <ApiErrorAlert title={errorState.title} error={errorState.error} /> : null}

      <form onSubmit={onSubmit}>
        <SearchBar
          mode={mode}
          value={searchBarValue}
          onChange={setSearchBarValue}
          inputRef={searchInputRef}
        />

        <SearchModeToggle
          mode={mode}
          onModeChange={(m) => dispatch(setMode(m))}
          isMobile={isMobile}
          showFilters={showFilters}
          onToggleFilters={() => setShowFilters((v) => !v)}
          onOpenFilterDrawer={() => setShowFilterDrawer(true)}
        />

        {!isMobile && showFilters ? (
          <SearchFiltersDesktop
            mode={mode}
            filters={filters}
            onFilterChange={onFilterChange}
            onClear={() => clearAll()}
          />
        ) : null}
      </form>

      <SearchFiltersDrawer
        open={isMobile && showFilterDrawer}
        mode={mode}
        filters={filters}
        onFilterChange={onFilterChange}
        onClose={() => setShowFilterDrawer(false)}
        onClear={() => clearAll({ closeFilterDrawer: true })}
      />

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

      <SearchPreviewModal
        open={isMobile && showPreviewModal}
        selected={selected}
        highlightQuery={docsHighlightQuery}
        onCopyPath={(p) => void onCopyPath(p)}
        onDismiss={() => setShowPreviewModal(false)}
        onClosePreview={() => {
          setShowPreviewModal(false)
          setSelected(null)
          setSelectedIndex(null)
        }}
      />

      <SearchActiveChips chips={activeChips} />

      {mode !== 'files' ? (
        <SearchDiagnosticsPanel
          hasSearched={hasSearched}
          docsTotal={docsTotal}
          diagnostics={docsDiagnostics}
          showDiagnostics={showDiagnostics}
          onToggleDiagnostics={() => setShowDiagnostics((v) => !v)}
          wsStatus={wsStatus}
        />
      ) : null}

      {mode === 'files' ? (
        <SearchFilesResults
          loading={filesLoading}
          hasSearched={hasSearched}
          total={filesTotal}
          results={filesResults}
          onCopyPath={(p) => void onCopyPath(p)}
        />
      ) : showDocsGrid ? (
        <div className={`results-grid ${selected ? 'split' : ''}`}>
          <div>
            <SearchDocsResults
              loading={docsLoading}
              hasSearched={hasSearched}
              docsResults={docsResults}
              selected={selected}
              onSelectIndex={(idx) => selectDocByIndex(idx)}
              onCopyPath={(p) => void onCopyPath(p)}
              highlightQuery={docsHighlightQuery}
              hasMore={docsNextCursor !== ''}
              onLoadMore={() => void onLoadMore()}
            />
          </div>
          {selected && !isMobile ? (
            <SearchPreviewPanel
              selected={selected}
              highlightQuery={docsHighlightQuery}
              onCopyPath={(p) => void onCopyPath(p)}
              onClose={() => {
                setSelected(null)
                setSelectedIndex(null)
              }}
            />
          ) : null}
        </div>
      ) : (
        <SearchDocsResults
          loading={docsLoading}
          hasSearched={hasSearched}
          docsResults={docsResults}
          selected={selected}
          onSelectIndex={(idx) => selectDocByIndex(idx)}
          onCopyPath={(p) => void onCopyPath(p)}
          highlightQuery={docsHighlightQuery}
          hasMore={docsNextCursor !== ''}
          onLoadMore={() => void onLoadMore()}
        />
      )}

      {/* Errors are rendered via the main error banner near the top. */}
    </div>
  )
}
