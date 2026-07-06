# docmgr - Document Manager for LLM Workflows

`docmgr` is a structured document manager for LLM-assisted workflows. It helps you create, organize, relate, search, and validate documentation workspaces with rich metadata and embedded, searchable help.

## Features

- Initialize docs root and ticket workspaces (`docmgr init`, `docmgr ticket create`), with a compact per-ticket overview (`docmgr ticket show`).
- Document templates and guidelines with embedded help (Glazed help system, `docmgr help --all`).
- Import external sources (files, snippets) into a workspace (`docmgr import file|snippet`).
- Frontmatter metadata management (`docmgr meta update`).
- Vocabulary management (`docmgr vocab list|add`), seeded by default on `init`.
- Powerful search across content and metadata (`docmgr search`, FTS5-backed with `-tags sqlite_fts5`).
- Task management in `tasks.md` with stable task IDs (`docmgr task add|list|check|edit|remove|migrate`).
- Changelog management (`docmgr changelog update`).
- Relate code files to docs/tickets with notes (`docmgr doc relate --file-note "path:why"`); paths are stored as explicit anchors (`repo://`, `ws://`, `docs://`, `abs://` — see `docmgr help path-anchors`).
- Workspace health checks (`docmgr doctor`) with a per-ticket rollup, safe auto-fixes (`--fix`, including anchor migration via `--fix-anchors`), and overall status (`docmgr status`).
- HTTP API server (`docmgr api serve`) with a versioned JSON API (`/api/v1/*`): search, docs, tickets, tasks, plus write endpoints for metadata, relate, changelog, and a doctor report (`docmgr help http-api`).
- Embedded web UI (React SPA served by `docmgr api serve`): workspace/ticket/topic browsing, search, doc viewer with mermaid/links/images, task and changelog editing, and a `/workspace/health` page (`docmgr help web-ui`).
- Skills: package docs into Agent-Skills format (`docmgr skill list|show|export|import`).
- Ticket graph rendering (`docmgr ticket graph`, Mermaid output).
- Verb output templates: post-render human output of key verbs with `.templ` files (`docmgr help verb-templates-and-schema`).

## Installation

Choose one of the following methods (mirroring other go-go-golems CLIs):

### Homebrew
```bash
brew tap go-go-golems/go-go-go
brew install go-go-golems/go-go-go/docmgr
```

### apt-get (Debian/Ubuntu)
```bash
echo "deb [trusted=yes] https://apt.fury.io/go-go-golems/ /" | sudo tee /etc/apt/sources.list.d/fury.list
sudo apt-get update
sudo apt-get install docmgr
```

### yum (RHEL/CentOS/Fedora)
```bash
sudo bash -c 'cat > /etc/yum.repos.d/fury.repo <<EOF
[fury]
name=Gemfury Private Repo
baseurl=https://yum.fury.io/go-go-golems/
enabled=1
gpgcheck=0
EOF'
sudo yum install docmgr
```

### go install
```bash
go install github.com/go-go-golems/docmgr/cmd/docmgr@latest
```

### Download binaries
Download prebuilt binaries from GitHub Releases.

### Run from source
```bash
git clone https://github.com/go-go-golems/docmgr
cd docmgr
go run ./cmd/docmgr --help
```

## Shell Completion

docmgr supports both dynamic completions (via carapace) and static completions (via cobra). Dynamic completions are recommended because they reflect live workspace state (tickets, vocabulary values, files).

### Dynamic (recommended) — carapace

- Bash (current session):
  ```bash
  source <(docmgr _carapace bash)
  ```
  Persist (typical):
  ```bash
  echo 'source <(docmgr _carapace bash)' >> ~/.bashrc
  # or system-wide (requires sudo):
  # docmgr _carapace bash | sudo tee /etc/bash_completion.d/docmgr >/dev/null
  ```

- Zsh:
  ```bash
  source <(docmgr _carapace zsh)
  # For persistent setup, add the same line to ~/.zshrc (after `compinit`)
  ```

- Fish:
  ```bash
  docmgr _carapace fish | source
  # Persist:
  docmgr _carapace fish > ~/.config/fish/completions/docmgr.fish
  ```

- PowerShell:
  ```powershell
  docmgr _carapace powershell | Out-String | Invoke-Expression
  # Persist: add the same line to $PROFILE
  ```

Notes:
- The dynamic snippet calls back into `docmgr` via a hidden `_carapace` subcommand to compute completions at runtime.
- This enables live suggestions for flags like `--ticket`, `--doc-type`, `--status`, `--intent`, `--topics`, and file/directory flags.

### Static — cobra

If you prefer cobra’s static completion scripts:

- Bash:
  ```bash
  docmgr completion bash | sudo tee /etc/bash_completion.d/docmgr >/dev/null
  # or user shell: echo 'source <(docmgr completion bash)' >> ~/.bashrc
  ```
- Zsh:
  ```bash
  docmgr completion zsh > ~/.zfunc/_docmgr
  echo 'fpath+=~/.zfunc' >> ~/.zshrc
  echo 'autoload -Uz compinit && compinit' >> ~/.zshrc
  ```
- Fish:
  ```bash
  docmgr completion fish > ~/.config/fish/completions/docmgr.fish
  ```
- PowerShell:
  ```powershell
  docmgr completion powershell | Out-String | Invoke-Expression
  ```

Static scripts don’t reflect live values for dynamic flags; use dynamic completions for most workflows.

## Quick Start

```bash
# Write a repo config (.ttmp.yaml) quickly
docmgr configure --root ttmp --owners manuel --intent long-term --vocabulary ttmp/vocabulary.yaml

# Initialize docs root (creates vocabulary/templates/guidelines if missing;
# vocabulary is seeded with defaults, pass --seed-vocabulary=false to skip)
docmgr init

# Create a new ticket workspace
docmgr ticket create --ticket MEN-1234 --title "Design Overview" --topics design,backend

# Show a compact ticket overview
docmgr ticket show MEN-1234

# Rename a ticket ID and move its workspace
docmgr ticket rename --ticket MEN-1234 --new-ticket MEN-5678

# Add a document to the ticket
docmgr doc add --ticket MEN-5678 --doc-type design-doc --title "System Overview"

# List tickets and docs
docmgr list tickets
docmgr list docs --ticket MEN-5678

# Search across content and metadata
docmgr search --query "design"

# See overall status
docmgr status

# Get help (topics and commands)
docmgr help
```

Workspaces are created under `ttmp/YYYY/MM/DD/<ticket>--<slug>/` by default. Use `--path-template` to customize the relative layout (placeholders: `{{YYYY}}`, `{{MM}}`, `{{DD}}`, `{{DATE}}`, `{{TICKET}}`, `{{SLUG}}`, `{{TITLE}}`).

## Usage

Run `docmgr --help` or `docmgr help` to see all commands and options. Use `--long-help` for detailed flags and topics.

### Help topics

```bash
# Full command/topic listing
docmgr help --all

# Tutorials
docmgr help how-to-setup
docmgr help how-to-use
```

## Development

- Go 1.25+
- Build: `go build ./...` (add `-tags sqlite_fts5` for full-text search; `make build` uses `-tags "sqlite_fts5,embed"` to also embed the web UI)
- Lint/Test: `make lint`, `make test`
- Web UI: `make ui-build` (Dagger + pnpm pipeline), then `make build-embed`
- Release (when configured): `make goreleaser`

See [CONTRIBUTING.md](./CONTRIBUTING.md) for development setup and contribution guidelines.

## Glossary

Key terms used throughout docmgr:

- **Workspace**: Root directory containing ticket documentation (default: `ttmp`). The workspace is organized hierarchically by date and ticket ID.

- **Ticket**: Work item identifier (e.g., `MEN-3475`, `DOCMGR-123`). Tickets group related documentation together.

- **Ticket Workspace**: Directory structure for a single ticket's documentation. Typically organized as `ttmp/YYYY/MM/DD/<ticket>--<slug>/` containing `index.md`, `tasks.md`, `changelog.md`, and per-doc-type subdirectories like `design-doc/`, `reference/`, `playbook/`.

- **Doc Type**: Category of document that determines its purpose and structure. Common types include:
  - `design-doc`: Architecture and design decisions
  - `playbook`: Step-by-step procedures and runbooks
  - `analysis`: Research and analysis documents
  - `reference`: API references and documentation
  - `log`: Implementation diaries and notes
  - `task-list`: Task tracking documents

- **Vocabulary**: Controlled vocabulary for standardizing metadata values. Defines allowed values for `topics`, `docTypes`, and `intent` fields across all documents. Stored in `vocabulary.yaml` at the workspace root.

- **Frontmatter**: YAML metadata block at the top of markdown documents containing fields like `Title`, `Ticket`, `DocType`, `Topics`, `Owners`, `RelatedFiles`, etc.

- **Related Files**: Code files linked to documentation via the `RelatedFiles` frontmatter field. Enables traceability from documentation to implementation.

## License

MIT

## Configuration (multi-repo friendly)

`docmgr` resolves its docs root in this order (see `internal/workspace/config.go`):

1. Flag: `--root /abs/or/relative/path` (relative paths are anchored on the current working directory)
2. Nearest `.ttmp.yaml`: located via the `DOCMGR_CONFIG` environment variable if set, otherwise by walking up from CWD; its `root: <path>` is resolved relative to the config file location
3. Git repository root: `<git-root>/ttmp`
4. Fallback: `<cwd>/ttmp`

Recommended multi-repo setup: place a `.ttmp.yaml` at the workspace root and point it at the repo-local `ttmp/`. You can create this file via `docmgr configure`:

```bash
docmgr configure --root ttmp --owners manuel --intent long-term --vocabulary ttmp/vocabulary.yaml
```

```yaml
root: go-go-mento/ttmp
defaults:
  owners: [manuel]
  intent: long-term
vocabulary: go-go-mento/ttmp/vocabulary.yaml
```

Environment variables:

- `DOCMGR_CONFIG`: explicit path to a `.ttmp.yaml` config file (absolute, or relative to the current working directory). Takes precedence over the walk-up search.
- `DOCMGR_DEBUG`: when set, logs each step of config/root resolution to stderr.
