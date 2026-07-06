---
Title: docmgr Web UI
Slug: web-ui
Short: Run the docmgr Search Web UI (React SPA) against the local docmgr HTTP API.
Topics:
- docmgr
- ui
- web
- search
IsTemplate: false
IsTopLevel: true
ShowPerDefault: true
SectionType: GeneralTopic
---

# docmgr Web UI

## 1. Overview

`docmgr` includes a web UI (React SPA) intended for local developer workflows: workspace browsing, search, ticket views, and a growing set of write operations.

It talks to the same HTTP API as other clients (see `docmgr help http-api` for the full route list), including:

- `GET /api/v1/search/docs` (cursor pagination) / `GET /api/v1/search/files`
- `GET /api/v1/docs/get` (doc viewer: frontmatter + markdown body)
- `GET /api/v1/files/get` (file viewer: text-only, safe roots) / `GET /api/v1/files/raw` (raw bytes, used for images)
- `POST /api/v1/docs/meta`, `POST /api/v1/docs/relate`, `POST /api/v1/tickets/changelog`, task add/check (write paths)
- `GET /api/v1/workspace/status`, `GET /api/v1/workspace/doctor`
- `POST /api/v1/index/refresh`

UI routes:
- `/workspace` home (also `/workspace/tickets`, `/workspace/topics`, `/workspace/topics/:topic`, `/workspace/recent`)
- `/workspace/health` workspace health page (doctor report)
- `/search` search (docs / reverse lookup / files modes)
- `/ticket/:ticket` ticket detail with tabs: overview, documents, tasks, graph, changelog timeline
- `/doc?path=...` document viewer (markdown rendering)
- `/file?root=repo|docs&path=...` file viewer (syntax highlighted)

Rendering and editing features:
- Markdown bodies render `mermaid` code fences as diagrams, resolve relative links between docs, and load relative images (via the raw-file endpoint).
- Status and Summary can be edited from the document viewer, which also has a relate form for adding `RelatedFiles` entries; changelog entries can be added from the ticket's changelog timeline tab.
- The tasks tab supports checking/adding tasks with a section selector (choose an existing `tasks.md` section or create a new one).

Search URL parameters (useful for sharing links):
- `sel=<docRelPath>` selects a result (opens the desktop preview sidebar)
- `preview=true` opens the preview modal on mobile

## 2. Development mode (recommended)

Two-process loop:

- Frontend (Vite): `http://localhost:3000`
- Backend (docmgr API): `http://127.0.0.1:3001`

The frontend proxies `/api/*` to the backend (no CORS setup needed).

### 2.1. Start backend

```bash
make dev-backend
```

### 2.2. Start frontend

```bash
cd ui
pnpm install
pnpm dev
```

Open:

- `http://localhost:3000`

## 3. Embedded mode (single binary)

In embedded mode, `docmgr api serve` serves both:

- `/api/v1/*` (JSON API)
- `/` + `/assets/*` (SPA + assets, with SPA fallback routing)

### 3.1. Generate embedded assets

```bash
go generate ./internal/web
```

### 3.2. Build with embed tag

```bash
go build -tags "sqlite_fts5,embed" ./cmd/docmgr
```

### 3.3. Run server and open UI

```bash
./docmgr api serve --addr 127.0.0.1:8787 --root ttmp
```

Open:

- `http://127.0.0.1:8787/`

## 4. Troubleshooting

- If `/` returns 404: run `go generate ./internal/web` (dev disk-serving) or build with `-tags embed` (embedded).
- If search returns `fts_not_available`: build/run with `-tags sqlite_fts5`.

## 5. Keyboard shortcuts (MVP)

- `/` focus search
- `?` open shortcuts modal
- `↑/↓` select result
- `Enter` open selected doc
- `Alt+1/2/3` switch Docs/Reverse/Files mode
- `Ctrl/Cmd+R` refresh index
- `Ctrl/Cmd+K` copy selected doc path
