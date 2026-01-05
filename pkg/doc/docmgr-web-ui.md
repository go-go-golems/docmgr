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

`docmgr` includes a Search Web UI (React SPA) intended for local developer workflows.

It talks to the same HTTP API as other clients:

- `GET /api/v1/search/docs` (cursor pagination)
- `GET /api/v1/search/files`
- `POST /api/v1/index/refresh`
- `GET /api/v1/workspace/status`

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
make ui-install
make dev-frontend
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
