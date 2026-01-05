import { Link } from 'react-router-dom'

import type { RelatedFile } from '../services/docmgrApi'

export function RelatedFilesList({
  files,
  onCopyPath,
  showCopy = true,
  showOpen = true,
  openLabel = 'Open',
}: {
  files: RelatedFile[]
  onCopyPath: (path: string) => void
  showCopy?: boolean
  showOpen?: boolean
  openLabel?: string
}) {
  if (files.length === 0) return null

  return (
    <ul className="small mb-0 list-unstyled vstack gap-2">
      {files.map((rf) => (
        <li key={`${rf.path}:${rf.note ?? ''}`} className="d-flex gap-2 align-items-start">
          <div className="flex-grow-1">
            <div className="font-monospace">{rf.path}</div>
            {rf.note ? <div className="text-muted">{rf.note}</div> : null}
          </div>
          <div className="d-flex gap-2">
            {showCopy ? (
              <button className="btn btn-sm btn-outline-secondary" onClick={() => onCopyPath(rf.path)}>
                Copy
              </button>
            ) : null}
            {showOpen ? (
              <Link
                className="btn btn-sm btn-outline-primary"
                to={`/file?root=repo&path=${encodeURIComponent(rf.path)}`}
              >
                {openLabel}
              </Link>
            ) : null}
          </div>
        </li>
      ))}
    </ul>
  )
}
