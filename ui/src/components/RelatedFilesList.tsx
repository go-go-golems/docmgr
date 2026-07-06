import { Link } from 'react-router-dom'

import type { RelatedFile } from '../services/docmgrApi'

// Compute the /file viewer target for a related file. Prefers the server-side
// resolution (root + resolvedPath); falls back to parsing anchored paths
// (repo://..., docs://...) client-side, then to the legacy repo-relative
// assumption. Returns null when the file cannot be served by the viewer
// (absolute paths outside the repo/docs roots).
function fileLinkTarget(rf: RelatedFile): { root: string; path: string } | null {
  if (rf.root && rf.resolvedPath) {
    if (rf.root === 'abs') return null
    return { root: rf.root, path: rf.resolvedPath }
  }
  const anchored = /^([a-z]+):\/\/(.*)$/.exec(rf.path)
  if (anchored) {
    const [, scheme, rest] = anchored
    if (scheme === 'repo') return { root: 'repo', path: rest }
    if (scheme === 'docs') return { root: 'docs', path: rest }
    return null
  }
  return { root: 'repo', path: rf.path }
}

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
      {files.map((rf) => {
        const target = fileLinkTarget(rf)
        return (
          <li key={`${rf.path}:${rf.note ?? ''}`} className="d-flex gap-2 align-items-start">
            <div className="flex-grow-1">
              <div className="font-monospace">{rf.path}</div>
              {rf.note ? <div className="text-muted">{rf.note}</div> : null}
            </div>
            <div className="d-flex gap-2">
              {showCopy ? (
                <button
                  className="btn btn-sm btn-outline-secondary"
                  onClick={() => onCopyPath(rf.resolvedPath ?? rf.path)}
                >
                  Copy
                </button>
              ) : null}
              {showOpen && target ? (
                <Link
                  className="btn btn-sm btn-outline-primary"
                  to={`/file?root=${target.root}&path=${encodeURIComponent(target.path)}`}
                >
                  {openLabel}
                </Link>
              ) : null}
            </div>
          </li>
        )
      })}
    </ul>
  )
}
