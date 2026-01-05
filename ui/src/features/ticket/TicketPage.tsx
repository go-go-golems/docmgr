import { useMemo, useState } from 'react'
import { Link, useParams, useSearchParams } from 'react-router-dom'

import {
  useAddTicketTaskMutation,
  useCheckTicketTasksMutation,
  useGetTicketDocsQuery,
  useGetTicketGraphQuery,
  useGetTicketQuery,
  useGetTicketTasksQuery,
  type TicketDocItem,
} from '../../services/docmgrApi'

type TabKey = 'overview' | 'documents' | 'tasks' | 'graph' | 'changelog'

function normalizeTab(raw: string | null): TabKey {
  const t = (raw ?? '').trim().toLowerCase()
  if (t === 'documents' || t === 'tasks' || t === 'graph' || t === 'changelog') return t
  return 'overview'
}

function formatDate(iso: string | undefined): string {
  if (!iso) return '—'
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return iso
  return d.toLocaleString()
}

function isRecord(v: unknown): v is Record<string, unknown> {
  return typeof v === 'object' && v !== null
}

function apiErrorMessage(err: unknown): string {
  const maybe = err as { data?: unknown } | undefined
  const data = isRecord(maybe?.data) ? maybe?.data : undefined
  const e = data && isRecord(data['error']) ? (data['error'] as Record<string, unknown>) : undefined
  const msg = e && typeof e['message'] === 'string' ? (e['message'] as string) : undefined
  if (msg && msg.trim() !== '') return msg
  if (err instanceof Error) return err.message
  if (typeof err === 'string') return err
  return String(err)
}

function groupByDocType(items: TicketDocItem[]): Record<string, TicketDocItem[]> {
  const out: Record<string, TicketDocItem[]> = {}
  for (const it of items) {
    const k = it.docType?.trim() || 'unknown'
    out[k] ||= []
    out[k].push(it)
  }
  return out
}

export function TicketPage() {
  const params = useParams()
  const ticket = (params.ticket ?? '').trim()
  const [searchParams, setSearchParams] = useSearchParams()

  const tab = normalizeTab(searchParams.get('tab'))
  const selectedDoc = (searchParams.get('doc') ?? '').trim()

  const { data: t, error: ticketError, isLoading: ticketLoading } = useGetTicketQuery(
    { ticket },
    { skip: ticket === '' },
  )

  const { data: docsData, error: docsError, isLoading: docsLoading } = useGetTicketDocsQuery(
    { ticket, pageSize: 500, orderBy: 'path' },
    { skip: ticket === '' || tab !== 'documents' },
  )

  const { data: tasksData, error: tasksError, isLoading: tasksLoading } = useGetTicketTasksQuery(
    { ticket },
    { skip: ticket === '' || tab !== 'tasks' },
  )

  const { data: graphData, error: graphError, isLoading: graphLoading } = useGetTicketGraphQuery(
    { ticket, direction: 'TD' },
    { skip: ticket === '' || tab !== 'graph' },
  )

  const [checkTask, checkTaskState] = useCheckTicketTasksMutation()
  const [addTask, addTaskState] = useAddTicketTaskMutation()
  const [newTaskText, setNewTaskText] = useState('')

  const selectedDocItem = useMemo(() => {
    const list = docsData?.results ?? []
    if (!selectedDoc) return null
    return list.find((d) => d.path === selectedDoc) ?? null
  }, [docsData, selectedDoc])

  const docsByType = useMemo(() => groupByDocType(docsData?.results ?? []), [docsData])
  const docTypeKeys = useMemo(() => Object.keys(docsByType).sort(), [docsByType])

  function setTab(next: TabKey) {
    const sp = new URLSearchParams(searchParams)
    sp.set('tab', next)
    if (next !== 'documents') sp.delete('doc')
    setSearchParams(sp, { replace: true })
  }

  function selectDoc(path: string) {
    const sp = new URLSearchParams(searchParams)
    sp.set('tab', 'documents')
    sp.set('doc', path)
    setSearchParams(sp, { replace: true })
  }

  function clearSelectedDoc() {
    const sp = new URLSearchParams(searchParams)
    sp.delete('doc')
    setSearchParams(sp, { replace: true })
  }

  return (
    <div className="container py-4">
      <div className="d-flex justify-content-between align-items-start gap-2 mb-3">
        <div className="flex-grow-1">
          <div className="h4 mb-0">
            Ticket: <span className="font-monospace">{ticket || '—'}</span>
          </div>
          {t?.title ? <div className="text-muted">{t.title}</div> : null}
          {t?.ticketDir ? <div className="small text-muted font-monospace">{t.ticketDir}</div> : null}
        </div>
        <div className="d-flex gap-2">
          <Link className="btn btn-outline-primary" to="/">
            Search
          </Link>
        </div>
      </div>

      {ticket === '' ? <div className="alert alert-info">Missing ticket id.</div> : null}

      {ticketError ? (
        <div className="alert alert-danger">
          Failed to load ticket: {apiErrorMessage(ticketError)}
        </div>
      ) : null}
      {ticketLoading ? (
        <div className="text-center my-4">
          <div className="spinner-border text-primary" role="status" />
        </div>
      ) : null}

      <div className="d-flex flex-wrap gap-2 mb-3">
        <button
          className={`btn btn-sm ${tab === 'overview' ? 'btn-primary' : 'btn-outline-primary'}`}
          onClick={() => setTab('overview')}
        >
          Overview
        </button>
        <button
          className={`btn btn-sm ${tab === 'documents' ? 'btn-primary' : 'btn-outline-primary'}`}
          onClick={() => setTab('documents')}
        >
          Documents
        </button>
        <button
          className={`btn btn-sm ${tab === 'tasks' ? 'btn-primary' : 'btn-outline-primary'}`}
          onClick={() => setTab('tasks')}
        >
          Tasks
        </button>
        <button
          className={`btn btn-sm ${tab === 'graph' ? 'btn-primary' : 'btn-outline-primary'}`}
          onClick={() => setTab('graph')}
        >
          Graph
        </button>
        <button
          className={`btn btn-sm ${tab === 'changelog' ? 'btn-primary' : 'btn-outline-primary'}`}
          onClick={() => setTab('changelog')}
        >
          Changelog
        </button>
      </div>

      {tab === 'overview' ? (
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
                        <td>{t?.status || '—'}</td>
                        <th className="text-muted">Intent</th>
                        <td>{t?.intent || '—'}</td>
                      </tr>
                      <tr>
                        <th className="text-muted">Created</th>
                        <td>{t?.createdAt || '—'}</td>
                        <th className="text-muted">Updated</th>
                        <td>{formatDate(t?.updatedAt)}</td>
                      </tr>
                      <tr>
                        <th className="text-muted">Topics</th>
                        <td colSpan={3}>{t?.topics?.length ? t.topics.join(', ') : '—'}</td>
                      </tr>
                      <tr>
                        <th className="text-muted">Owners</th>
                        <td colSpan={3}>{t?.owners?.length ? t.owners.join(', ') : '—'}</td>
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
                    Docs: <span className="fw-semibold">{t?.stats?.docsTotal ?? 0}</span>
                  </span>
                  <span className="badge text-bg-light text-dark">
                    Tasks:{' '}
                    <span className="fw-semibold">
                      {t?.stats?.tasksDone ?? 0}/{t?.stats?.tasksTotal ?? 0}
                    </span>
                  </span>
                  <span className="badge text-bg-light text-dark">
                    Files: <span className="fw-semibold">{t?.stats?.relatedFilesTotal ?? 0}</span>
                  </span>
                </div>
                <div className="mt-3 d-flex flex-wrap gap-2">
                  {t?.indexPath ? (
                    <Link className="btn btn-sm btn-outline-primary" to={`/doc?path=${encodeURIComponent(t.indexPath)}`}>
                      Open index.md
                    </Link>
                  ) : null}
                  <button className="btn btn-sm btn-outline-secondary" onClick={() => setTab('documents')}>
                    View documents
                  </button>
                  <button className="btn btn-sm btn-outline-secondary" onClick={() => setTab('tasks')}>
                    View tasks
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      ) : null}

      {tab === 'documents' ? (
        <div className="row g-3">
          <div className={selectedDocItem ? 'col-12 col-lg-7' : 'col-12'}>
            {docsError ? (
              <div className="alert alert-danger">
                Failed to load docs: {apiErrorMessage(docsError)}
              </div>
            ) : null}
            {docsLoading ? (
              <div className="text-center my-4">
                <div className="spinner-border text-primary" role="status" />
              </div>
            ) : null}
            {!docsLoading && docsData ? (
              <div className="vstack gap-3">
                {docTypeKeys.map((k) => (
                  <div key={k} className="card">
                    <div className="card-header fw-semibold d-flex justify-content-between align-items-center">
                      <span className="text-uppercase">{k}</span>
                      <span className="text-muted small">{docsByType[k].length}</span>
                    </div>
                    <div className="card-body">
                      <div className="vstack gap-2">
                        {docsByType[k].map((d) => (
                          <div
                            key={d.path}
                            role="button"
                            tabIndex={0}
                            onClick={() => selectDoc(d.path)}
                            className={`p-2 border rounded ${selectedDoc === d.path ? 'border-primary' : ''}`}
                          >
                            <div className="d-flex justify-content-between align-items-start gap-2">
                              <div className="flex-grow-1">
                                <div className="fw-semibold">{d.title || d.path}</div>
                                <div className="small text-muted">
                                  {d.status ? (
                                    <span className="me-2">
                                      <span className="badge text-bg-primary">{d.status}</span>
                                    </span>
                                  ) : null}
                                  <span className="font-monospace">{d.path}</span>
                                </div>
                                {d.summary ? <div className="small text-muted mt-1">{d.summary}</div> : null}
                              </div>
                              <div className="d-flex gap-2">
                                <Link
                                  className="btn btn-sm btn-outline-primary"
                                  to={`/doc?path=${encodeURIComponent(d.path)}`}
                                  onClick={(e) => e.stopPropagation()}
                                >
                                  Open
                                </Link>
                              </div>
                            </div>
                          </div>
                        ))}
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            ) : null}
          </div>

          {selectedDocItem ? (
            <div className="col-12 col-lg-5">
              <div className="card">
                <div className="card-header d-flex justify-content-between align-items-center">
                  <span className="fw-semibold">Preview</span>
                  <button className="btn btn-sm btn-outline-secondary" onClick={clearSelectedDoc}>
                    Close
                  </button>
                </div>
                <div className="card-body">
                  <div className="fw-semibold mb-1">{selectedDocItem.title || selectedDocItem.path}</div>
                  <div className="small text-muted font-monospace mb-2">{selectedDocItem.path}</div>
                  <div className="d-flex flex-wrap gap-2 mb-2">
                    <span className="badge text-bg-light text-dark">{selectedDocItem.docType}</span>
                    {selectedDocItem.status ? (
                      <span className="badge text-bg-primary">{selectedDocItem.status}</span>
                    ) : null}
                    {selectedDocItem.lastUpdated ? (
                      <span className="badge text-bg-light text-dark">
                        Updated: {formatDate(selectedDocItem.lastUpdated)}
                      </span>
                    ) : null}
                  </div>
                  {selectedDocItem.summary ? <div className="text-muted small mb-3">{selectedDocItem.summary}</div> : null}

                  {selectedDocItem.relatedFiles?.length ? (
                    <div>
                      <div className="fw-semibold mb-2">Related files</div>
                      <ul className="list-unstyled mb-0 vstack gap-2">
                        {selectedDocItem.relatedFiles.slice(0, 12).map((rf) => (
                          <li key={`${rf.path}:${rf.note ?? ''}`}>
                            <div className="font-monospace">{rf.path}</div>
                            {rf.note ? <div className="small text-muted">{rf.note}</div> : null}
                            <div className="mt-1">
                              <Link
                                className="btn btn-sm btn-outline-primary"
                                to={`/file?root=repo&path=${encodeURIComponent(rf.path)}`}
                              >
                                Open file
                              </Link>
                            </div>
                          </li>
                        ))}
                        {selectedDocItem.relatedFiles.length > 12 ? (
                          <li className="text-muted small">… {selectedDocItem.relatedFiles.length - 12} more</li>
                        ) : null}
                      </ul>
                    </div>
                  ) : (
                    <div className="text-muted small">No related files.</div>
                  )}
                </div>
              </div>
            </div>
          ) : null}
        </div>
      ) : null}

      {tab === 'tasks' ? (
        <div className="row g-3">
          <div className="col-12 col-lg-7">
            {tasksError ? (
              <div className="alert alert-danger">
                Failed to load tasks: {apiErrorMessage(tasksError)}
              </div>
            ) : null}
            {tasksLoading ? (
              <div className="text-center my-4">
                <div className="spinner-border text-primary" role="status" />
              </div>
            ) : null}
            {!tasksLoading && tasksData ? (
              <div className="vstack gap-3">
                <div className="card">
                  <div className="card-body d-flex flex-wrap gap-2 justify-content-between align-items-center">
                    <div>
                      <span className="fw-semibold">Progress:</span>{' '}
                      {tasksData.stats.done}/{tasksData.stats.total}
                      {!tasksData.exists ? <span className="text-muted ms-2">(no tasks.md)</span> : null}
                    </div>
                    {tasksData.tasksPath ? (
                      <Link className="btn btn-sm btn-outline-secondary" to={`/doc?path=${encodeURIComponent(tasksData.tasksPath)}`}>
                        Open tasks.md
                      </Link>
                    ) : null}
                  </div>
                </div>

                {tasksData.sections.map((sec) => (
                  <div key={sec.title} className="card">
                    <div className="card-header fw-semibold">{sec.title}</div>
                    <div className="card-body">
                      {sec.items.length === 0 ? (
                        <div className="text-muted small">No tasks in this section.</div>
                      ) : (
                        <div className="vstack gap-2">
                          {sec.items.map((it) => (
                            <label key={it.id} className="d-flex gap-2 align-items-start">
                              <input
                                type="checkbox"
                                checked={it.checked}
                                onChange={(e) => void checkTask({ ticket, ids: [it.id], checked: e.target.checked })}
                                disabled={checkTaskState.isLoading}
                              />
                              <span>
                                <span className="text-muted me-2">#{it.id}</span>
                                {it.text}
                              </span>
                            </label>
                          ))}
                        </div>
                      )}
                    </div>
                  </div>
                ))}
              </div>
            ) : null}
          </div>

          <div className="col-12 col-lg-5">
            <div className="card">
              <div className="card-header fw-semibold">Add task</div>
              <div className="card-body">
                <div className="mb-2 text-muted small">Adds to section “TODO”.</div>
                <div className="input-group">
                  <input
                    className="form-control"
                    value={newTaskText}
                    onChange={(e) => setNewTaskText(e.target.value)}
                    placeholder="New task…"
                  />
                  <button
                    className="btn btn-primary"
                    disabled={addTaskState.isLoading || newTaskText.trim() === ''}
                    onClick={() => {
                      const text = newTaskText.trim()
                      if (!text) return
                      void addTask({ ticket, section: 'TODO', text }).then(() => setNewTaskText(''))
                    }}
                  >
                    Add
                  </button>
                </div>
                {addTaskState.error ? (
                  <div className="alert alert-danger mt-3 py-2">
                    Add failed: {apiErrorMessage(addTaskState.error)}
                  </div>
                ) : null}
              </div>
            </div>
          </div>
        </div>
      ) : null}

      {tab === 'graph' ? (
        <div className="card">
          <div className="card-header fw-semibold d-flex justify-content-between align-items-center">
            <span>Graph</span>
            <span className="text-muted small">
              {graphData ? `${graphData.stats.nodes} nodes • ${graphData.stats.edges} edges` : ''}
            </span>
          </div>
          <div className="card-body">
            {graphError ? (
              <div className="alert alert-danger">
                Failed to load graph: {apiErrorMessage(graphError)}
              </div>
            ) : null}
            {graphLoading ? (
              <div className="text-center my-4">
                <div className="spinner-border text-primary" role="status" />
              </div>
            ) : null}
            {graphData ? (
              <div>
                <div className="text-muted small mb-2">
                  Mermaid rendering is coming next; for now this shows the Mermaid DSL.
                </div>
                <pre className="bg-light p-2 rounded small overflow-auto" style={{ maxHeight: 500 }}>
                  {graphData.mermaid}
                </pre>
              </div>
            ) : null}
          </div>
        </div>
      ) : null}

      {tab === 'changelog' ? (
        <div className="card">
          <div className="card-header fw-semibold">Changelog</div>
          <div className="card-body">
            {t?.ticketDir ? (
              <Link
                className="btn btn-outline-primary"
                to={`/doc?path=${encodeURIComponent(`${t.ticketDir}/changelog.md`)}`}
              >
                Open changelog.md
              </Link>
            ) : (
              <div className="text-muted">Missing ticketDir.</div>
            )}
          </div>
        </div>
      ) : null}
    </div>
  )
}
