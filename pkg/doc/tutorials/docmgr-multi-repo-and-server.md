---
Title: Tutorial — Multi-Repo Setup and Server Usage for docmgr
Slug: docmgr-multi-repo-and-server
Short: Configure the docs root with .ttmp.yaml or DOCMGR_ROOT and use the HTTP server, including /api/status.
Topics:
- docmgr
- setup
- server
- tutorial
IsTemplate: false
IsTopLevel: true
ShowPerDefault: true
SectionType: Tutorial
---

## 1. Overview

This tutorial shows how to set up `docmgr` in a multi-repo workspace and how to run the HTTP server. You will configure the docs root via `.ttmp.yaml` or the `DOCMGR_ROOT` environment variable, verify resolution, and use the `/api/status` endpoint for quick health checks.

## 2. Prerequisites

- `docmgr` installed and on your `PATH`
- A code workspace that contains one or more repositories (multi-repo)

## 3. Configure the Docs Root (Multi-Repo)

In a multi-repo layout, place a `.ttmp.yaml` at the workspace root and point it at the repository-local `ttmp/` directory. Relative paths are resolved relative to the `.ttmp.yaml` file location.

```yaml
root: go-go-mento/ttmp
defaults:
  owners: [manuel]
  intent: long-term
vocabulary: go-go-mento/ttmp/vocabulary.yaml
```

Why this matters: Resolving the root from a single, shared config prevents accidental writes to the wrong `ttmp/` when working across multiple repos or subdirectories.

## 4. Initialize and Verify

Initialize scaffolding (vocabulary, templates, guidelines) if it doesn’t exist yet, then verify the resolved root.

```bash
docmgr init
docmgr status
```

You should see the resolved `root=` pointing at the repository’s `ttmp/` directory.

## 5. Seed Minimal Vocabulary (Optional)

A small vocabulary keeps metadata consistent and improves search/validation. You can evolve it over time.

```bash
docmgr vocab add --category topics   --slug backend --description "Backend services"
docmgr vocab add --category topics   --slug frontend --description "Frontend app"
docmgr vocab add --category docTypes --slug design-doc --description "Design document"
docmgr vocab add --category docTypes --slug reference  --description "API/reference"
docmgr vocab add --category docTypes --slug playbook   --description "Operational playbook"
```

## 6. Create a Ticket Workspace and Add Docs

Create a consistent workspace and add a couple of documents. Use short, descriptive titles.

```bash
docmgr create-ticket --ticket DOC --title "Docmgr — multi-repo setup and server"
docmgr add --ticket DOC --doc-type design-doc --title "Design — Root resolution and server"
docmgr add --ticket DOC --doc-type reference  --title "Reference — API /api/status"
```

Tip: Use `docmgr meta update` to set `Owners`, `Summary`, and `RelatedFiles` on the ticket index to streamline reviews.

## 7. Run the HTTP Server

The server reads `DOCMGR_ROOT` (or discovers `.ttmp.yaml`). Set it to the same root you verified with `docmgr status`.

```bash
DOCMGR_ROOT=go-go-mento/ttmp /tmp/docmgr-server
```

On startup, the server logs the resolved root and config paths. This makes it easy to confirm you’re serving the correct workspace.

## 8. Use /api/status for Health

Query basic health and resolved paths via the status endpoint:

```bash
curl -s http://localhost:8080/api/status | jq .
```

Expected fields include `root`, `configPath`, `vocabularyPath`, and simple `tickets`/`docs` counters.

## 9. Add and List Documents via API (Optional)

You can add documents programmatically. Unknown `docType` values are accepted and stored under `various/`.

```bash
curl -s -X POST http://localhost:8080/api/add \
  -H 'Content-Type: application/json' \
  -d '{
        "ticket":  "DOC",
        "docType": "design-doc",
        "title":   "Design — Workspace conventions"
      }' | jq .

curl -s 'http://localhost:8080/api/list' | jq .
```

## 10. Validate Regularly with doctor

Run `doctor` to catch stale docs, missing required fields, or broken relationships. Start lenient and tighten as your cadence stabilizes.

```bash
docmgr doctor --ignore-dir _templates --ignore-dir _guidelines --stale-after 30 --fail-on error
```

## 11. Writing Style and Structure

When writing documentation, follow these practices:

- Start each section with a short, topic-focused paragraph that explains the core idea, not just the contents.
- Keep examples minimal, runnable, and focused on one concept.
- Use bulleted lists to improve scannability; prefer short paragraphs.
- Use the embedded help system frontmatter fields (`Title`, `Slug`, `Short`, `Topics`, etc.) consistently.

These guidelines are adapted from our broader documentation style guides and help entry conventions.

