import { useMemo, useState } from 'react'
import { Link, useNavigate, useSearchParams } from 'react-router-dom'

import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import rehypeHighlight from 'rehype-highlight'

import { useGetDocQuery } from '../../services/docmgrApi'
import type { DiagnosticTaxonomy, RelatedFile } from '../../services/docmgrApi'

type ToastState = { kind: 'success' | 'error'; message: string } | null

type APIErrorPayload = {
  error?: {
    code?: string
    message?: string
    details?: unknown
  }
}

type ErrorBanner = {
  title: string
  code?: string
  message: string
  details?: unknown
}

function toErrorBanner(err: unknown, title: string): ErrorBanner {
  const maybe = err as { data?: unknown; status?: number } | undefined
  const data = maybe?.data as APIErrorPayload | undefined
  const code = data?.error?.code
  const message =
    data?.error?.message ??
    (typeof err === 'string' ? err : err instanceof Error ? err.message : String(err))
  const details = data?.error?.details
  return { title, code, message, details }
}

function DiagnosticCard({ diag }: { diag: DiagnosticTaxonomy }) {
  const severity = (diag.Severity || 'info').toLowerCase()
  const badge =
    severity === 'warning'
      ? 'warning'
      : severity === 'error'
        ? 'danger'
        : severity === 'info'
          ? 'info'
          : 'secondary'

  const reason =
    typeof diag.Context === 'object' && diag.Context != null ? (diag.Context['Reason'] as unknown) : undefined

  return (
    <div className="alert alert-warning">
      <div className="fw-semibold mb-1">
        <span className={`badge text-bg-${badge} me-2`}>{diag.Severity ?? 'info'}</span>
        Parse diagnostics
      </div>
      <div className="small">
        {(diag.Stage ?? 'unknown') + (diag.Symptom ? ` • ${diag.Symptom}` : '')}
      </div>
      {diag.Path ? (
        <div className="small mt-1">
          <span className="text-muted">Path: </span>
          <span className="font-monospace">{diag.Path}</span>
        </div>
      ) : null}
      {typeof reason === 'string' && reason.trim() !== '' ? (
        <div className="small mt-1 text-muted">{reason}</div>
      ) : null}
    </div>
  )
}

export function DocViewerPage() {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const [toast, setToast] = useState<ToastState>(null)

  const path = (searchParams.get('path') ?? '').trim()
  const { data, error, isLoading } = useGetDocQuery({ path }, { skip: path === '' })

  const doc = data?.doc
  const title = doc?.title?.trim() ? doc.title : data?.path ?? 'Document'

  const relatedFiles: RelatedFile[] = useMemo(() => {
    if (data?.relatedFiles) return data.relatedFiles
    if (doc?.relatedFiles) return doc.relatedFiles
    return []
  }, [data, doc])

  async function onCopy(text: string) {
    try {
      if (!navigator.clipboard) throw new Error('clipboard not available')
      await navigator.clipboard.writeText(text)
      setToast({ kind: 'success', message: 'Copied' })
      setTimeout(() => setToast(null), 1200)
    } catch (e) {
      setToast({ kind: 'error', message: `Copy failed: ${String(e)}` })
      setTimeout(() => setToast(null), 2500)
    }
  }

  return (
    <div className="container py-4">
      <div className="d-flex justify-content-between align-items-center mb-3 gap-2">
        <div>
          <div className="h4 mb-0">{title}</div>
          <div className="text-muted small font-monospace">{data?.path ?? path}</div>
        </div>
        <div className="d-flex gap-2">
          <button className="btn btn-outline-secondary" onClick={() => navigate(-1)}>
            Back
          </button>
          <Link className="btn btn-outline-primary" to="/">
            Search
          </Link>
        </div>
      </div>

      {toast ? (
        <div className={`alert ${toast.kind === 'success' ? 'alert-success' : 'alert-danger'} py-2`}>
          {toast.message}
        </div>
      ) : null}

      {path === '' ? <div className="alert alert-info">Missing doc path.</div> : null}

      {error ? (
        (() => {
          const b = toErrorBanner(error, 'Failed to load document')
          return (
            <div className="alert alert-danger">
              <div className="fw-semibold">{b.title}</div>
              <div className="small">
                {b.code ? <span className="me-2">({b.code})</span> : null}
                {b.message}
              </div>
              {b.details ? <pre className="small mb-0 mt-2">{JSON.stringify(b.details, null, 2)}</pre> : null}
            </div>
          )
        })()
      ) : null}

      {isLoading ? (
        <div className="text-center my-4">
          <div className="spinner-border text-primary" role="status" />
        </div>
      ) : null}

      {data?.diagnostic ? <DiagnosticCard diag={data.diagnostic} /> : null}

      {doc ? (
        <div className="card mb-3">
          <div className="card-body">
            <div className="d-flex flex-wrap gap-2 align-items-center mb-2">
              <Link
                className="badge text-bg-secondary text-decoration-none"
                to={`/ticket/${encodeURIComponent(doc.ticket)}`}
              >
                {doc.ticket}
              </Link>
              <span className="badge text-bg-light text-dark">{doc.docType}</span>
              {doc.status ? <span className="badge text-bg-primary">{doc.status}</span> : null}
              <span className="ms-auto d-flex gap-2">
                <button className="btn btn-sm btn-outline-primary" onClick={() => void onCopy(data.path)}>
                  Copy path
                </button>
                <button className="btn btn-sm btn-outline-secondary" onClick={() => void onCopy(data.body)}>
                  Copy markdown
                </button>
              </span>
            </div>

            <div className="table-responsive">
              <table className="table table-sm mb-0">
                <tbody>
                  <tr>
                    <th className="text-muted">Ticket</th>
                    <td>
                      <Link to={`/ticket/${encodeURIComponent(doc.ticket)}`}>{doc.ticket}</Link>
                    </td>
                    <th className="text-muted">Last updated</th>
                    <td>{doc.lastUpdated ? new Date(doc.lastUpdated).toLocaleString() : '—'}</td>
                  </tr>
                  <tr>
                    <th className="text-muted">Status</th>
                    <td>{doc.status || '—'}</td>
                    <th className="text-muted">Intent</th>
                    <td>{doc.intent || '—'}</td>
                  </tr>
                  <tr>
                    <th className="text-muted">Topics</th>
                    <td>{doc.topics?.length ? doc.topics.join(', ') : '—'}</td>
                    <th className="text-muted">Owners</th>
                    <td>{doc.owners?.length ? doc.owners.join(', ') : '—'}</td>
                  </tr>
                  <tr>
                    <th className="text-muted">File mtime</th>
                    <td>{data.stats?.modTime ? new Date(data.stats.modTime).toLocaleString() : '—'}</td>
                    <th className="text-muted">Size</th>
                    <td>{data.stats?.sizeBytes ? `${data.stats.sizeBytes} bytes` : '—'}</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>
      ) : null}

      {relatedFiles.length > 0 ? (
        <div className="card mb-3">
          <div className="card-header fw-semibold">Related files</div>
          <div className="card-body">
            <ul className="list-unstyled mb-0 vstack gap-2">
              {relatedFiles.map((rf) => (
                <li key={`${rf.path}:${rf.note ?? ''}`} className="d-flex gap-2 align-items-start">
                  <div className="flex-grow-1">
                    <div className="font-monospace">{rf.path}</div>
                    {rf.note ? <div className="text-muted small">{rf.note}</div> : null}
                  </div>
                  <div className="d-flex gap-2">
                    <button className="btn btn-sm btn-outline-secondary" onClick={() => void onCopy(rf.path)}>
                      Copy
                    </button>
                    <Link
                      className="btn btn-sm btn-outline-primary"
                      to={`/file?root=repo&path=${encodeURIComponent(rf.path)}`}
                    >
                      Open
                    </Link>
                  </div>
                </li>
              ))}
            </ul>
          </div>
        </div>
      ) : null}

      {data?.body ? (
        <div className="card">
          <div className="card-header d-flex justify-content-between align-items-center">
            <span className="fw-semibold">Content</span>
            {data?.path ? (
              <button className="btn btn-sm btn-outline-secondary" onClick={() => void onCopy(data.body)}>
                Copy
              </button>
            ) : null}
          </div>
          <div className="card-body docmgr-markdown">
            <ReactMarkdown remarkPlugins={[remarkGfm]} rehypePlugins={[rehypeHighlight]}>
              {data.body}
            </ReactMarkdown>
          </div>
        </div>
      ) : null}
    </div>
  )
}
