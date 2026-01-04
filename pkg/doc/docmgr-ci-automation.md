---
Title: CI and Automation Guide for docmgr
Slug: ci-and-automation
Short: Integrate docmgr validation into CI/CD pipelines with GitHub Actions, GitLab CI, pre-commit hooks, and automation patterns.
Topics:
- docmgr
- ci-cd
- automation
- validation
IsTemplate: false
IsTopLevel: true
ShowPerDefault: true
SectionType: Tutorial
---

# CI and Automation Guide for docmgr

## Overview

Automated validation prevents documentation drift by catching issues—broken links, missing files, unknown topics, stale docs—before they reach production. docmgr's `doctor` command is designed for CI/CD integration with proper exit codes, structured output for reporting, and configurable strictness levels. This guide shows how to integrate docmgr into various CI systems (GitHub Actions, GitLab CI, pre-commit hooks) and provides patterns for bulk operations, monitoring, and automated reporting.

**This guide covers:** CI validation, pre-commit hooks, Makefile integration, automated reporting, bulk operation patterns, and monitoring strategies.

**Prerequisites:**
- docmgr initialized in your repository (see `docmgr help how-to-setup`)
- CI/CD system available (GitHub Actions, GitLab CI, or equivalent)
- Basic understanding of shell scripting for automation patterns

---

## 1. GitHub Actions Integration

GitHub Actions integration provides automated documentation validation on every pull request, preventing broken links and stale docs from being merged. The workflow triggers only on documentation changes to save CI minutes, installs docmgr, and runs `doctor` with configurable strictness. This section provides complete workflow examples from basic validation to advanced reporting with structured output.

### Basic Validation

Add to `.github/workflows/docs-validation.yml`:

```yaml
name: Validate Documentation

on:
  pull_request:
    paths:
      - 'ttmp/**'
      - '.ttmp.yaml'
      - 'pkg/doc/**'

jobs:
  validate-docs:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Install docmgr
        run: go install -tags sqlite_fts5 github.com/go-go-golems/docmgr@latest
      
      - name: Validate documentation
        run: |
          docmgr doctor --all \
            --stale-after 30 \
            --fail-on error
```

**Trigger:** Only runs when docs are changed (saves CI minutes).

---

### Validation with Reporting

```yaml
name: Validate and Report Docs

on:
  pull_request:
    paths:
      - 'ttmp/**'

jobs:
  validate-docs:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Install docmgr
        run: go install -tags sqlite_fts5 github.com/go-go-golems/docmgr@latest
      
      - name: Validate documentation
        run: |
          docmgr doctor --all --stale-after 30 --fail-on error
      
      - name: Generate report on failure
        if: failure()
        run: |
          echo "=== Documentation Issues ==="
          docmgr doctor --all --with-glaze-output --output json | \
            jq -r '.[] | select(.issue != "none") | "[\(.ticket)] \(.path): \(.message)"'
```

---

### Adjusting Strictness Over Time

**Phase-based approach:**

```yaml
# Phase 1: Start lenient (errors only)
- name: Validate (Phase 1)
  run: docmgr doctor --all --fail-on error

# Phase 2: Add staleness (60 days)
- name: Validate (Phase 2)
  run: docmgr doctor --all --stale-after 60 --fail-on error

# Phase 3: Stricter staleness (30 days)
- name: Validate (Phase 3)
  run: docmgr doctor --all --stale-after 30 --fail-on error

# Phase 4: Warnings as errors (strict)
- name: Validate (Phase 4)
  run: docmgr doctor --all --stale-after 30 --fail-on warning
```

Use `.docmgrignore` to suppress false positives rather than lowering standards.

---

## 2. GitLab CI Integration

Add to `.gitlab-ci.yml`:

```yaml
validate-docs:
  stage: test
  image: golang:1.21
  script:
    - go install -tags sqlite_fts5 github.com/go-go-golems/docmgr@latest
    - docmgr doctor --all --stale-after 30 --fail-on error
  only:
    changes:
      - ttmp/**
      - .ttmp.yaml
  allow_failure: false
```

### With Reporting

```yaml
validate-docs:
  stage: test
  image: golang:1.21
  script:
    - go install -tags sqlite_fts5 github.com/go-go-golems/docmgr@latest
    - docmgr doctor --all --stale-after 30 --fail-on error || EXIT_CODE=$?
    - |
      if [ $EXIT_CODE -ne 0 ]; then
        echo "=== Documentation Issues ==="
        docmgr doctor --all --with-glaze-output --output json | \
          jq -r '.[] | select(.issue != "none") | "[\(.ticket)] \(.path): \(.message)"'
        exit $EXIT_CODE
      fi
```

---

## 3. Pre-commit Hook

Validate docs before allowing commits.

### Basic Hook

Create `.git/hooks/pre-commit`:

```bash
#!/bin/bash

# Validate docs before commit
if ! docmgr doctor --all --fail-on error; then
  echo "❌ Documentation validation failed"
  echo ""
  echo "Fix issues or bypass with: git commit --no-verify"
  exit 1
fi

echo "✅ Documentation validation passed"
```

Make it executable:
```bash
chmod +x .git/hooks/pre-commit
```

---

### Hook with Focused Validation

```bash
#!/bin/bash

# Only validate tickets with changed docs
CHANGED_TICKETS=$(git diff --cached --name-only | \
  grep "^ttmp/" | \
  cut -d/ -f2 | \
  sort -u)

if [ -z "$CHANGED_TICKETS" ]; then
  echo "No doc changes, skipping validation"
  exit 0
fi

echo "Validating changed tickets: $CHANGED_TICKETS"

for ticket in $CHANGED_TICKETS; do
  if ! docmgr doctor --ticket "$ticket" --fail-on error; then
    echo "❌ Validation failed for $ticket"
    exit 1
  fi
done

echo "✅ All changed tickets validated"
```

**Benefits:** Only validates tickets you changed, faster feedback.

---

## 4. Makefile Integration

Add docmgr targets to your Makefile:

```makefile
.PHONY: docs-validate docs-status docs-report docs-clean

# Validate all docs
docs-validate:
	@echo "Validating documentation..."
	@docmgr doctor --all --stale-after 30 --fail-on error

# Show status
docs-status:
	@docmgr status

# Generate report
docs-report:
	@echo "=== Documentation Summary ==="
	@docmgr status --summary-only
	@echo ""
	@echo "=== Stale Docs (>30 days) ==="
	@docmgr status --stale-after 30 --with-glaze-output --output json | \
	  jq -r '.docs[] | select(.stale) | "[\(.ticket)] \(.title) — stale \(.days_since_update) days"'
	@echo ""
	@echo "=== Recent Activity (7 days) ==="
	@docmgr doc search --updated-since "7 days ago"

# Clean up old date-based tickets
docs-clean:
	@echo "Archiving old tickets..."
	@find ttmp/202[0-3]-* -type d -maxdepth 0 -exec mv {} ttmp/archive/ \; 2>/dev/null || true

# Add to CI target
ci: test lint docs-validate

# Add to pre-commit
pre-commit: fmt lint docs-validate
```

**Usage:**
```bash
make docs-validate   # Before committing
make docs-status     # Check health
make docs-report     # Weekly review
make ci              # Full CI locally
```

---

## 5. Structured Output (Glaze Framework)

Every docmgr command that produces output can render it in multiple structured formats (JSON, CSV, YAML, TSV) through the Glaze framework. This design decouples the command's business logic from its output format, enabling the same command to serve both human users (with readable tables and text) and automation scripts (with parseable JSON or CSV).

### Available Output Formats

- `json` — Valid JSON, parseable
- `csv` — Comma-separated (for spreadsheets)
- `tsv` — Tab-separated
- `yaml` — YAML format
- `table` — ASCII table (human-readable)

### Stable Field Names (API Contract)

Use these with `--fields`, `--filter`, `--select`:

**Tickets:**
- `ticket`, `title`, `status`, `topics`, `path`, `last_updated`

**Docs:**
- `ticket`, `doc_type`, `title`, `status`, `topics`, `path`, `last_updated`

**Tasks:**
- `index`, `checked`, `text`, `file`

**Vocabulary:**
- `category`, `slug`, `description`

### Field Selection Examples

```bash
# Paths only (newline-separated)
docmgr list docs --ticket MEN-4242 --with-glaze-output --select path

# Custom columns (CSV)
docmgr list docs --with-glaze-output --output csv \
  --fields doc_type,title,path

# Templated output
docmgr list docs --ticket MEN-4242 --with-glaze-output \
  --select-template '{{.doc_type}}: {{.title}}' --select _0
```

The stable field contracts ensure your scripts won't break when docmgr is updated, making it safe to build CI/CD integrations, reporting dashboards, and bulk operation scripts on top of docmgr.

---

## 6. Automation Patterns

**Pattern 1: Find and update stale docs**

```bash
# Find docs older than 60 days, mark for review
docmgr doc search --updated-since "60 days ago" --with-glaze-output --output json | \
  jq -r '.[] | .path' | \
  while read doc; do
    docmgr meta update --doc "$doc" --field Status --value "needs-review"
  done
```

**Pattern 2: CI validation**

```bash
#!/bin/bash
# .github/workflows/validate-docs.yml

if ! docmgr doctor --all --stale-after 14 --fail-on error; then
  echo "ERROR: Documentation validation failed"
  # Get list of issues
  docmgr doctor --all --with-glaze-output --output json | \
    jq -r '.[] | select(.issue != "none") | "\(.path): \(.message)"'
  exit 1
fi
```

**Pattern 2b: Export the workspace index as a CI artifact (debugging)**

When CI fails or behaves differently than local, exporting the workspace index to SQLite can make investigation much faster: you can download the DB artifact and inspect exactly what docs/topics/related-files were indexed.

```bash
# Export index for offline inspection (artifact)
docmgr workspace export-sqlite --out diagnostics/docmgr-index.sqlite --force

# Optional: include markdown bodies for deeper debugging (larger file)
docmgr workspace export-sqlite --out diagnostics/docmgr-index-with-body.sqlite --force --include-body
```

The exported DB includes a `README` table populated from docmgr’s embedded docs (`pkg/doc/*.md`), so the artifact is self-describing even outside the repo.

**Pattern 3: Weekly doc report**

```bash
# Generate report of doc activity
docmgr status --stale-after 7 --with-glaze-output --output json | \
  jq -r '.docs[] | select(.stale) | "\(.ticket): \(.title) (stale \(.days_since_update) days)"'
```

**Pattern 4: Bulk operations**

```bash
# Create similar tickets
for i in {1..5}; do
    TICKET=PROJ-00$i
    docmgr ticket create-ticket --ticket $TICKET --title "Feature $i" --topics backend
    docmgr doc add --ticket $TICKET --doc-type design-doc --title "Design $i"
done

# Update all docs of a type
docmgr meta update --ticket MEN-4242 --doc-type design-doc \
    --field Status --value complete
```

---

## 7. Automated Reporting

### Weekly Documentation Report

```bash
#!/bin/bash
# scripts/weekly-docs-report.sh

echo "╔════════════════════════════════════════════════════════════╗"
echo "║  Weekly Documentation Report                               ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo ""

echo "📊 Overview:"
docmgr status --summary-only
echo ""

echo "⚠️  Stale Docs (>30 days):"
docmgr doc search --updated-since "30 days ago" --with-glaze-output --output json | \
  jq -r '.[] | "  • [\(.ticket)] \(.title) (updated: \(.last_updated))"'
echo ""

echo "📝 Recent Activity (last 7 days):"
docmgr doc search --updated-since "7 days ago" --with-glaze-output --output json | \
  jq -r '.[] | "  • [\(.ticket)] \(.title) — \(.doc_type)"' | head -10
echo ""

echo "🔍 Top Topics:"
docmgr doc list --with-glaze-output --output json | \
  jq -r '.[].topics' | tr ',' '\n' | sort | uniq -c | sort -rn | head -5
```

Run weekly in CI or via cron:

```yaml
# .github/workflows/weekly-report.yml
on:
  schedule:
    - cron: '0 9 * * MON'  # Every Monday at 9am

jobs:
  report:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Install docmgr
        run: go install -tags sqlite_fts5 github.com/go-go-golems/docmgr@latest
      - name: Generate report
        run: bash scripts/weekly-docs-report.sh
```

---

### Stale Doc Notifications

```bash
#!/bin/bash
# Find docs that haven't been updated in 60 days

echo "Docs requiring attention (>60 days since update):"

docmgr doc search --updated-since "60 days ago" --with-glaze-output --output json | \
  jq -r '.[] | {
    ticket: .ticket,
    title: .title,
    path: .path,
    updated: .last_updated
  } | "[\(.ticket)] \(.title)\n  Path: \(.path)\n  Last updated: \(.updated)\n"'
```

Send to Slack:
```bash
REPORT=$(bash scripts/stale-docs.sh)
curl -X POST "$SLACK_WEBHOOK_URL" \
  -d "{\"text\": \"$REPORT\"}"
```

---

## 6. Bulk Operations and Scripting

### Update All Docs in Batch

```bash
# Update status on all docs for completed ticket
docmgr meta update --ticket MEN-4242 --field Status --value complete

# Update owners across entire workspace
for ticket in $(docmgr ticket list --with-glaze-output --select ticket); do
  docmgr meta update --ticket "$ticket" --field Owners --value "new,team"
done
```

### Relate Files Automatically

```bash
# Auto-relate files from feature branch (notes required for each path)
git diff main --name-only | \
  grep -E '\.(go|ts|tsx|py)$' | \
  xargs -I FILE docmgr doc relate --ticket FEAT-042 \
    --file-note "FILE:Auto-related from git diff"

# With git commit messages as notes
for file in $(git diff main --name-only); do
  NOTE=$(git log -1 --pretty=%B "$file" | head -1)
  docmgr doc relate --ticket FEAT-042 --file-note "$file:$NOTE"
done
```

### Find Impacted Docs Before Refactoring

```bash
# Before renaming/moving files, find docs that reference them
docmgr doc search --file pkg/auth/service.go --with-glaze-output --output json | \
  jq -r '.[] | .path'

# Or search by directory
docmgr doc search --dir pkg/auth/ --with-glaze-output --output json | \
  jq -r '.[] | "\(.ticket): \(.title)"' | sort -u
```

### Generate Ticket-to-Code Map

```bash
#!/bin/bash
# Export all ticket → code relationships

echo "# Ticket to Code Map"
echo ""

for ticket in $(docmgr ticket list --with-glaze-output --select ticket); do
  echo "## $ticket"
  docmgr doc search --ticket "$ticket" --with-glaze-output --output json | \
    jq -r '.[0].related_files[]? | "- \(.path) — \(.note)"' 2>/dev/null
  echo ""
done
```

---

## 7. CI Strictness Strategies

### Strategy 1: Gradual Enforcement

```bash
# Week 1-2: Validate on warning (soft)
docmgr doctor --all --fail-on none  # Just print warnings

# Week 3-4: Block on errors only
docmgr doctor --all --fail-on error

# Week 5+: Block on errors and warnings
docmgr doctor --all --fail-on warning
```

### Strategy 2: Per-Branch Rules

```yaml
# Strict on main/production
on:
  push:
    branches: [main]
# run: docmgr doctor --all --fail-on warning

# Lenient on feature branches
on:
  pull_request:
# run: docmgr doctor --all --fail-on error
```

### Strategy 3: Focus on Changed Files

```bash
#!/bin/bash
# Only validate tickets with changes

CHANGED_TICKETS=$(git diff origin/main --name-only | \
  grep "^ttmp/" | cut -d/ -f2 | sort -u)

for ticket in $CHANGED_TICKETS; do
  docmgr doctor --ticket "$ticket" --fail-on error
done
```

---

## 8. Monitoring and Alerts

### Slack Integration

```bash
#!/bin/bash
# scripts/docs-health-slack.sh

STALE_COUNT=$(docmgr status --stale-after 30 --with-glaze-output --output json | \
  jq '.docs | map(select(.stale)) | length')

if [ "$STALE_COUNT" -gt 5 ]; then
  curl -X POST "$SLACK_WEBHOOK_URL" \
    -H 'Content-Type: application/json' \
    -d "{
      \"text\": \"⚠️  Documentation Health Alert\",
      \"blocks\": [{
        \"type\": \"section\",
        \"text\": {
          \"type\": \"mrkdwn\",
          \"text\": \"*$STALE_COUNT docs* are stale (>30 days). Review needed!\"
        }
      }]
    }"
fi
```

### Dashboard Metrics

```bash
#!/bin/bash
# Export metrics for dashboard

docmgr status --with-glaze-output --output json | \
  jq '{
    total_tickets: .tickets,
    total_docs: .docs,
    stale_docs: (.docs | map(select(.stale)) | length),
    by_type: (.docs | group_by(.doc_type) | map({type: .[0].doc_type, count: length}))
  }'
```

Send to monitoring system (Prometheus, Datadog, etc.).

---

## 9. Common Automation Patterns

Bulk operations and automation patterns transform docmgr from a documentation tool into a programmable documentation system. These patterns leverage structured output (`--with-glaze-output`) and command composability to automate tedious tasks like syncing metadata from external systems, bulk-updating stale docs, and generating documentation indexes. The key insight is using `docmgr doc search` to find target docs, then piping paths to `docmgr meta update` or `docmgr doc relate` for batch modifications.

### Pattern 1: Sync Metadata from External System

```bash
#!/bin/bash
# Sync owners from Jira

for ticket in $(docmgr list tickets --with-glaze-output --select ticket); do
  OWNERS=$(curl -s "jira.com/api/ticket/$ticket" | jq -r '.assignees | join(",")')
  if [ -n "$OWNERS" ]; then
    docmgr meta update --ticket "$ticket" --field Owners --value "$OWNERS"
  fi
done
```

### Pattern 2: Auto-Update Stale Doc Status

```bash
#!/bin/bash
# Mark stale docs for review

docmgr doc search --updated-since "60 days ago" --with-glaze-output --output json | \
  jq -r '.[] | .path' | \
  while read doc; do
    docmgr meta update --doc "$doc" --field Status --value "needs-review"
  done
```

### Pattern 3: Generate Documentation Index

```bash
#!/bin/bash
# Create DOCS.md with all tickets

echo "# Documentation Index" > DOCS.md
echo "" >> DOCS.md
echo "Auto-generated: $(date)" >> DOCS.md
echo "" >> DOCS.md

docmgr ticket list --with-glaze-output --output json | \
  jq -r '.[] | "## [\(.ticket)] \(.title)\n\n**Topics:** \(.topics)\n**Status:** \(.status)\n\n"' \
  >> DOCS.md
```

Run in CI, commit back to repo.

---

## 10. Troubleshooting CI Issues

### Issue: Doctor Fails in CI But Passes Locally

**Possible causes:**
- Different working directory
- Missing `.docmgrignore`
- Different `--stale-after` threshold

**Debug:**
```bash
# In CI, print what it sees
- name: Debug
  run: |
    pwd
    ls -la ttmp/
    docmgr status --summary-only
    docmgr doctor --all --fail-on none  # See all warnings
```

---

### Issue: False Positives from Archived Docs

**Solution:** Use `.docmgrignore`

```
# In ttmp/.docmgrignore
archive/
2023-*/
2024-*/
LEGACY-*/
```

---

### Issue: CI Takes Too Long

**Solutions:**

1. **Only validate changed tickets:**
```bash
CHANGED=$(git diff origin/main --name-only | grep "^ttmp/" | cut -d/ -f2 | sort -u)
for t in $CHANGED; do docmgr doctor --ticket "$t"; done
```

2. **Cache docmgr installation:**
```yaml
- name: Cache Go modules
  uses: actions/cache@v3
  with:
    path: ~/go/bin/docmgr
    key: ${{ runner.os }}-docmgr-${{ hashFiles('go.sum') }}
```

3. **Skip on draft PRs:**
```yaml
if: github.event.pull_request.draft == false
```

---

## 11. Advanced: Custom Validation Scripts

### Enforce Ticket Naming Convention

```bash
#!/bin/bash
# Validate ticket IDs match pattern

docmgr list tickets --with-glaze-output --output json | \
  jq -r '.[] | .ticket' | \
  while read ticket; do
    if ! [[ "$ticket" =~ ^[A-Z]+-[0-9]+$ ]]; then
      echo "❌ Invalid ticket ID: $ticket (expected: PROJ-123)"
      exit 1
    fi
  done
```

### Enforce Required RelatedFiles

```bash
#!/bin/bash
# Ensure design docs have related files

docmgr list docs --doc-type design-doc --with-glaze-output --output json | \
  jq -r '.[] | select(.related_files == null or (.related_files | length) == 0) | .path' | \
  while read doc; do
    echo "⚠️  Design doc missing RelatedFiles: $doc"
  done
```

### Enforce Summary Field

```bash
#!/bin/bash
# Check all docs have non-empty summaries

docmgr list docs --with-glaze-output --output json | \
  jq -r '.[] | select(.summary == "" or .summary == null) | "\(.ticket)/\(.path)"' | \
  while read doc; do
    echo "❌ Missing Summary: $doc"
    EXIT_CODE=1
  done

exit ${EXIT_CODE:-0}
```

---

## 12. Performance Optimization

### Caching Strategies

**Cache docmgr binary:**
```yaml
- name: Cache docmgr
  uses: actions/cache@v3
  with:
    path: ~/go/bin/docmgr
    key: docmgr-${{ hashFiles('**/go.sum') }}
  
- name: Install docmgr
  run: |
    if [ ! -f ~/go/bin/docmgr ]; then
      go install -tags sqlite_fts5 github.com/go-go-golems/docmgr@latest
    fi
```

**Parallel validation:**
```bash
# Validate tickets in parallel
docmgr ticket list --with-glaze-output --select ticket | \
  xargs -P 4 -I {} docmgr doctor --ticket {} --fail-on error
```

---

## Quick Reference

### CI Validation Commands

```bash
# Basic validation
docmgr doctor --all --fail-on error

# With staleness
docmgr doctor --all --stale-after 30 --fail-on error

# Specific ticket
docmgr doctor --ticket MEN-4242 --fail-on error

# JSON output for reporting
docmgr doctor --all --with-glaze-output --output json
```

### Common Make Targets

```bash
make docs-validate    # Run validation
make docs-status      # Check health
make docs-report      # Generate report
make ci               # Full CI including docs
```

### Pre-commit Hook Location

```bash
.git/hooks/pre-commit
chmod +x .git/hooks/pre-commit
```

---

## Related Documentation

- **Repository setup:** `docmgr help how-to-setup` — Initialize workspace
- **Daily usage:** `docmgr help how-to-use` — Creating and managing docs
- **Templates:** `docmgr help templates-and-guidelines` — Customization

---

## Examples from Real Projects

### Strict Validation (High Standards)

```yaml
# Requires: no errors, no warnings, no stale docs >14 days
validate-docs:
  script:
    - docmgr doctor --all --stale-after 14 --fail-on warning
```

### Lenient Validation (Growing Project)

```yaml
# Only blocks on broken links, allows stale docs
validate-docs:
  script:
    - docmgr doctor --all --fail-on error
    - docmgr status  # Print warnings but don't fail
```

### Balanced Validation (Most Teams)

```yaml
# Blocks on errors, warns on staleness
validate-docs:
  script:
    - docmgr doctor --all --stale-after 30 --fail-on error
    - |
      STALE=$(docmgr status --stale-after 30 --with-glaze-output --output json | \
        jq '.docs | map(select(.stale)) | length')
      if [ "$STALE" -gt 10 ]; then
        echo "⚠️  Warning: $STALE stale docs"
      fi
```

---

**This guide covers CI integration, automation, reporting, and monitoring for docmgr.**
