import { useCallback, useEffect, useRef, useState } from 'react'

import type { SearchDocResult } from '../../../services/docmgrApi'

export function useSearchSelection({
  docsResults,
  isMobile,
  urlSyncReady,
  desiredSelectedPath,
  desiredPreviewOpen,
  openPreviewModal,
  closePreviewModal,
}: {
  docsResults: SearchDocResult[]
  isMobile: boolean
  urlSyncReady: boolean
  desiredSelectedPath: string
  desiredPreviewOpen: boolean
  openPreviewModal: () => void
  closePreviewModal: () => void
}): {
  selected: SearchDocResult | null
  selectedIndex: number | null
  setSelected: (next: SearchDocResult | null) => void
  setSelectedIndex: (next: number | null) => void
  clearSelection: (opts?: { keepIndex?: boolean; closePreview?: boolean }) => void
  selectDocByIndex: (idx: number, opts?: { openModal?: boolean }) => void
} {
  const [selected, setSelected] = useState<SearchDocResult | null>(null)
  const [selectedIndex, setSelectedIndex] = useState<number | null>(null)

  const selectionAppliedRef = useRef(false)

  const selectDocByIndex = useCallback(
    (idx: number, opts?: { openModal?: boolean }) => {
      if (idx < 0 || idx >= docsResults.length) return
      setSelected(docsResults[idx])
      setSelectedIndex(idx)
      if (isMobile && (opts?.openModal ?? true)) openPreviewModal()
    },
    [docsResults, isMobile, openPreviewModal],
  )

  const clearSelection = useCallback(
    (opts?: { keepIndex?: boolean; closePreview?: boolean }) => {
      setSelected(null)
      if (!(opts?.keepIndex ?? false)) setSelectedIndex(null)
      if (opts?.closePreview ?? true) closePreviewModal()
    },
    [closePreviewModal],
  )

  useEffect(() => {
    if (!urlSyncReady) return
    if (selectionAppliedRef.current) return
    if (desiredSelectedPath.trim() === '') return
    if (docsResults.length === 0) return

    const idx = docsResults.findIndex((d) => d.path === desiredSelectedPath.trim())
    selectionAppliedRef.current = true
    if (idx < 0) return

    let canceled = false
    queueMicrotask(() => {
      if (canceled) return
      setSelected(docsResults[idx])
      setSelectedIndex(idx)
      if (isMobile && desiredPreviewOpen) openPreviewModal()
    })

    return () => {
      canceled = true
    }
  }, [desiredPreviewOpen, desiredSelectedPath, docsResults, isMobile, openPreviewModal, urlSyncReady])

  return { selected, selectedIndex, setSelected, setSelectedIndex, clearSelection, selectDocByIndex }
}
