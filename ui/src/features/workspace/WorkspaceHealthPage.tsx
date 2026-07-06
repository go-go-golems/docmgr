import { useMemo } from 'react'
import { Link, useSearchParams } from 'react-router-dom'

import { ApiErrorAlert } from '../../components/ApiErrorAlert'
import { EmptyState } from '../../components/EmptyState'
import { LoadingSpinner } from '../../components/LoadingSpinner'
import { DiagnosticList } from '../search/components/DiagnosticList'
import type { DiagnosticTaxonomy, DoctorFinding } from '../../services/docmgrApi'
import { useGetWorkspaceDoctorQuery } from '../../services/docmgrApi'

function findingToTaxonomy(f: DoctorFinding): DiagnosticTaxonomy {
  return {
    Tool: 'doctor',
    Stage: f.issue,
    Symptom: f.message,
    Path: f.path,
    Severity: f.severity,
    Context: f.ticket ? { Ticket: f.ticket } : undefined,
  }
}

function statusBadgeVariant(status: string): string {
  switch (status) {
    case 'error':
      return 'danger'
    case 'warning':
      return 'warning'
    default:
      return 'success'
  }
}

export function WorkspaceHealthPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const ticket = (searchParams.get('ticket') ?? '').trim()

  const { data, error, isLoading, isFetching } = useGetWorkspaceDoctorQuery({ ticket: ticket || undefined })

  const diagnostics = useMemo(
    () => (data?.findings ?? []).filter((f) => f.severity !== 'ok').map(findingToTaxonomy),
    [data],
  )

  function selectTicket(next: string) {
    const sp = new URLSearchParams(searchParams)
    if (next) sp.set('ticket', next)
    else sp.delete('ticket')
    setSearchParams(sp, { replace: true })
  }

  return (
    <div className="vstack gap-3">
      <div className="card">
        <div className="card-header fw-semibold d-flex justify-content-between align-items-center">
          <span>
            Health{ticket ? <span className="font-monospace ms-2">{ticket}</span> : null}
            {isFetching ? <span className="text-muted small ms-2">checking…</span> : null}
          </span>
          {ticket ? (
            <button className="btn btn-sm btn-outline-secondary" onClick={() => selectTicket('')}>
              All tickets
            </button>
          ) : null}
        </div>
        <div className="card-body">
          {error ? <ApiErrorAlert title="Doctor run failed" error={error} /> : null}
          {isLoading ? <LoadingSpinner /> : null}

          {data ? (
            <div className="d-flex flex-wrap gap-2">
              <span className="badge text-bg-light text-dark">
                Tickets checked: <span className="fw-semibold">{data.totals.ticketsChecked}</span>
              </span>
              <span className="badge text-bg-danger">
                Errors: <span className="fw-semibold">{data.totals.errors}</span>
              </span>
              <span className="badge text-bg-warning text-dark">
                Warnings: <span className="fw-semibold">{data.totals.warnings}</span>
              </span>
              <span className="badge text-bg-info text-dark">
                Info: <span className="fw-semibold">{data.totals.infos}</span>
              </span>
            </div>
          ) : null}
        </div>
      </div>

      {data && !ticket ? (
        <div className="card">
          <div className="card-header fw-semibold">Per-ticket rollup</div>
          <div className="card-body p-0">
            <div className="table-responsive">
              <table className="table table-sm table-hover mb-0 align-middle">
                <thead>
                  <tr>
                    <th>Ticket</th>
                    <th>Status</th>
                    <th className="text-end">Errors</th>
                    <th className="text-end">Warnings</th>
                    <th></th>
                  </tr>
                </thead>
                <tbody>
                  {data.rollup.map((item) => (
                    <tr key={item.ticket || '(workspace)'}>
                      <td className="font-monospace">
                        {item.ticket ? (
                          <Link to={`/ticket/${encodeURIComponent(item.ticket)}`} className="text-decoration-none">
                            {item.ticket}
                          </Link>
                        ) : (
                          <span className="text-muted">(workspace)</span>
                        )}
                      </td>
                      <td>
                        <span className={`badge text-bg-${statusBadgeVariant(item.status)}`}>{item.status}</span>
                      </td>
                      <td className="text-end">{item.errors}</td>
                      <td className="text-end">{item.warnings}</td>
                      <td className="text-end">
                        {item.ticket ? (
                          <button
                            className="btn btn-sm btn-outline-primary"
                            onClick={() => selectTicket(item.ticket)}
                          >
                            Details
                          </button>
                        ) : null}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </div>
      ) : null}

      {data ? (
        <div className="card">
          <div className="card-header fw-semibold">Findings</div>
          <div className="card-body">
            {diagnostics.length === 0 ? (
              <EmptyState title="All checks passed">No doctor findings for this scope.</EmptyState>
            ) : (
              <DiagnosticList diagnostics={diagnostics} />
            )}
          </div>
        </div>
      ) : null}
    </div>
  )
}
