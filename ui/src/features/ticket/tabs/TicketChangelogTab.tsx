import { Link } from 'react-router-dom'

export function TicketChangelogTab({ ticketDir }: { ticketDir?: string }) {
  return (
    <div className="card">
      <div className="card-header fw-semibold">Changelog</div>
      <div className="card-body">
        {ticketDir ? (
          <Link className="btn btn-outline-primary" to={`/doc?path=${encodeURIComponent(`${ticketDir}/changelog.md`)}`}>
            Open changelog.md
          </Link>
        ) : (
          <div className="text-muted">Missing ticketDir.</div>
        )}
      </div>
    </div>
  )
}

