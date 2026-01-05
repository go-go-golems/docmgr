export type TicketTabKey = 'overview' | 'documents' | 'tasks' | 'graph' | 'changelog'

export function TicketTabs({
  tab,
  onTabChange,
}: {
  tab: TicketTabKey
  onTabChange: (tab: TicketTabKey) => void
}) {
  return (
    <div className="d-flex flex-wrap gap-2 mb-3">
      <button
        className={`btn btn-sm ${tab === 'overview' ? 'btn-primary' : 'btn-outline-primary'}`}
        onClick={() => onTabChange('overview')}
      >
        Overview
      </button>
      <button
        className={`btn btn-sm ${tab === 'documents' ? 'btn-primary' : 'btn-outline-primary'}`}
        onClick={() => onTabChange('documents')}
      >
        Documents
      </button>
      <button
        className={`btn btn-sm ${tab === 'tasks' ? 'btn-primary' : 'btn-outline-primary'}`}
        onClick={() => onTabChange('tasks')}
      >
        Tasks
      </button>
      <button
        className={`btn btn-sm ${tab === 'graph' ? 'btn-primary' : 'btn-outline-primary'}`}
        onClick={() => onTabChange('graph')}
      >
        Graph
      </button>
      <button
        className={`btn btn-sm ${tab === 'changelog' ? 'btn-primary' : 'btn-outline-primary'}`}
        onClick={() => onTabChange('changelog')}
      >
        Changelog
      </button>
    </div>
  )
}

