# docmgr - Document Manager for LLM Workflows

`docmgr` is a structured document manager for LLM-assisted workflows. It helps you create, organize, relate, search, and validate documentation workspaces with rich metadata and embedded, searchable help.

## Features

- Initialize docs root and ticket workspaces (`docmgr init`, `docmgr create-ticket`).
- Document templates and guidelines with embedded help (Glazed help system).
- Import external sources (files, snippets) into a workspace.
- Frontmatter metadata management (`docmgr meta update`).
- Vocabulary management (`docmgr vocab list|add`).
- Powerful search across content and metadata (`docmgr search`).
- Tasks management in `tasks.md` (`docmgr tasks ...`).
- Changelog management (`docmgr changelog update`).
- Workspace health checks (`docmgr doctor`) and overall status (`docmgr status`).
- Relate code files to docs/tickets (`docmgr relate`).

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

## Quick Start

```bash
# Write a repo config (.ttmp.yaml) quickly
docmgr configure --root ttmp --owners manuel --intent long-term --vocabulary ttmp/vocabulary.yaml

# Initialize docs root (creates vocabulary/templates/guidelines if missing)
docmgr init --seed-vocabulary

# Create a new ticket workspace
docmgr create-ticket --ticket MEN-1234 --title "Design Overview" --topics design,backend

Workspaces are created under `ttmp/YYYY/MM/DD/<ticket>-<slug>/` by default. Use `--path-template` to customize the relative layout (placeholders: `{{YYYY}}`, `{{MM}}`, `{{DD}}`, `{{DATE}}`, `{{TICKET}}`, `{{SLUG}}`, `{{TITLE}}`).

# Add a document to the ticket
docmgr add --ticket MEN-1234 --doc-type design-doc --title "System Overview"

# List tickets and docs
docmgr list tickets
docmgr list docs --ticket MEN-1234

# Search across content and metadata
docmgr search --query "glazed"

# See overall status
docmgr status

# Get help (topics and commands)
docmgr help
```

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

- Go 1.24+
- Build: `go build ./...`
- Lint/Test (if configured): `make lint`, `make test`
- Release (when configured): `make goreleaser`

See [CONTRIBUTING.md](./CONTRIBUTING.md) for development setup and contribution guidelines.

## Glossary

Key terms used throughout docmgr:

- **Workspace**: Root directory containing ticket documentation (default: `ttmp`). The workspace is organized hierarchically by date and ticket ID.

- **Ticket**: Work item identifier (e.g., `MEN-3475`, `DOCMGR-123`). Tickets group related documentation together.

- **Ticket Workspace**: Directory structure for a single ticket's documentation. Typically organized as `ttmp/YYYY/MM/DD/<ticket>-<slug>/` containing subdirectories like `analysis/`, `design/`, `playbooks/`, etc.

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

`docmgr` resolves its docs root in this order:

1. Flag: `--root /abs/or/relative/path`
2. Nearest `.ttmp.yaml` walking up from CWD: `root: <path>` (relative paths are resolved relative to the config file location)
3. Git repository root: `<git-root>/ttmp` (CLI)
4. Fallback: `<cwd>/ttmp` (CLI) or `docs` (server)

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

Environment overrides:

- `DOCMGR_ROOT`: absolute or relative path to the docs root. If relative, it is resolved against the current working directory.
