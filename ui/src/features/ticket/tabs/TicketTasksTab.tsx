import { useState } from 'react'
import { Link } from 'react-router-dom'

import { ApiErrorAlert } from '../../../components/ApiErrorAlert'
import { LoadingSpinner } from '../../../components/LoadingSpinner'
import type { TicketTasksResponse } from '../../../services/docmgrApi'

function asArray<T>(v: T[] | null | undefined): T[] {
  return Array.isArray(v) ? v : []
}

export function TicketTasksTab({
  ticket,
  tasksData,
  tasksError,
  tasksLoading,
  checkTask,
  checkTaskLoading,
  addTask,
  addTaskLoading,
  addTaskError,
}: {
  ticket: string
  tasksData?: TicketTasksResponse
  tasksError?: unknown
  tasksLoading: boolean
  checkTask: (args: { ticket: string; ids: number[]; checked: boolean }) => Promise<unknown>
  checkTaskLoading: boolean
  addTask: (args: { ticket: string; section: string; text: string }) => Promise<unknown>
  addTaskLoading: boolean
  addTaskError?: unknown
}) {
  const [newTaskText, setNewTaskText] = useState('')

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
                      {asArray(sec.items).map((it) => (
                        <label key={it.id} className="d-flex gap-2 align-items-start">
                          <input
                            type="checkbox"
                            checked={it.checked}
                            onChange={(e) =>
                              void checkTask({ ticket, ids: [it.id], checked: e.target.checked })
                            }
                            disabled={checkTaskLoading}
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
                disabled={addTaskLoading || newTaskText.trim() === ''}
                onClick={() => {
                  const text = newTaskText.trim()
                  if (!text) return
                  void addTask({ ticket, section: 'TODO', text }).then(() => setNewTaskText(''))
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

