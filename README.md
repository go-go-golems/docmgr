# docmgr - Document Manager for LLM Workflows

`docmgr` is a structured document manager for LLM-assisted workflows. It helps you create, organize, relate, search, and validate documentation workspaces with rich metadata and embedded, searchable help.

## Features

- Initialize ticket/document workspaces with sane defaults (`docmgr init`).
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
# Initialize a new workspace at the current directory
docmgr init --workspace .

# Add a document to the current workspace
docmgr add --title "Design Overview" --type note

# List tickets and docs
docmgr list tickets
docmgr list docs

# Search across content and metadata
docmgr search --query "glazed" --type markdown

# See overall status
docmgr status

# Get help (topics and commands)
docmgr help
```

## Usage

Run `docmgr --help` or `docmgr help` to see all commands and options. Use `--long-help` for detailed flags and topics.

## Development

- Go 1.24+
- Build: `go build ./...`
- Lint/Test (if configured): `make lint`, `make test`
- Release (when configured): `make goreleaser`

## License

MIT
