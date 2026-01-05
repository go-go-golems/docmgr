import { Link } from 'react-router-dom'

import { PathHeader } from '../../../components/PathHeader'
import { RelatedFilesList } from '../../../components/RelatedFilesList'
import { timeAgo } from '../../../lib/time'
import type { SearchDocResult } from '../../../services/docmgrApi'
import { MarkdownSnippet } from '../components/MarkdownSnippet'

export function SearchPreviewContent({
  doc,
  highlightQuery,
  onCopyPath,
  showMeta = true,
}: {
  doc: SearchDocResult
  highlightQuery: string
  onCopyPath: (path: string) => void
  showMeta?: boolean
}) {
  return (
    <>
      {showMeta ? (
        <div className="text-muted small mb-2">
          <Link to={`/ticket/${encodeURIComponent(doc.ticket)}`} className="text-decoration-none">
            {doc.ticket}
          </Link>{' '}
          • {doc.docType} • {doc.status}
          {doc.lastUpdated ? <span className="ms-2">Updated {timeAgo(doc.lastUpdated)}</span> : null}
        </div>
      ) : null}

      <PathHeader
        path={doc.path}
        actions={
          <>
            <button className="btn btn-sm btn-outline-primary" onClick={() => onCopyPath(doc.path)}>
              Copy path
            </button>
            <Link className="btn btn-sm btn-primary" to={`/doc?path=${encodeURIComponent(doc.path)}`}>
              Open doc
            </Link>
          </>
        }
      />

      <div className="mb-3">
        <div className="text-muted small mb-1">Snippet</div>
        <div className="small">
          <MarkdownSnippet markdown={doc.snippet} query={highlightQuery} />
        </div>
      </div>

      {doc.relatedFiles && doc.relatedFiles.length > 0 ? (
        <div>
          <div className="text-muted small mb-1">Related files</div>
          <RelatedFilesList files={doc.relatedFiles} onCopyPath={onCopyPath} />
        </div>
      ) : null}
    </>
  )
}
