import { Link } from 'react-router-dom'

import { timeAgo } from '../../../lib/time'
import { SearchPreviewContent } from './SearchPreviewContent'

import type { SearchDocResult } from '../../../services/docmgrApi'

export function SearchPreviewPanel({
  selected,
  highlightQuery,
  onCopyPath,
  onClose,
}: {
  selected: SearchDocResult
  highlightQuery: string
  onCopyPath: (path: string) => void
  onClose: () => void
}) {
  return (
    <div className="preview-panel">
      <div className="d-flex justify-content-between align-items-start mb-2">
        <div>
          <div className="h5 mb-1">{selected.title}</div>
          <div className="text-muted small">
            <Link to={`/ticket/${encodeURIComponent(selected.ticket)}`} className="text-decoration-none">
              {selected.ticket}
            </Link>{' '}
            • {selected.docType} • {selected.status}
            {selected.lastUpdated ? <span className="ms-2">Updated {timeAgo(selected.lastUpdated)}</span> : null}
          </div>
        </div>
        <button className="btn btn-sm btn-outline-secondary" onClick={onClose}>
          Close
        </button>
      </div>
      <SearchPreviewContent
        doc={selected}
        highlightQuery={highlightQuery}
        onCopyPath={onCopyPath}
        showMeta={false}
      />
    </div>
  )
}
