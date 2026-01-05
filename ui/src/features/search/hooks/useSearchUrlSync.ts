import { useEffect, useRef, useState } from 'react'
import type { AppDispatch } from '../../../app/store'
import { setFilter, setMode, setQuery, type SearchFilters, type SearchMode } from '../searchSlice'

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

function normalizeMode(raw: string | null): SearchMode {
  const m = (raw ?? '').trim().toLowerCase()
  if (m === 'reverse' || m === 'files') return m
  return 'docs'
}

export function useSearchUrlSync({
  dispatch,
  mode,
  query,
  filters,
  selectedPath,
  previewOpen,
}: {
  dispatch: AppDispatch
  mode: SearchMode
  query: string
  filters: SearchFilters
  selectedPath: string
  previewOpen: boolean
}): { urlSyncReady: boolean; desiredSelectedPath: string; desiredPreviewOpen: boolean } {
  const [urlSyncReady, setURLSyncReady] = useState(false)
  const [desiredSelectedPath, setDesiredSelectedPath] = useState<string>('')
  const [desiredPreviewOpen, setDesiredPreviewOpen] = useState<boolean>(false)

  const urlWriteTimerRef = useRef<number | null>(null)

  useEffect(() => {
    const params = new URLSearchParams(window.location.search)
    const nextMode = normalizeMode(params.get('mode'))

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
      if (filters.orderBy.trim() !== '' && filters.orderBy.trim() !== 'rank') params.set('orderBy', filters.orderBy.trim())
      if (filters.includeArchived !== true) params.set('includeArchived', String(filters.includeArchived))
      if (filters.includeScripts !== true) params.set('includeScripts', String(filters.includeScripts))
      if (filters.includeControlDocs !== true) params.set('includeControlDocs', String(filters.includeControlDocs))

      if (selectedPath.trim() !== '') params.set('sel', selectedPath.trim())
      if (previewOpen) params.set('preview', 'true')

      const next = params.toString()
      const url = next === '' ? window.location.pathname : `${window.location.pathname}?${next}`
      window.history.replaceState({}, '', url)
    }, 250)

    return () => {
      if (urlWriteTimerRef.current != null) window.clearTimeout(urlWriteTimerRef.current)
    }
  }, [filters, mode, previewOpen, query, selectedPath, urlSyncReady])

  return { urlSyncReady, desiredSelectedPath, desiredPreviewOpen }
}

