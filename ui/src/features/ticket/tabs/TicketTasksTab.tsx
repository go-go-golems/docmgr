import { useMemo, useState } from 'react'
import { Link } from 'react-router-dom'

import { ApiErrorAlert } from '../../../components/ApiErrorAlert'
import { LoadingSpinner } from '../../../components/LoadingSpinner'
import type { TicketTasksResponse } from '../../../services/docmgrApi'

function asArray<T>(v: T[] | null | undefined): T[] {
  return Array.isArray(v) ? v : []
}

const NEW_SECTION = '__new__'

export function TicketTasksTab({
  ticket,
  tasksData,
  tasksError,
  tasksLoading,
  checkTask,
  addTask,
  addTaskLoading,
  addTaskError,
}: {
  ticket: string
  tasksData?: TicketTasksResponse
  tasksError?: unknown
  tasksLoading: boolean
  checkTask: (args: { ticket: string; refs: string[]; checked: boolean }) => Promise<unknown>
  addTask: (args: { ticket: string; section: string; text: string }) => Promise<unknown>
  addTaskLoading: boolean
  addTaskError?: unknown
}) {
  const [newTaskText, setNewTaskText] = useState('')
  const [section, setSection] = useState('TODO')
  const [customSection, setCustomSection] = useState('')

  const sectionTitles = useMemo(() => {
    const titles = (tasksData?.sections ?? []).map((s) => s.title).filter((t) => t.trim() !== '')
    if (!titles.some((t) => t.toLowerCase() === 'todo')) titles.unshift('TODO')
    return titles
  }, [tasksData])

  const effectiveSection = section === NEW_SECTION ? customSection.trim() : section

  return (
    <div className="row g-3">
      <div className="col-12 col-lg-7">
        {tasksError ? <ApiErrorAlert title="Failed to load tasks" error={tasksError} /> : null}
        {tasksLoading ? <LoadingSpinner /> : null}
        {!tasksLoading && tasksData ? (
          <div className="vstack gap-3">
            <div className="card">
              <div className="card-body d-flex flex-wrap gap-2 justify-content-between align-items-center">
                <div>
                  <span className="fw-semibold">Progress:</span> {tasksData.stats.done}/{tasksData.stats.total}
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
                  {asArray(sec.items).length === 0 ? (
                    <div className="text-muted small">No tasks in this section.</div>
                  ) : (
                    <div className="vstack gap-2">
                      {asArray(sec.items).map((it) => {
                        const taskRef = it.stableId ?? String(it.id)
                        return (
                          <label key={taskRef} className="d-flex gap-2 align-items-start">
                            <input
                              type="checkbox"
                              checked={it.checked}
                              onChange={(e) =>
                                void checkTask({ ticket, refs: [taskRef], checked: e.target.checked })
                              }
                            />
                            <span>
                              <span className="text-muted me-2">#{taskRef}</span>
                              {it.text}
                            </span>
                          </label>
                        )
                      })}
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
          <div className="card-body vstack gap-2">
            <div>
              <label className="form-label small mb-1">Section</label>
              <select
                className="form-select form-select-sm"
                value={section}
                onChange={(e) => setSection(e.target.value)}
              >
                {sectionTitles.map((t) => (
                  <option key={t} value={t}>
                    {t}
                  </option>
                ))}
                <option value={NEW_SECTION}>New section…</option>
              </select>
            </div>
            {section === NEW_SECTION ? (
              <input
                className="form-control form-control-sm"
                value={customSection}
                onChange={(e) => setCustomSection(e.target.value)}
                placeholder="Section title"
              />
            ) : null}
            <div className="input-group">
              <input
                className="form-control"
                value={newTaskText}
                onChange={(e) => setNewTaskText(e.target.value)}
                placeholder="New task…"
              />
              <button
                className="btn btn-primary"
                disabled={addTaskLoading || newTaskText.trim() === '' || effectiveSection === ''}
                onClick={() => {
                  const text = newTaskText.trim()
                  if (!text || effectiveSection === '') return
                  void addTask({ ticket, section: effectiveSection, text }).then(() => setNewTaskText(''))
                }}
              >
                Add
              </button>
            </div>
            {addTaskError ? <ApiErrorAlert title="Add failed" error={addTaskError} /> : null}
          </div>
        </div>
      </div>
    </div>
  )
}
