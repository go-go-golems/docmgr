import { useMemo, useState } from 'react'
import { Link, useNavigate, useSearchParams } from 'react-router-dom'

import { ApiErrorAlert } from '../../components/ApiErrorAlert'
import { DiagnosticCard } from '../../components/DiagnosticCard'
import { LoadingSpinner } from '../../components/LoadingSpinner'
import { MarkdownBlock } from '../../components/MarkdownBlock'
import { PageHeader } from '../../components/PageHeader'
import { RelatedFilesList } from '../../components/RelatedFilesList'
import { copyToClipboard } from '../../lib/clipboard'
import { apiErrorFromUnknown } from '../../lib/apiError'
import { formatDate } from '../../lib/time'
import {
  useGetDocQuery,
  useGetWorkspaceFacetsQuery,
  useRelateDocMutation,
  useUpdateDocMetaMutation,
} from '../../services/docmgrApi'
import type { RelatedFile } from '../../services/docmgrApi'
import { useToast } from '../toast/useToast'

function StatusEditor({ path, status }: { path: string; status: string }) {
  const toast = useToast()
  const { data: facets } = useGetWorkspaceFacetsQuery(undefined)
  const [updateMeta, updateState] = useUpdateDocMetaMutation()

  const statuses = useMemo(() => {
    const base = facets?.statuses ?? []
    if (status && !base.includes(status)) return [status, ...base]
    return base
  }, [facets, status])

  async function onChange(next: string) {
    if (!next || next === status) return
    try {
      await updateMeta({ path, field: 'Status', value: next }).unwrap()
      toast.success(`Status set to ${next}`, { timeoutMs: 1500 })
    } catch (e) {
      toast.error(`Status update failed: ${apiErrorFromUnknown(e).message}`, { timeoutMs: 3000 })
    }
  }

  return (
    <select
      className="form-select form-select-sm w-auto d-inline-block"
      value={status}
      disabled={updateState.isLoading}
      onChange={(e) => void onChange(e.target.value)}
    >
      {status === '' ? <option value="">—</option> : null}
      {statuses.map((s) => (
        <option key={s} value={s}>
          {s}
        </option>
      ))}
    </select>
  )
}

function SummaryEditor({ path, summary }: { path: string; summary: string }) {
  const toast = useToast()
  const [updateMeta, updateState] = useUpdateDocMetaMutation()
  const [editing, setEditing] = useState(false)
  const [draft, setDraft] = useState(summary)

  function startEditing() {
    setDraft(summary)
    setEditing(true)
  }

  async function onSave() {
    try {
      await updateMeta({ path, field: 'Summary', value: draft.trim() }).unwrap()
      toast.success('Summary updated', { timeoutMs: 1500 })
      setEditing(false)
    } catch (e) {
      toast.error(`Summary update failed: ${apiErrorFromUnknown(e).message}`, { timeoutMs: 3000 })
    }
  }

  if (!editing) {
    return (
      <div className="d-flex gap-2 align-items-start">
        <span className={summary ? '' : 'text-muted'}>{summary || '—'}</span>
        <button className="btn btn-sm btn-outline-secondary ms-auto" onClick={startEditing}>
          Edit
        </button>
      </div>
    )
  }

  return (
    <div className="vstack gap-2">
      <textarea
        className="form-control form-control-sm"
        rows={3}
        value={draft}
        onChange={(e) => setDraft(e.target.value)}
      />
      <div className="d-flex gap-2">
        <button className="btn btn-sm btn-primary" disabled={updateState.isLoading} onClick={() => void onSave()}>
          {updateState.isLoading ? 'Saving…' : 'Save'}
        </button>
        <button className="btn btn-sm btn-outline-secondary" onClick={() => setEditing(false)}>
          Cancel
        </button>
      </div>
    </div>
  )
}

function RelateFileForm({ path }: { path: string }) {
  const toast = useToast()
  const [relateDoc, relateState] = useRelateDocMutation()
  const [filePath, setFilePath] = useState('')
  const [note, setNote] = useState('')

  async function onRelate() {
    const p = filePath.trim()
    if (!p) return
    try {
      const res = await relateDoc({ path, add: [{ path: p, note: note.trim() }] }).unwrap()
      if (res.added > 0) {
        toast.success(`Related ${p}`, { timeoutMs: 1500 })
      } else if (res.updated > 0) {
        toast.success(`Updated note for ${p}`, { timeoutMs: 1500 })
      } else {
        toast.success('Already related (no change)', { timeoutMs: 1500 })
      }
      setFilePath('')
      setNote('')
    } catch (e) {
      toast.error(`Relate failed: ${apiErrorFromUnknown(e).message}`, { timeoutMs: 3000 })
    }
  }

  return (
    <div className="row g-2 align-items-end">
      <div className="col-12 col-md-5">
        <label className="form-label small mb-1">Relate file…</label>
        <input
          className="form-control form-control-sm font-monospace"
          value={filePath}
          onChange={(e) => setFilePath(e.target.value)}
          placeholder="pkg/foo/bar.go or repo://…"
        />
      </div>
      <div className="col-12 col-md-5">
        <label className="form-label small mb-1">Note</label>
        <input
          className="form-control form-control-sm"
          value={note}
          onChange={(e) => setNote(e.target.value)}
          placeholder="Why this file matters"
        />
      </div>
      <div className="col-12 col-md-2">
        <button
          className="btn btn-sm btn-outline-primary w-100"
          disabled={relateState.isLoading || filePath.trim() === ''}
          onClick={() => void onRelate()}
        >
          {relateState.isLoading ? 'Relating…' : 'Relate'}
        </button>
      </div>
    </div>
  )
}

export function DocViewerPage() {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const toast = useToast()

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
      toast.success('Copied', { timeoutMs: 1200 })
    } catch (e) {
      toast.error(`Copy failed: ${String(e)}`, { timeoutMs: 2500 })
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
            <Link className="btn btn-outline-primary" to="/search">
              Search
            </Link>
          </>
        }
      />

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
                    <td>
                      <StatusEditor path={data.path} status={doc.status ?? ''} />
                    </td>
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
                    <th className="text-muted">Summary</th>
                    <td colSpan={3}>
                      <SummaryEditor path={data.path} summary={doc.summary ?? ''} />
                    </td>
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

      {doc ? (
        <div className="card mb-3">
          <div className="card-header fw-semibold">Related files</div>
          <div className="card-body vstack gap-3">
            {relatedFiles.length > 0 ? (
              <RelatedFilesList files={relatedFiles} onCopyPath={(p) => void onCopy(p)} />
            ) : (
              <div className="text-muted small">No related files yet.</div>
            )}
            <RelateFileForm path={data?.path ?? path} />
          </div>
        </div>
      ) : relatedFiles.length > 0 ? (
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
            <MarkdownBlock markdown={data.body} docPath={data.path} />
          </div>
        </div>
      ) : null}
    </div>
  )
}
