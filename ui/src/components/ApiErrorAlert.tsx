import { apiErrorFromUnknown } from '../lib/apiError'

export function ApiErrorAlert({
  title,
  error,
}: {
  title: string
  error: unknown
}) {
  const parsed = apiErrorFromUnknown(error)

  return (
    <div className="alert alert-danger">
      <div className="fw-semibold">{title}</div>
      <div className="small">
        {parsed.code ? <span className="me-2">({parsed.code})</span> : null}
        {parsed.message}
      </div>
      {parsed.details ? (
        <details className="mt-2">
          <summary className="small">Details</summary>
          <pre className="small mb-0">{JSON.stringify(parsed.details, null, 2)}</pre>
        </details>
      ) : null}
    </div>
  )
}

