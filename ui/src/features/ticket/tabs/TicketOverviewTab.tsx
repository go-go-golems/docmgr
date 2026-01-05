import { Link } from 'react-router-dom'

import { ApiErrorAlert } from '../../../components/ApiErrorAlert'
import { EmptyState } from '../../../components/EmptyState'
import { MarkdownBlock } from '../../../components/MarkdownBlock'
import type { TicketDocItem, TicketGetResponse } from '../../../services/docmgrApi'

type OpenTask = { id: number; text: string; checked: boolean }

export function TicketOverviewTab({
  ticket,
  ticketData,
  docsLoading,
  docsError,
  keyDocs,
  tasksLoading,
  tasksError,
  openTasks,
  indexDocError,
  indexBody,
  onSetTab,
  onCheckTask,
  checkTaskLoading,
  formatDate,
}: {
  ticket: string
  ticketData?: TicketGetResponse
  docsLoading: boolean
  docsError?: unknown
  keyDocs: TicketDocItem[]
  tasksLoading: boolean
  tasksError?: unknown
  openTasks: OpenTask[]
  indexDocError?: unknown
  indexBody?: string
  onSetTab: (tab: 'documents' | 'tasks') => void
  onCheckTask: (args: { ticket: string; ids: number[]; checked: boolean }) => Promise<unknown>
  checkTaskLoading: boolean
  formatDate: (iso?: string) => string
}) {
  return (
    <div className="row g-3">
      <div className="col-12 col-lg-6">
        <div className="card">
          <div className="card-header fw-semibold">Metadata</div>
          <div className="card-body">
            <div className="table-responsive">
              <table className="table table-sm mb-0">
                <tbody>
                  <tr>
                    <th className="text-muted">Status</th>
                    <td>{ticketData?.status || '—'}</td>
                    <th className="text-muted">Intent</th>
                    <td>{ticketData?.intent || '—'}</td>
                  </tr>
                  <tr>
                    <th className="text-muted">Created</th>
                    <td>{ticketData?.createdAt || '—'}</td>
                    <th className="text-muted">Updated</th>
                    <td>{formatDate(ticketData?.updatedAt)}</td>
                  </tr>
                  <tr>
                    <th className="text-muted">Topics</th>
                    <td colSpan={3}>{ticketData?.topics?.length ? ticketData.topics.join(', ') : '—'}</td>
                  </tr>
                  <tr>
                    <th className="text-muted">Owners</th>
                    <td colSpan={3}>{ticketData?.owners?.length ? ticketData.owners.join(', ') : '—'}</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </div>

      <div className="col-12 col-lg-6">
        <div className="card">
          <div className="card-header fw-semibold">Quick stats</div>
          <div className="card-body">
            <div className="d-flex flex-wrap gap-2">
              <span className="badge text-bg-light text-dark">
                Docs: <span className="fw-semibold">{ticketData?.stats?.docsTotal ?? 0}</span>
              </span>
              <span className="badge text-bg-light text-dark">
                Tasks:{' '}
                <span className="fw-semibold">
                  {ticketData?.stats?.tasksDone ?? 0}/{ticketData?.stats?.tasksTotal ?? 0}
                </span>
              </span>
              <span className="badge text-bg-light text-dark">
                Files: <span className="fw-semibold">{ticketData?.stats?.relatedFilesTotal ?? 0}</span>
              </span>
            </div>
            <div className="mt-3 d-flex flex-wrap gap-2">
              {ticketData?.indexPath ? (
                <Link className="btn btn-sm btn-outline-primary" to={`/doc?path=${encodeURIComponent(ticketData.indexPath)}`}>
                  Open index.md
                </Link>
              ) : null}
              <button className="btn btn-sm btn-outline-secondary" onClick={() => onSetTab('documents')}>
                Documents
              </button>
              <button className="btn btn-sm btn-outline-secondary" onClick={() => onSetTab('tasks')}>
                Tasks
              </button>
            </div>
          </div>
        </div>
      </div>

      <div className="col-12 col-lg-6">
        <div className="card">
          <div className="card-header fw-semibold">Key documents</div>
          <div className="card-body">
            {docsLoading ? (
              <div className="text-muted">Loading…</div>
            ) : docsError ? (
              <ApiErrorAlert title="Failed to load docs" error={docsError} />
            ) : keyDocs.length === 0 ? (
              <div className="text-muted">No documents found.</div>
            ) : (
              <ul className="list-unstyled mb-0 vstack gap-2">
                {keyDocs.map((d) => (
                  <li key={d.path} className="d-flex justify-content-between align-items-start gap-2">
                    <div className="flex-grow-1">
                      <div className="fw-semibold">{d.title || d.path}</div>
                      <div className="small text-muted font-monospace">{d.path}</div>
                    </div>
                    <Link className="btn btn-sm btn-outline-primary" to={`/doc?path=${encodeURIComponent(d.path)}`}>
                      Open
                    </Link>
                  </li>
                ))}
              </ul>
            )}
          </div>
        </div>
      </div>

      <div className="col-12 col-lg-6">
        <div className="card">
          <div className="card-header fw-semibold">Open tasks</div>
          <div className="card-body">
            {tasksLoading ? (
              <div className="text-muted">Loading…</div>
            ) : tasksError ? (
              <ApiErrorAlert title="Failed to load tasks" error={tasksError} />
            ) : openTasks.length === 0 ? (
              <div className="text-muted">No open tasks.</div>
            ) : (
              <div className="vstack gap-2">
                {openTasks.map((it) => (
                  <label key={it.id} className="d-flex gap-2 align-items-start">
                    <input
                      type="checkbox"
                      checked={it.checked}
                      onChange={(e) => void onCheckTask({ ticket, ids: [it.id], checked: e.target.checked })}
                      disabled={checkTaskLoading}
                    />
                    <span>
                      <span className="text-muted me-2">#{it.id}</span>
                      {it.text}
                    </span>
                  </label>
                ))}
                <button className="btn btn-sm btn-outline-secondary align-self-start" onClick={() => onSetTab('tasks')}>
                  View all tasks
                </button>
              </div>
            )}
          </div>
        </div>
      </div>

      <div className="col-12">
        <div className="card">
          <div className="card-header fw-semibold">index.md</div>
          <div className="card-body docmgr-markdown">
            {indexDocError ? (
              <ApiErrorAlert title="Failed to load index.md" error={indexDocError} />
            ) : indexBody ? (
              <MarkdownBlock markdown={indexBody} />
            ) : (
              <EmptyState title="No content." />
            )}
          </div>
        </div>
      </div>
    </div>
  )
}
