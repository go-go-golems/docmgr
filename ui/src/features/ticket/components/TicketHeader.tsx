import { Link } from 'react-router-dom'

import { PageHeader } from '../../../components/PageHeader'

export function TicketHeader({
  ticket,
  title,
  ticketDir,
}: {
  ticket: string
  title?: string
  ticketDir?: string
}) {
  const subtitle =
    title || ticketDir ? (
      <>
        {title ? <div>{title}</div> : null}
        {ticketDir ? <div className="font-monospace">{ticketDir}</div> : null}
      </>
    ) : undefined

  return (
    <PageHeader
      title={
        <>
          Ticket: <span className="font-monospace">{ticket || '—'}</span>
        </>
      }
      subtitle={subtitle}
      actions={
        <Link className="btn btn-outline-primary" to="/search">
          Search
        </Link>
      }
    />
  )
}
