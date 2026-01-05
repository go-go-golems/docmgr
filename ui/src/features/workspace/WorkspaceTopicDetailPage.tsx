import { useParams } from 'react-router-dom'

import { EmptyState } from '../../components/EmptyState'

export function WorkspaceTopicDetailPage() {
  const params = useParams()
  const topic = (params.topic ?? '').trim()

  return (
    <div className="card">
      <div className="card-header fw-semibold">Topic: {topic || '—'}</div>
      <div className="card-body">
        <EmptyState title="Not implemented yet">
          <p className="mb-0">
            This page needs the topic detail endpoint from `design/03-workspace-rest-api.md`.
          </p>
        </EmptyState>
      </div>
    </div>
  )
}

