# docmgr web UI

React 19 + TypeScript + Vite single-page app for browsing (and now editing) a
docmgr workspace. It talks to the JSON API served by `docmgr api serve`
(`internal/httpapi`) and is embedded into the docmgr binary for production via
`go:embed` (`internal/web`).

## Development

Two processes: the Go API backend on port 3001 and the Vite dev server on
port 3000 (which proxies `/api` to 3001, see `vite.config.ts`).

```bash
# 1. From the repo root: run the API against the repo's own ttmp workspace
make dev-backend        # go run -tags sqlite_fts5 ./cmd/docmgr api serve --addr 127.0.0.1:3001 --root ttmp

# 2. From ui/: run the dev server with HMR
pnpm install
pnpm dev                # http://localhost:3000
```

Checks:

```bash
pnpm tsc -b --noEmit    # typecheck
pnpm lint               # eslint
pnpm build              # production build into dist/public
```

## Production embed

The embedded binary is built through the Dagger pipeline (do not run `pnpm
build` by hand for this):

```bash
make build-embed        # ui build via internal/web/generate_build.go + go build -tags "sqlite_fts5,embed"
```

## Layout

- `src/services/docmgrApi.ts` — RTK Query API client (all `/api/v1` endpoints).
- `src/features/` — route-level pages: workspace shell (home/tickets/topics/
  recent/health), search, ticket tabs (overview/documents/tasks/graph/
  changelog), doc viewer, file viewer.
- `src/components/` — shared presentational components (`MarkdownBlock` with
  mermaid + relative link/image handling, `StatusBadge`, `DiagnosticCard`, ...).

See `docmgr help web-ui` and `docmgr help http-api` for the server-side docs.
