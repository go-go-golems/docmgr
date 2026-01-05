import { useMemo, useState } from 'react'
import { Link, useNavigate, useSearchParams } from 'react-router-dom'

import { ApiErrorAlert } from '../../components/ApiErrorAlert'
import { DiagnosticCard } from '../../components/DiagnosticCard'
import { LoadingSpinner } from '../../components/LoadingSpinner'
import { MarkdownBlock } from '../../components/MarkdownBlock'
import { PageHeader } from '../../components/PageHeader'
import { RelatedFilesList } from '../../components/RelatedFilesList'
import { copyToClipboard } from '../../lib/clipboard'
import { formatDate } from '../../lib/time'
import { useGetDocQuery } from '../../services/docmgrApi'
import type { RelatedFile } from '../../services/docmgrApi'

type ToastState = { kind: 'success' | 'error'; message: string } | null

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
      await copyToClipboard(text)
      setToast({ kind: 'success', message: 'Copied' })
      setTimeout(() => setToast(null), 1200)
    } catch (e) {
      setToast({ kind: 'error', message: `Copy failed: ${String(e)}` })
      setTimeout(() => setToast(null), 2500)
    }
  }

  return (
    <div className="container py-4">
      <PageHeader
        title={title}
        subtitle={data?.path ?? path}
        subtitleClassName="font-monospace"
        actions={
          <>
            <button className="btn btn-outline-secondary" onClick={() => navigate(-1)}>
              Back
            </button>
            <Link className="btn btn-outline-primary" to="/">
              Search
            </Link>
          </>
        }
      />

      {toast ? (
        <div className={`alert ${toast.kind === 'success' ? 'alert-success' : 'alert-danger'} py-2`}>
          {toast.message}
        </div>
      ) : null}

      {path === '' ? <div className="alert alert-info">Missing doc path.</div> : null}

      {error ? <ApiErrorAlert title="Failed to load document" error={error} /> : null}

      {isLoading ? <LoadingSpinner /> : null}

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
                    <td>{formatDate(doc.lastUpdated)}</td>
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
                    <td>{formatDate(data.stats?.modTime)}</td>
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
            <RelatedFilesList files={relatedFiles} onCopyPath={(p) => void onCopy(p)} />
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
            <MarkdownBlock markdown={data.body} />
          </div>
        </div>
      ) : null}
    </div>
  )
}
