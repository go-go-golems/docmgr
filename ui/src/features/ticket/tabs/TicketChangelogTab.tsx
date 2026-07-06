import { useMemo, useState } from 'react'
import { Link } from 'react-router-dom'

import { ApiErrorAlert } from '../../../components/ApiErrorAlert'
import { LoadingSpinner } from '../../../components/LoadingSpinner'
import { MarkdownBlock } from '../../../components/MarkdownBlock'
import { apiErrorFromUnknown } from '../../../lib/apiError'
import {
  useAppendTicketChangelogMutation,
  useGetTicketChangelogQuery,
} from '../../../services/docmgrApi'
import { useToast } from '../../toast/useToast'

export function TicketChangelogTab({ ticket, ticketDir }: { ticket: string; ticketDir?: string }) {
  const toast = useToast()
  const { data, error, isLoading } = useGetTicketChangelogQuery({ ticket }, { skip: ticket === '' })
  const [appendEntry, appendState] = useAppendTicketChangelogMutation()

  const [title, setTitle] = useState('')
  const [entry, setEntry] = useState('')

  // Entries come in file order (newest appended last); show newest first.
  const entries = useMemo(() => [...(data?.entries ?? [])].reverse(), [data])

  async function onAppend() {
    const text = entry.trim()
    if (!text) return
    try {
      await appendEntry({ ticket, title: title.trim(), entry: text }).unwrap()
      toast.success('Changelog entry added', { timeoutMs: 1500 })
      setTitle('')
      setEntry('')
    } catch (e) {
      toast.error(`Append failed: ${apiErrorFromUnknown(e).message}`, { timeoutMs: 3000 })
    }
  }

  return (
    <div className="row g-3">
      <div className="col-12 col-lg-7">
        {error ? <ApiErrorAlert title="Failed to load changelog" error={error} /> : null}
        {isLoading ? <LoadingSpinner /> : null}

        {!isLoading && data ? (
          entries.length === 0 ? (
            <div className="card">
              <div className="card-body text-muted">
                {data.exists ? 'changelog.md has no dated entries yet.' : 'No changelog.md yet — add the first entry.'}
              </div>
            </div>
          ) : (
            <div className="vstack gap-3">
              {entries.map((e, idx) => (
                <div key={`${e.date}:${e.title}:${idx}`} className="card">
                  <div className="card-header d-flex flex-wrap gap-2 align-items-center">
                    <span className="badge text-bg-secondary">{e.date || '—'}</span>
                    <span className="fw-semibold">{e.title || 'Entry'}</span>
                  </div>
                  <div className="card-body docmgr-markdown">
                    <MarkdownBlock markdown={e.body} docPath={data.path} />
                  </div>
                </div>
              ))}
            </div>
          )
        ) : null}
      </div>

      <div className="col-12 col-lg-5">
        <div className="card">
          <div className="card-header fw-semibold d-flex justify-content-between align-items-center">
            <span>Add entry</span>
            {ticketDir ? (
              <Link
                className="btn btn-sm btn-outline-secondary"
                to={`/doc?path=${encodeURIComponent(`${ticketDir}/changelog.md`)}`}
              >
                Open changelog.md
              </Link>
            ) : null}
          </div>
          <div className="card-body vstack gap-2">
            <input
              className="form-control form-control-sm"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              placeholder="Title (optional)"
            />
            <textarea
              className="form-control form-control-sm"
              rows={4}
              value={entry}
              onChange={(e) => setEntry(e.target.value)}
              placeholder="What changed?"
            />
            <button
              className="btn btn-primary btn-sm"
              disabled={appendState.isLoading || entry.trim() === ''}
              onClick={() => void onAppend()}
            >
              {appendState.isLoading ? 'Appending…' : 'Append entry'}
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}
