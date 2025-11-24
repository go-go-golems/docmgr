# Testing Scenario for docmgr (Documentation Manager)

This directory contains an end-to-end scenario to validate the `docmgr` CLI against a mock repository and a realistic ticket workflow. It creates a temporary repo, scaffolds a ticket workspace under `ttmp/`, adds documents, enriches metadata, runs doctor checks, and exercises search capabilities (including reverse lookup, external sources, and date filters). It also demonstrates managing `tasks.md` and updating `changelog.md` via dedicated verbs.

## Prerequisites

- `docmgr` binary available in PATH or export `DOCMGR_PATH=/absolute/path/to/docmgr`
- `git` installed
- `bash` (POSIX shell compatible)
- Optional for better file suggestions: `rg` (ripgrep) or `grep`

## Contents

- `00-reset.sh` — reset working directory
- `01-create-mock-codebase.sh` — create mock repo `acme-chat-app` with backend/frontend code
- `02-init-ticket.sh` — initialize docs root, seed vocabulary, and create the ticket workspace
- `03-create-docs-and-meta.sh` — add documents and update frontmatter metadata
- `04-relate-and-doctor.sh` — relate code files and run doctor checks
- `05-search-scenarios.sh` — run search scenarios (content, metadata, reverse lookup, external sources, dates, file suggestions)
- `06-doctor-advanced.sh` — exercise doctor warnings, ignore-glob, and recovery
- `07-status.sh` — show workspace summary and staleness
- `13-template-schema-output.sh` — verify --print-template-schema outputs only schema (no human output)
- `run-all.sh` — convenience script to run all steps in order
- `SCENARIO.md` — detailed explanation of the scenario and expected outputs

## Quick Start

```bash
cd vibes/ttmp/YYYY-MM-DD/testing-doc-manager/
chmod +x 00-reset.sh 01-create-mock-codebase.sh 02-init-ticket.sh 03-create-docs-and-meta.sh 04-relate-and-doctor.sh 05-search-scenarios.sh 06-doctor-advanced.sh run-all.sh

# Optionally set the docmgr path if not in PATH
export DOCMGR_PATH=/absolute/path/to/docmgr

# Run the full scenario (uses /tmp/docmgr-scenario by default)
./run-all.sh /tmp/docmgr-scenario

# Run advanced doctor test
./06-doctor-advanced.sh /tmp/docmgr-scenario

# Show status
./07-status.sh /tmp/docmgr-scenario
```

## Notes

- The scenario uses ticket `MEN-4242` with title "Normalize chat API paths and WebSocket lifecycle".
- The ticket workspace will be created under `ttmp/` using the default date-based structure, for example:
  - `ttmp/2025/11/03/MEN-4242-normalize-chat-api-paths-and-websocket-lifecycle/`
- Doctor now supports:
  - `--ignore-dir` and `--ignore-glob` to filter out paths (scenario demonstrates `_templates`, `_guidelines` and design-doc/index.md via `--ignore-glob`)
  - `--stale-after <days>` to tune staleness threshold
  - `--fail-on {none,warning,error}` to fail the command for CI integration
