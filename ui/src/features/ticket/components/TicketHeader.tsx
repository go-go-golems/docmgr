import { Link } from 'react-router-dom'

export function TicketHeader({
  ticket,
  title,
  ticketDir,
}: {
  ticket: string
  title?: string
  ticketDir?: string
}) {
  return (
    <div className="d-flex justify-content-between align-items-start gap-2 mb-3">
      <div className="flex-grow-1">
        <div className="h4 mb-0">
          Ticket: <span className="font-monospace">{ticket || '—'}</span>
        </div>
        {title ? <div className="text-muted">{title}</div> : null}
        {ticketDir ? <div className="small text-muted font-monospace">{ticketDir}</div> : null}
      </div>
      <div className="d-flex gap-2">
        <Link className="btn btn-outline-primary" to="/">
          Search
        </Link>
      </div>
    </div>
  )
}

