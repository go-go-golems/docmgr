package httpapi

import (
	"context"
	"sync"
	"time"

	"github.com/go-go-golems/docmgr/internal/workspace"
)

type IndexSnapshot struct {
	Workspace   *workspace.Workspace
	IndexedAt   time.Time
	DocsIndexed int
}

type IndexManager struct {
	rootOverride string

	mu        sync.RWMutex
	workspace *workspace.Workspace
	indexedAt time.Time
	docsCount int
}

func NewIndexManager(rootOverride string) *IndexManager {
	return &IndexManager{rootOverride: rootOverride}
}

func (m *IndexManager) Refresh(ctx context.Context) (IndexSnapshot, error) {
	ws, err := workspace.DiscoverWorkspace(ctx, workspace.DiscoverOptions{RootOverride: m.rootOverride})
	if err != nil {
		return IndexSnapshot{}, err
	}
	if err := ws.InitIndex(ctx, workspace.BuildIndexOptions{IncludeBody: true}); err != nil {
		return IndexSnapshot{}, err
	}

	count, err := countAllDocs(ctx, ws)
	if err != nil {
		return IndexSnapshot{}, err
	}

	now := time.Now()

	m.mu.Lock()
	m.workspace = ws
	m.indexedAt = now
	m.docsCount = count
	m.mu.Unlock()

	return IndexSnapshot{Workspace: ws, IndexedAt: now, DocsIndexed: count}, nil
}

func (m *IndexManager) Snapshot() IndexSnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return IndexSnapshot{
		Workspace:   m.workspace,
		IndexedAt:   m.indexedAt,
		DocsIndexed: m.docsCount,
	}
}

func (m *IndexManager) WithWorkspace(fn func(ws *workspace.Workspace) error) error {
	m.mu.RLock()
	ws := m.workspace
	m.mu.RUnlock()

	if ws == nil {
		return ErrIndexNotReady
	}
	return fn(ws)
}

func countAllDocs(ctx context.Context, ws *workspace.Workspace) (int, error) {
	res, err := ws.QueryDocs(ctx, workspace.DocQuery{
		Scope:   workspace.Scope{Kind: workspace.ScopeRepo},
		Filters: workspace.DocFilters{},
		Options: workspace.DocQueryOptions{
			IncludeBody:         false,
			IncludeErrors:       true,
			IncludeDiagnostics:  false,
			IncludeArchivedPath: true,
			IncludeScriptsPath:  true,
			IncludeControlDocs:  true,
			OrderBy:             workspace.OrderByPath,
			Reverse:             false,
		},
	})
	if err != nil {
		return 0, err
	}
	return len(res.Docs), nil
}
