import { useMemo } from 'react'
import type { ReactNode } from 'react'
import { Link } from 'react-router-dom'

import type { RelatedFile } from '../services/docmgrApi'
import { timeAgo } from '../lib/time'

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

export function DocCard({
  title,
  ticket,
  docType,
  status,
  topics,
  path,
  lastUpdated,
  relatedFiles,
  selected,
  snippet,
  onSelect,
  onCopyPath,
  actions,
}: {
  title: string
  ticket: string
  docType: string
  status: string
  topics: string[]
  path: string
  lastUpdated?: string
  relatedFiles?: RelatedFile[]
  selected: boolean
  snippet?: ReactNode
  onSelect: () => void
  onCopyPath?: (path: string) => void
  actions?: ReactNode
}) {
  const topRelated = useMemo(() => (relatedFiles ?? []).slice(0, 3), [relatedFiles])
  const moreCount = (relatedFiles?.length ?? 0) - topRelated.length

  return (
    <div className={`result-card ${selected ? 'selected' : ''}`} onClick={onSelect} role="button" tabIndex={0}>
      <div className="d-flex justify-content-between align-items-start">
        <div className="flex-grow-1">
          <div className="result-title">{title}</div>
          <div className="result-meta">
            <Link
              to={`/ticket/${encodeURIComponent(ticket)}`}
              onClick={(e) => e.stopPropagation()}
              className="text-decoration-none"
            >
              {ticket}
            </Link>{' '}
            • {docType}
            <StatusBadge status={status} />
            {lastUpdated ? <span className="ms-2 text-muted">Updated {timeAgo(lastUpdated)}</span> : null}
          </div>

          <div className="mb-2">
            {(topics ?? []).map((topic) => (
              <span key={topic} className="badge text-bg-secondary topic-badge">
                {topic}
              </span>
            ))}
          </div>

          {snippet ? (
            <div className="result-snippet">
              <div className="small">{snippet}</div>
            </div>
          ) : null}

          <div className="result-path">{path}</div>

          {topRelated.length > 0 ? (
            <div className="mt-2">
              <div className="small text-muted mb-1">Related files</div>
              <ul className="mb-0 small">
                {topRelated.map((rf) => (
                  <li key={`${rf.path}:${rf.note ?? ''}`}>
                    <span className="font-monospace">{rf.path}</span>
                    {rf.note ? <span className="text-muted ms-2">{rf.note}</span> : null}
                  </li>
                ))}
                {moreCount > 0 ? <li className="text-muted">… {moreCount} more</li> : null}
              </ul>
            </div>
          ) : null}
        </div>

        {actions ? (
          <div className="ms-2 d-flex flex-column gap-2 align-items-end">{actions}</div>
        ) : onCopyPath ? (
          <button
            className="btn btn-sm btn-outline-primary copy-btn ms-2"
            onClick={(e) => {
              e.stopPropagation()
              onCopyPath(path)
            }}
          >
            Copy
          </button>
        ) : null}
      </div>
    </div>
  )
}
