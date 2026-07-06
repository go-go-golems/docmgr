# Contributing to docmgr

Thank you for your interest in contributing to docmgr! This guide will help you get started with development.

## Development Setup

### Prerequisites

- Go 1.25 or later
- Git
- (Web UI only) pnpm and a Dagger-capable Docker setup, for `make ui-build`

### Getting Started

```bash
# Clone the repository
git clone https://github.com/go-go-golems/docmgr
cd docmgr

# Download dependencies
go mod download

# Build the project (sqlite_fts5 enables full-text search)
go build -tags sqlite_fts5 ./cmd/docmgr

# Run tests
go test ./...

# Install locally (optional; also builds/embeds the web UI)
make install
```

## Architecture

docmgr is built using the [Glazed](https://github.com/go-go-golems/glazed) framework, which provides a consistent CLI structure and output formatting (dual-mode: human text by default, structured output via `--with-glaze-output`).

### Directory Structure

```
docmgr/
├── cmd/docmgr/           # Main CLI entry point
│   ├── main.go           # Help system setup + Execute
│   └── cmds/             # Command registration
│       ├── root.go       # Root command; calls <group>.Attach(rootCmd)
│       ├── ticket/       # ticket create/show/list/rename/close/move/graph
│       ├── doc/          # doc add/list/search/relate/move/...
│       ├── tasks/        # task add/list/check/uncheck/edit/remove/migrate
│       └── ...           # api, changelog, meta, vocab, skill, validate, ...
├── pkg/
│   ├── commands/         # Command implementations (business logic)
│   ├── models/           # Core data structures (Document, Vocabulary, ...)
│   ├── diagnostics/      # Diagnostics taxonomies + rendering rules
│   └── doc/              # Embedded help documentation (docmgr help --all)
├── internal/
│   ├── workspace/        # Workspace discovery, config resolution, SQLite index
│   ├── paths/            # Anchored-path parser + resolver (repo://, ws://, ...)
│   ├── httpapi/          # HTTP API server (/api/v1/*)
│   └── web/              # Embedded SPA build/serve plumbing
├── ui/                   # Web UI (pnpm + Vite + React + RTK Query)
├── test-scenarios/       # Bash E2E scenario suite
└── ttmp/                 # Documentation workspace (for docmgr itself)
```

### Key Components

#### Commands (`pkg/commands/` + `cmd/docmgr/cmds/`)

Each verb is implemented as a Glazed command:

- Business logic lives in `pkg/commands/` (one file per verb)
- Cobra wiring lives in `cmd/docmgr/cmds/<group>/`; each group package exposes an `Attach(rootCmd *cobra.Command) error` function that builds and registers its subcommands
- Groups are registered centrally in `cmd/docmgr/cmds/root.go`
- Commands support both human-friendly output and structured output (JSON/YAML/CSV via `--with-glaze-output`)

#### Data Models (`pkg/models/`)

- **Document**: Core document metadata structure with YAML frontmatter
- **Vocabulary**: Controlled vocabulary for topics, doc types, intents, and statuses
- **RelatedFiles**: List of related code files with notes; paths are persisted as explicit anchors (`repo://`, `ws://`, `docs://`, `abs://` — see `docmgr help path-anchors`)

#### Configuration (`internal/workspace/config.go`)

docmgr resolves its documentation root using this fallback chain (`workspace.ResolveRoot`):

1. `--root` flag (explicit command-line argument; relative paths are anchored on CWD)
2. `.ttmp.yaml` config file — located via the `DOCMGR_CONFIG` environment variable if set, otherwise by walking up the directory tree from CWD; its `root:` value is resolved relative to the config file's directory
3. Git repository root: `<git-root>/ttmp`
4. Default: `ttmp` in the current directory

Set `DOCMGR_DEBUG=1` to trace the resolution. Most commands should not call `ResolveRoot` directly — use `workspace.DiscoverWorkspace()` which wraps it.

## Adding a New Command

The authoritative, up-to-date walkthrough is the embedded help topic:

```bash
docmgr help how-to-add-cli-verbs
```

(Source: `pkg/doc/docmgr-how-to-add-cli-verbs.md`.) It covers the current Glazed APIs — command structs embedding `cmds.NewCommandDescription`, settings structs with `glazed:` tags, dual-mode (`BareCommand`/`GlazeCommand`) output, the `Attach()` registration pattern, and workspace integration. In short:

1. Implement the command in `pkg/commands/<action>.go`
2. Wire it up in `cmd/docmgr/cmds/<group>/<action>.go` and register it in the group's `Attach()`
3. If it is a new group, add a `<group>.Attach(rootCmd)` call in `cmd/docmgr/cmds/root.go`
4. Add tests (`pkg/commands/<action>_test.go`) and, if the command introduces new concepts, a help topic in `pkg/doc/`

## Testing

### Running Tests

```bash
# Run all tests
go test ./...

# Include FTS5-backed search paths
go test -tags sqlite_fts5 ./...

# Run tests with verbose output
go test -v ./...

# Run tests for a specific package
go test ./pkg/commands

# Run tests with coverage
go test -cover ./...
```

### Integration Tests (scenario suite)

docmgr includes end-to-end scenarios in `test-scenarios/testing-doc-manager/`. The runner refuses to fall back to a `docmgr` from PATH — build a pinned binary and pass it via `DOCMGR_PATH`:

```bash
go build -tags sqlite_fts5 -o /tmp/docmgr-local ./cmd/docmgr
DOCMGR_PATH=/tmp/docmgr-local bash test-scenarios/testing-doc-manager/run-all.sh /tmp/docmgr-scenario
```

These scenarios test end-to-end workflows and help ensure commands work correctly together.

## Code Style

### Formatting

- Use `gofmt` for formatting (or `goimports` for import organization)
- Run `golangci-lint` before committing: `make lint`

### Documentation

- **Package-level docs**: Every package should have a package comment explaining its purpose
- **Exported types**: Add godoc comments with examples where helpful
- **Complex functions**: Add comments explaining non-obvious logic
- **Commands**: Use descriptive short/long descriptions and examples in the command description

### Error Handling

- Use `github.com/pkg/errors` for error wrapping: `errors.Wrap(err, "context")`
- Return errors from functions; don't log and continue silently
- Provide helpful error messages that guide users to solutions
- Mutating verbs must exit non-zero on failure (agents and CI rely on exit codes)

### Naming Conventions

- Commands: `New{Action}Command()` (e.g., `NewAddCommand()`)
- Types: Use descriptive names (`WorkspaceConfig`, not `TTMPConfig`)
- Functions: Self-documenting names (`ResolveRoot()`, not `Resolve()`)

## Common Patterns

### Discovering the Workspace

```go
ws, err := workspace.DiscoverWorkspace(ctx, workspace.DiscoverOptions{
    RootOverride: settings.Root,
})
if err != nil {
    return errors.Wrap(err, "failed to discover workspace")
}
// ws.Context().Root is the resolved absolute docs root
```

### Reading Configuration

```go
cfg, err := workspace.LoadWorkspaceConfig() // nil cfg when no .ttmp.yaml exists
```

### Working with Documents

```go
doc := &models.Document{
    Title:   "My Document",
    Ticket:  "MEN-1234",
    DocType: "design-doc",
    Topics:  []string{"api", "backend"},
}
```

## Getting Help

- **CLI Help**: `docmgr help` or `docmgr help how-to-use`
- **Embedded Docs**: `docmgr help --all` lists all available help topics
- **Adding verbs**: `docmgr help how-to-add-cli-verbs`
- **Issues**: Open an issue on GitHub for questions or bug reports

## Submitting Changes

1. **Fork the repository** and create a feature branch
2. **Write tests** for your changes
3. **Run tests and linting**: `make test && make lint`
4. **Update documentation** as needed (embedded help in `pkg/doc/` compiles into the binary)
5. **Commit** with clear, descriptive messages
6. **Push** to your fork and open a pull request

## Code Review Process

- All PRs require review before merging
- Ensure tests pass and code is properly formatted
- Address review feedback promptly
- Keep PRs focused and reasonably sized

## Questions?

If you have questions about contributing, feel free to:
- Open an issue on GitHub
- Check existing documentation: `docmgr help`
- Review the codebase structure and existing command implementations

Thank you for contributing to docmgr!
