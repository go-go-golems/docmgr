import { Link } from 'react-router-dom'

import { ApiErrorAlert } from '../../../components/ApiErrorAlert'
import { EmptyState } from '../../../components/EmptyState'
import { LoadingSpinner } from '../../../components/LoadingSpinner'
import { PathHeader } from '../../../components/PathHeader'
import { RelatedFilesList } from '../../../components/RelatedFilesList'
import { DocCard } from '../../../components/DocCard'
import type { TicketDocItem } from '../../../services/docmgrApi'

export function TicketDocumentsTab({
  ticket,
  docsError,
  docsLoading,
  docTypeKeys,
  docsByType,
  selectedDoc,
  selectedDocItem,
  onSelectDoc,
  onClearSelectedDoc,
  onCopyPath,
  formatDate,
}: {
  ticket: string
  docsError?: unknown
  docsLoading: boolean
  docTypeKeys: string[]
  docsByType: Record<string, TicketDocItem[]>
  selectedDoc: string
  selectedDocItem: TicketDocItem | null
  onSelectDoc: (path: string) => void
  onClearSelectedDoc: () => void
  onCopyPath: (path: string) => void
  formatDate: (iso?: string) => string
}) {
  return (
    <div className="row g-3">
      <div className={selectedDocItem ? 'col-12 col-lg-7' : 'col-12'}>
        {docsError ? <ApiErrorAlert title="Failed to load docs" error={docsError} /> : null}
        {docsLoading ? <LoadingSpinner /> : null}
        {!docsLoading && docTypeKeys.length === 0 ? <EmptyState title="No documents found" /> : null}
        {!docsLoading && docTypeKeys.length > 0 ? (
          <div className="vstack gap-3">
            {docTypeKeys.map((k) => (
              <div key={k} className="card">
                <div className="card-header fw-semibold d-flex justify-content-between align-items-center">
                  <span className="text-uppercase">{k}</span>
                  <span className="text-muted small">{docsByType[k].length}</span>
                </div>
                <div className="card-body">
                  <div className="vstack gap-2">
                    {docsByType[k].map((d) => (
                      <DocCard
                        key={d.path}
                        title={d.title || d.path}
                        ticket={ticket}
                        docType={d.docType}
                        status={d.status}
                        topics={d.topics}
                        path={d.path}
                        lastUpdated={d.lastUpdated}
                        relatedFiles={d.relatedFiles}
                        selected={selectedDoc === d.path}
                        snippet={d.summary ? <span className="text-muted">{d.summary}</span> : null}
                        onSelect={() => onSelectDoc(d.path)}
                        actions={
                          <>
                            <Link
                              className="btn btn-sm btn-outline-primary"
                              to={`/doc?path=${encodeURIComponent(d.path)}`}
                              onClick={(e) => e.stopPropagation()}
                            >
                              Open
                            </Link>
                            <button
                              className="btn btn-sm btn-outline-secondary"
                              onClick={(e) => {
                                e.stopPropagation()
                                onCopyPath(d.path)
                              }}
                            >
                              Copy
                            </button>
                          </>
                        }
                      />
                    ))}
                  </div>
                </div>
              </div>
            ))}
          </div>
        ) : null}
      </div>

      {selectedDocItem ? (
        <div className="col-12 col-lg-5">
          <div className="card">
            <div className="card-header d-flex justify-content-between align-items-center">
              <span className="fw-semibold">Preview</span>
              <button className="btn btn-sm btn-outline-secondary" onClick={onClearSelectedDoc}>
                Close
              </button>
            </div>
            <div className="card-body">
              <div className="fw-semibold mb-1">{selectedDocItem.title || selectedDocItem.path}</div>
              <div className="small text-muted font-monospace mb-2">{selectedDocItem.path}</div>
              <div className="d-flex flex-wrap gap-2 mb-2">
                <span className="badge text-bg-light text-dark">{selectedDocItem.docType}</span>
                {selectedDocItem.status ? <span className="badge text-bg-primary">{selectedDocItem.status}</span> : null}
                {selectedDocItem.lastUpdated ? (
                  <span className="badge text-bg-light text-dark">Updated: {formatDate(selectedDocItem.lastUpdated)}</span>
                ) : null}
              </div>
              {selectedDocItem.summary ? <div className="text-muted small mb-3">{selectedDocItem.summary}</div> : null}

              <PathHeader
                path={selectedDocItem.path}
                actions={
                  <>
                    <button className="btn btn-sm btn-outline-primary" onClick={() => onCopyPath(selectedDocItem.path)}>
                      Copy path
                    </button>
                    <Link className="btn btn-sm btn-primary" to={`/doc?path=${encodeURIComponent(selectedDocItem.path)}`}>
                      Open doc
                    </Link>
                  </>
                }
              />

              {selectedDocItem.relatedFiles?.length ? (
                <div>
                  <div className="fw-semibold mb-2">Related files</div>
                  <RelatedFilesList
                    files={selectedDocItem.relatedFiles.slice(0, 12)}
                    onCopyPath={onCopyPath}
                    showCopy={false}
                    openLabel="Open file"
                  />
                  {selectedDocItem.relatedFiles.length > 12 ? (
                    <div className="text-muted small mt-2">
                      … {selectedDocItem.relatedFiles.length - 12} more
                    </div>
                  ) : null}
                </div>
              ) : (
                <div className="text-muted small">No related files.</div>
              )}
            </div>
          </div>
        </div>
      ) : null}
    </div>
  )
}

