import type { SearchDocResult } from '../../../services/docmgrApi'
import { SearchPreviewContent } from './SearchPreviewContent'

export function SearchPreviewModal({
  open,
  selected,
  highlightQuery,
  onCopyPath,
  onDismiss,
  onClosePreview,
}: {
  open: boolean
  selected: SearchDocResult | null
  highlightQuery: string
  onCopyPath: (path: string) => void
  onDismiss: () => void
  onClosePreview: () => void
}) {
  if (!open || !selected) return null

  return (
    <>
      <div className="modal-backdrop show" />
      <div className="modal show d-block" tabIndex={-1} role="dialog" aria-modal="true">
        <div className="modal-dialog modal-fullscreen-sm-down modal-dialog-scrollable" role="document">
          <div className="modal-content">
            <div className="modal-header">
              <h5 className="modal-title">{selected.title}</h5>
              <button type="button" className="btn-close" onClick={onDismiss} />
            </div>
            <div className="modal-body">
              <SearchPreviewContent doc={selected} highlightQuery={highlightQuery} onCopyPath={onCopyPath} />
            </div>
            <div className="modal-footer">
              <button type="button" className="btn btn-outline-secondary" onClick={onClosePreview}>
                Close preview
              </button>
            </div>
          </div>
        </div>
      </div>
    </>
  )
}
