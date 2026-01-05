import { DocCard } from '../../../components/DocCard'
import { EmptyState } from '../../../components/EmptyState'
import { LoadingSpinner } from '../../../components/LoadingSpinner'

import type { SearchDocResult } from '../../../services/docmgrApi'
import { MarkdownSnippet } from '../components/MarkdownSnippet'

export function SearchDocsResults({
  loading,
  hasSearched,
  docsResults,
  selected,
  onSelectIndex,
  onCopyPath,
  highlightQuery,
  hasMore,
  onLoadMore,
}: {
  loading: boolean
  hasSearched: boolean
  docsResults: SearchDocResult[]
  selected: SearchDocResult | null
  onSelectIndex: (idx: number) => void
  onCopyPath: (path: string) => void
  highlightQuery: string
  hasMore: boolean
  onLoadMore: () => void
}) {
  if (loading) return <LoadingSpinner />

  if (!hasSearched) {
    return (
      <EmptyState title="Search docmgr documentation">
        <p className="text-muted mb-0">Enter a query or use filters to browse documentation.</p>
      </EmptyState>
    )
  }

  if (docsResults.length === 0) {
    return (
      <EmptyState title="No results found">
        <p className="mb-0">Try adjusting your query or filters.</p>
      </EmptyState>
    )
  }

  return (
    <>
      {docsResults.map((r, idx) => (
        <DocCard
          key={`${r.path}:${r.ticket}`}
          title={r.title}
          ticket={r.ticket}
          docType={r.docType}
          status={r.status}
          topics={r.topics}
          path={r.path}
          lastUpdated={r.lastUpdated}
          relatedFiles={r.relatedFiles}
          snippet={<MarkdownSnippet markdown={r.snippet} query={highlightQuery} />}
          selected={selected?.path === r.path && selected?.ticket === r.ticket}
          onCopyPath={onCopyPath}
          onSelect={() => onSelectIndex(idx)}
        />
      ))}

      {hasMore ? (
        <div className="text-center mt-3">
          <button className="btn btn-outline-primary" onClick={onLoadMore}>
            Load more
          </button>
        </div>
      ) : null}
    </>
  )
}

