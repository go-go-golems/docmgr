import { useMemo, useState } from 'react'
import { Link, useNavigate, useSearchParams } from 'react-router-dom'

import hljs from 'highlight.js'

import { useGetFileQuery } from '../../services/docmgrApi'

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

export function FileViewerPage() {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const [toast, setToast] = useState<ToastState>(null)

  const path = (searchParams.get('path') ?? '').trim()
  const rootParam = (searchParams.get('root') ?? 'repo').trim().toLowerCase()
  const root: 'repo' | 'docs' = rootParam === 'docs' ? 'docs' : 'repo'

  const { data, error, isLoading } = useGetFileQuery({ path, root }, { skip: path === '' })

  const highlighted = useMemo(() => {
    if (!data) return ''
    const lang = (data.language || '').trim()
    try {
      if (lang && hljs.getLanguage(lang)) {
        return hljs.highlight(data.content, { language: lang }).value
      }
      return hljs.highlightAuto(data.content).value
    } catch {
      return data.content
        .replaceAll('&', '&amp;')
        .replaceAll('<', '&lt;')
        .replaceAll('>', '&gt;')
        .replaceAll('"', '&quot;')
        .replaceAll("'", '&#39;')
    }
  }, [data])

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
          <div className="h4 mb-0">File</div>
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

      {path === '' ? <div className="alert alert-info">Missing file path.</div> : null}

      {error ? (
        (() => {
          const b = toErrorBanner(error, 'Failed to load file')
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

      {data ? (
        <>
          {data.truncated ? (
            <div className="alert alert-warning">
              File is truncated for safety. Showing the first {data.content.length} bytes.
            </div>
          ) : null}

          <div className="d-flex gap-2 mb-2">
            <button className="btn btn-sm btn-outline-primary" onClick={() => void onCopy(data.path)}>
              Copy path
            </button>
            <button className="btn btn-sm btn-outline-secondary" onClick={() => void onCopy(data.content)}>
              Copy content
            </button>
            <div className="ms-auto text-muted small">
              {root} • {data.language || 'text'} • {data.stats?.sizeBytes ?? 0} bytes
            </div>
          </div>

          <pre className="docmgr-code">
            <code
              className={`hljs ${data.language ? `language-${data.language}` : ''}`}
              dangerouslySetInnerHTML={{ __html: highlighted }}
            />
          </pre>
        </>
      ) : null}
    </div>
  )
}
