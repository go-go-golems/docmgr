import { useMemo } from 'react'
import { Link, useNavigate, useSearchParams } from 'react-router-dom'

import hljs from 'highlight.js'

import { ApiErrorAlert } from '../../components/ApiErrorAlert'
import { CodeBlock } from '../../components/CodeBlock'
import { LoadingSpinner } from '../../components/LoadingSpinner'
import { PageHeader } from '../../components/PageHeader'
import { copyToClipboard } from '../../lib/clipboard'
import { useGetFileQuery } from '../../services/docmgrApi'
import { useToast } from '../toast/useToast'

export function FileViewerPage() {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const toast = useToast()

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
      await copyToClipboard(text)
      toast.success('Copied', { timeoutMs: 1200 })
    } catch (e) {
      toast.error(`Copy failed: ${String(e)}`, { timeoutMs: 2500 })
    }
  }

  return (
    <div className="container py-4">
      <PageHeader
        title="File"
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

      {path === '' ? <div className="alert alert-info">Missing file path.</div> : null}

      {error ? <ApiErrorAlert title="Failed to load file" error={error} /> : null}

      {isLoading ? <LoadingSpinner /> : null}

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

          <CodeBlock html={highlighted} language={data.language} />
        </>
      ) : null}
    </div>
  )
}
