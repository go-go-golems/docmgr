package workspace

import (
	"context"
	"database/sql"
	"path/filepath"

	"github.com/go-go-golems/docmgr/internal/paths"
	"github.com/pkg/errors"
)

// Workspace is the centralized repository lookup entry point.
//
// It owns:
// - discovery (Root/ConfigDir/RepoRoot + config best-effort load),
// - path normalization via internal/paths.Resolver,
// - (later) an in-memory SQLite index for querying documents and reverse lookups.
type Workspace struct {
	ctx      WorkspaceContext
	resolver *paths.Resolver
	db       *sql.DB
}

// WorkspaceContext captures the resolved "environment" for a workspace instance.
//
// Spec: §5.1 Construction.
type WorkspaceContext struct {
	Root      string
	ConfigDir string
	RepoRoot  string
	Config    *WorkspaceConfig // best-effort loaded config (may be nil)
}

// DiscoverOptions customizes workspace discovery.
type DiscoverOptions struct {
	// RootOverride overrides the default docs root (typically "ttmp").
	// If empty, discovery uses the default resolution chain (config/git/cwd).
	RootOverride string
}

// DiscoverWorkspace resolves a best-effort WorkspaceContext from process state and options.
//
// Spec: §5.1 Construction. Failure / fallback semantics are an open design note (see spec).
func DiscoverWorkspace(ctx context.Context, opts DiscoverOptions) (*Workspace, error) {
	_ = ctx // reserved for future cancellation / IO-bound discovery work

	root := opts.RootOverride
	if root == "" {
		root = "ttmp"
	}
	root = ResolveRoot(root)
	if root == "" {
		return nil, errors.New("failed to resolve docs root")
	}

	// Best-effort config load. LoadWorkspaceConfig currently returns an error for malformed
	// config even though it prints a warning; we treat this as non-fatal here and keep
	// Config=nil so commands can proceed.
	cfg, cfgErr := LoadWorkspaceConfig()
	if cfgErr != nil {
		VerboseLog("warning: failed to load workspace config (continuing): %v", cfgErr)
		cfg = nil
	}

	// Best-effort config directory discovery.
	configDir := ""
	if cfgPath, err := FindTTMPConfigPath(); err == nil && cfgPath != "" {
		configDir = filepath.Dir(cfgPath)
	} else {
		// Heuristic: if root is ".../ttmp", config lives at its parent (repo root in most setups).
		configDir = filepath.Dir(root)
	}
	if configDir == "" {
		return nil, errors.New("failed to resolve config dir")
	}

	repoRoot, err := FindRepositoryRoot()
	if err != nil || repoRoot == "" {
		return nil, errors.Wrap(err, "failed to resolve repository root")
	}

	return NewWorkspaceFromContext(WorkspaceContext{
		Root:      root,
		ConfigDir: configDir,
		RepoRoot:  repoRoot,
		Config:    cfg,
	})
}

// NewWorkspaceFromContext constructs a Workspace from an explicit context.
//
// This is primarily intended for tests; CLI code should typically call DiscoverWorkspace.
//
// Spec: §5.1 Construction.
func NewWorkspaceFromContext(ctx WorkspaceContext) (*Workspace, error) {
	if ctx.Root == "" {
		return nil, errors.New("workspace context missing Root")
	}
	if ctx.ConfigDir == "" {
		return nil, errors.New("workspace context missing ConfigDir")
	}
	if ctx.RepoRoot == "" {
		return nil, errors.New("workspace context missing RepoRoot")
	}

	resolver := paths.NewResolver(paths.ResolverOptions{
		DocsRoot:  ctx.Root,
		ConfigDir: ctx.ConfigDir,
		RepoRoot:  ctx.RepoRoot,
	})

	return &Workspace{
		ctx:      ctx,
		resolver: resolver,
	}, nil
}

// Context returns the workspace context used for construction.
func (w *Workspace) Context() WorkspaceContext {
	return w.ctx
}

// Resolver returns the resolver used for normalizing paths.
func (w *Workspace) Resolver() *paths.Resolver {
	return w.resolver
}

// DB returns the in-memory SQLite database backing this workspace (if initialized).
func (w *Workspace) DB() *sql.DB {
	return w.db
}
