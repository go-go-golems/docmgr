# Agent Guidelines for docmgr

## Build Commands
- Build (CLI only): `go build -tags sqlite_fts5 ./cmd/docmgr` — the `sqlite_fts5` tag enables full-text search; without it `--query` searches error with a rebuild hint.
- Build (full, embedded web UI): `make build` / `make build-embed` (uses `-tags "sqlite_fts5,embed"`). `make ui-build` runs the Dagger + pnpm pipeline that produces the embedded SPA assets (`GOWORK=off go run ./internal/web/generate_build.go`).
- Run without building: `go run -tags sqlite_fts5 ./cmd/docmgr <verb>`.
- Test: `go test ./...` (add `-tags sqlite_fts5` to also exercise FTS paths).
- Run single test: `go test ./pkg/path/to/package -run TestName`
- Generate: `go generate ./...`
- Lint: `golangci-lint run -v` or `make lint`
- Format: `go fmt ./...`

IMPORTANT: To run a server and do some interaction with it, use tmux, this makes it very easy to kill a server.
Use capture-pane to read the output.

## Project Structure
- `cmd/docmgr/`: CLI entry point; command registration lives in `cmd/docmgr/cmds/root.go` with one package per verb group (`ticket/`, `doc/`, `tasks/`, `api/`, ...), each exposing an `Attach(rootCmd)` function.
- `pkg/commands/`: business logic for the CLI verbs (glazed dual-mode commands).
- `pkg/doc/`: embedded help topics (glazed help system; `docmgr help --all`).
- `pkg/models/`, `pkg/diagnostics/`: document model and diagnostics taxonomy/rules.
- `internal/workspace/`: workspace discovery, config resolution, in-memory SQLite index.
- `internal/paths/`: anchored-path parser and the single resolver (`repo://`, `ws://`, `docs://`, `doc://`, `abs://`).
- `internal/httpapi/`: HTTP API server (`/api/v1/*`); `internal/web/`: embedded SPA plumbing.
- `ui/`: web UI sources (pnpm + Vite + React 19 + Redux Toolkit / RTK Query, TypeScript).
- `test-scenarios/testing-doc-manager/`: bash E2E scenario suite.
- `examples/`: example configurations and verb-output templates.
- `ttmp/`: docmgr's own docs workspace, managed by docmgr itself. Layout is `ttmp/YYYY/MM/DD/<TICKET>--<slug>/` with `index.md`, `tasks.md`, `changelog.md`, and per-doc-type subdirectories. Use `docmgr` verbs (ticket create / doc add / doc relate / changelog update / task ...) instead of hand-editing where possible.

## Test Scenarios (E2E)
The scenario suite refuses to run against an ambiguous `docmgr` from PATH. Build a pinned binary and pass it explicitly:

```bash
go build -tags sqlite_fts5 -o /tmp/docmgr-local ./cmd/docmgr
DOCMGR_PATH=/tmp/docmgr-local bash test-scenarios/testing-doc-manager/run-all.sh /tmp/docmgr-scenario
```

<runningProcessesGuidelines>
- When testing TUIs, use tmux and capture-pane to interact with the UI.
- When using tmux, try to batch as many commands as possible when using send-keys.
- When running long-running processes (servers, etc...), use tmux to more easily interact and kill them.
- Kill a process using port $PORT: `lsof-who -p $PORT -k`. When building a web server, ALWAYS use this command to kill the process.
</runningProcessesGuidelines>

<goGuidelines>
- When implementing go interfaces, use the var _ Interface = &Foo{} to make sure the interface is always implemented correctly.
- Always use a context argument when appropriate.
- Use cobra for command-line applications; new docmgr verbs are glazed dual-mode commands (see `docmgr help how-to-add-cli-verbs`).
- Use the "defaults" package name, instead of "default" package name, as it's reserved in go.
- Use github.com/pkg/errors for wrapping errors.
- When starting goroutines, use errgroup.

- Only use the toplevel go.mod, don't create new ones. (Exception that already exists: `scenariolog/` is a separate module and needs `GOWORK=off`.)
- When writing a new experiment / app, add zerolog logging to help debug and figure out how it works, add --log-level flag to set the log level.
- When using go:embed, import embed as `_ "embed"`
- When using build tagged features, make sure the software compiles without the tag as well (docmgr must build without `sqlite_fts5` and without `embed`).
</goGuidelines>

<webGuidelines>
- The web UI lives in `ui/` and uses pnpm + Vite + React 19 + Redux Toolkit (RTK Query) + react-router, in TypeScript. No bootstrap, no templ, no bun.
- Dev loop: `make dev-backend` (API on 127.0.0.1:3001) + `cd ui && pnpm install && pnpm dev` (Vite on localhost:3000, proxies `/api/*`).
- Production build: `make ui-build` runs a Dagger pipeline (pinned pnpm) and copies the bundle into `internal/web` for `go:embed`; then build with `-tags "sqlite_fts5,embed"`.
- API access goes through RTK Query endpoints in `ui/src/services/docmgrApi.ts`; keep new endpoints there.
</webGuidelines>

<debuggingGuidelines>
If me or you the LLM agent seem to go down too deep in a debugging/fixing rabbit hole in our conversations, remind me to take a breath and think about the bigger picture instead of hacking away. Say: "I think I'm stuck, let's TOUCH GRASS".  IMPORTANT: Don't try to fix errors by yourself more than twice in a row. Then STOP. Don't do anything else.

</debuggingGuidelines>

<generalGuidelines>
Don't add backwards compatibility layers or adapters unless explicitly asked. If you think there is a need for a backwards compatibility or adapting to an existing interface, STOP AND ASK ME IF THAT IS NECESSARY. Usually, I don't need backwards compatibility.

If it looks like your edits aren't applied, stop immediately and say "STOPPING BECAUSE EDITING ISN'T WORKING".

Run the format_file tool at the end of each response.
</generalGuidelines>
