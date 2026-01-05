import { useMemo } from 'react'
import type { ReactNode } from 'react'
import { Link } from 'react-router-dom'

import type { RelatedFile } from '../services/docmgrApi'
import { timeAgo } from '../lib/time'

export type DocCardDoc = {
  ticket: string
  path: string
  title?: string
  docType?: string
  status?: string
  topics?: string[]
  lastUpdated?: string
  relatedFiles?: RelatedFile[]
}

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
  doc,
  selected,
  snippet,
  onSelect,
  onCopyPath,
  actions,
  showRelatedFiles,
}: {
  doc: DocCardDoc
  selected: boolean
  snippet?: ReactNode
  onSelect: () => void
  onCopyPath?: (path: string) => void
  actions?: ReactNode
  showRelatedFiles?: boolean
}) {
  const title = doc.title?.trim() ? doc.title : doc.path
  const topics = doc.topics ?? []

  const topRelated = useMemo(() => (doc.relatedFiles ?? []).slice(0, 3), [doc.relatedFiles])
  const moreCount = (doc.relatedFiles?.length ?? 0) - topRelated.length
  const shouldShowRelated = (showRelatedFiles ?? true) && topRelated.length > 0

  return (
    <div
      className={`dm-card ${selected ? 'dm-card--selected' : ''}`.trim()}
      onClick={onSelect}
      onKeyDown={(e) => {
        if (e.key === 'Enter' || e.key === ' ') {
          e.preventDefault()
          onSelect()
        }
      }}
      role="button"
      tabIndex={0}
    >
      <div className="d-flex justify-content-between align-items-start">
        <div className="flex-grow-1">
          <div className="dm-card-title">{title}</div>
          <div className="dm-card-meta">
            <Link
              to={`/ticket/${encodeURIComponent(doc.ticket)}`}
              onClick={(e) => e.stopPropagation()}
              className="text-decoration-none"
            >
              {doc.ticket}
            </Link>
            {doc.docType ? (
              <>
                {' '}
                • {doc.docType}
              </>
            ) : null}
            {doc.status ? <StatusBadge status={doc.status} /> : null}
            {doc.lastUpdated ? <span className="ms-2 text-muted">Updated {timeAgo(doc.lastUpdated)}</span> : null}
          </div>

          <div className="mb-2">
            {topics.map((topic) => (
              <span key={topic} className="badge text-bg-secondary dm-topic-badge">
                {topic}
              </span>
            ))}
          </div>

          {snippet ? (
            <div className="dm-card-snippet">
              <div className="small">{snippet}</div>
            </div>
          ) : null}

          <div className="dm-path-pill">{doc.path}</div>

          {shouldShowRelated ? (
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
            className="btn btn-sm btn-outline-primary dm-copy-btn ms-2"
            onClick={(e) => {
              e.stopPropagation()
              onCopyPath(doc.path)
            }}
          >
            Copy
          </button>
        ) : null}
      </div>
    </div>
  )
}
