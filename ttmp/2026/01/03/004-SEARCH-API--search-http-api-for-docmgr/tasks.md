# Tasks

## TODO

- [x] Define HTTP surface (`docmgr api serve`) + flags
- [x] Implement `internal/httpapi.IndexManager` (build on startup; refresh on demand)
- [x] Implement HTTP server wiring + JSON error format
- [x] Add endpoints: `GET /api/v1/healthz`, `GET /api/v1/workspace/status`
- [x] Add endpoint: `POST /api/v1/index/refresh`
- [x] Add endpoint: `GET /api/v1/search/docs` (shared engine + cursor pagination)
- [x] Add endpoint: `GET /api/v1/search/files` (optional; maps to `--files`)
- [x] Add smoke tests (cursor + index-not-ready)
- [ ] Update embedded docs / help text for server mode
