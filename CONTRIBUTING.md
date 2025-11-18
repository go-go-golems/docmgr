# Contributing to docmgr

Thank you for your interest in contributing to docmgr! This guide will help you get started with development.

## Development Setup

### Prerequisites

- Go 1.24 or later
- Git

### Getting Started

```bash
# Clone the repository
git clone https://github.com/go-go-golems/docmgr
cd docmgr

# Download dependencies
go mod download

# Build the project
go build ./cmd/docmgr

# Run tests
go test ./...

# Install locally (optional)
make install
```

## Architecture

docmgr is built using the [Glazed](https://github.com/go-go-golems/glazed) framework, which provides a consistent CLI structure and output formatting.

### Directory Structure

```
docmgr/
├── cmd/docmgr/          # Main CLI entry point
│   └── main.go          # Command registration and Cobra setup
├── pkg/
│   ├── commands/         # Command implementations
│   │   ├── add.go        # Add document command
│   │   ├── config.go     # Configuration management
│   │   ├── create_ticket.go
│   │   ├── doctor.go     # Validation command
│   │   └── ...           # Other commands
│   ├── models/           # Core data structures
│   │   └── document.go   # Document, Vocabulary, RelatedFiles types
│   ├── doc/              # Embedded help documentation
│   │   ├── docmgr-how-to-use.md
│   │   ├── docmgr-how-to-setup.md
│   │   └── ...
│   └── utils/            # Utility functions
│       └── slug.go       # Slug generation
├── test-scenarios/       # Integration test scenarios
└── ttmp/                 # Documentation workspace (for docmgr itself)
```

### Key Components

#### Commands (`pkg/commands/`)

Each command is implemented as a Glazed command:

- Commands implement the `glazed.Command` interface
- Use Glazed layers for parameter parsing
- Support both human-friendly and structured output (JSON/YAML/CSV)
- Commands are registered in `cmd/docmgr/main.go`

#### Data Models (`pkg/models/`)

- **Document**: Core document metadata structure with YAML frontmatter
- **Vocabulary**: Controlled vocabulary for topics, doc types, and intents
- **RelatedFiles**: List of related code files with optional notes
- **TicketWorkspace**: Metadata about a ticket's documentation workspace

#### Configuration (`pkg/commands/config.go`)

docmgr resolves its documentation root using a 6-level fallback chain:

1. `--root` flag (explicit command-line argument)
2. `.ttmp.yaml` in current directory
3. `.ttmp.yaml` in parent directories (walk up tree)
4. `DOCMGR_ROOT` environment variable
5. Git repository root: `<git-root>/ttmp`
6. Default: `ttmp` in current directory

## Adding a New Command

### Step 1: Create Command File

Create a new file in `pkg/commands/` following the naming pattern `{action}.go`:

```go
package commands

import (
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
)

func NewMyActionCommand() (*cmds.Command, error) {
	return cmds.NewCommand(
		cmds.WithShort("Short description"),
		cmds.WithLong("Long description"),
		cmds.WithFlags(
			parameters.NewParameterDefinition(
				"flag-name",
				parameters.ParameterTypeString,
				parameters.WithHelp("Flag description"),
				parameters.WithDefault("default-value"),
			),
		),
		cmds.WithRunFunc(func(ctx context.Context, ps map[string]interface{}) error {
			// Command implementation
			return nil
		}),
	)
}
```

### Step 2: Register Command in main.go

Add the command to `cmd/docmgr/main.go`:

```go
// Create my-action command
myActionCmd, err := commands.NewMyActionCommand()
if err != nil {
    fmt.Fprintf(os.Stderr, "Error creating my-action command: %v\n", err)
    os.Exit(1)
}

cobraMyActionCmd, err := cli.BuildCobraCommand(myActionCmd,
    cli.WithParserConfig(cli.CobraParserConfig{
        ShortHelpLayers: []string{layers.DefaultSlug},
        MiddlewaresFunc: cli.CobraCommandDefaultMiddlewares,
    }),
    cli.WithCobraMiddlewaresFunc(cli.CobraCommandDefaultMiddlewares),
    cli.WithCobraShortHelpLayers(layers.DefaultSlug),
)
if err != nil {
    fmt.Fprintf(os.Stderr, "Error building my-action command: %v\n", err)
    os.Exit(1)
}

rootCmd.AddCommand(cobraMyActionCmd)
```

### Step 3: Add Tests

Create `{action}_test.go` in the same directory:

```go
package commands

import (
	"testing"
)

func TestMyActionCommand(t *testing.T) {
	// Test implementation
}
```

### Step 4: Update Documentation

- Add help text in the command's `WithLong()` description
- If the command introduces new concepts, consider adding to embedded docs in `pkg/doc/`
- Update README.md if the command is a major feature

## Testing

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests for a specific package
go test ./pkg/commands

# Run tests with coverage
go test -cover ./...
```

### Integration Tests

docmgr includes integration test scenarios in `test-scenarios/testing-doc-manager/`:

```bash
cd test-scenarios/testing-doc-manager
./run-all.sh
```

These scenarios test end-to-end workflows and help ensure commands work correctly together.

## Code Style

### Formatting

- Use `gofmt` for formatting (or `goimports` for import organization)
- Run `golangci-lint` before committing: `make lint`

### Documentation

- **Package-level docs**: Every package should have a package comment explaining its purpose
- **Exported types**: Add godoc comments with examples where helpful
- **Complex functions**: Add comments explaining non-obvious logic
- **Commands**: Use descriptive `WithShort()` and `WithLong()` descriptions

### Error Handling

- Use `github.com/pkg/errors` for error wrapping: `errors.Wrap(err, "context")`
- Return errors from functions; don't log and continue silently
- Provide helpful error messages that guide users to solutions

### Naming Conventions

- Commands: `New{Action}Command()` (e.g., `NewAddCommand()`)
- Types: Use descriptive names (`WorkspaceConfig`, not `TTMPConfig`)
- Functions: Self-documenting names (`ResolveRoot()`, not `Resolve()`)

## Using Glazed Framework

docmgr uses the Glazed framework for CLI commands. Key concepts:

### Parameters

Define parameters using `parameters.NewParameterDefinition()`:

```go
parameters.NewParameterDefinition(
    "ticket",
    parameters.ParameterTypeString,
    parameters.WithHelp("Ticket ID (e.g., MEN-1234)"),
    parameters.WithRequired(true),
)
```

### Output Formats

Commands can support structured output (JSON/YAML/CSV) via Glazed:

```go
cobraCmd, err := cli.BuildCobraCommand(cmd,
    cli.WithDualMode(true),                    // Enable both human and structured output
    cli.WithGlazeToggleFlag("with-glaze-output"), // Add --with-glaze-output flag
    // ...
)
```

### Layers

Glazed uses "layers" for parameter parsing. Most commands use `layers.DefaultSlug`.

## Common Patterns

### Reading Configuration

```go
cfg, err := commands.LoadTTMPConfig()
if err != nil {
    return fmt.Errorf("failed to load config: %w", err)
}
```

### Resolving Documentation Root

```go
root := commands.ResolveRoot(providedRoot)
```

### Working with Documents

```go
doc := &models.Document{
    Title: "My Document",
    Ticket: "MEN-1234",
    DocType: "design-doc",
    Topics: []string{"api", "backend"},
}
```

## Getting Help

- **CLI Help**: `docmgr help` or `docmgr help how-to-use`
- **Embedded Docs**: `docmgr help --all` lists all available help topics
- **Issues**: Open an issue on GitHub for questions or bug reports

## Submitting Changes

1. **Fork the repository** and create a feature branch
2. **Write tests** for your changes
3. **Run tests and linting**: `make test && make lint`
4. **Update documentation** as needed
5. **Commit** with clear, descriptive messages
6. **Push** to your fork and open a pull request

## Code Review Process

- All PRs require review before merging
- Ensure tests pass and code is properly formatted
- Address review feedback promptly
- Keep PRs focused and reasonably sized

## Questions?

If you have questions about contributing, feel free to:
- Open an issue on GitHub
- Check existing documentation: `docmgr help`
- Review the codebase structure and existing command implementations

Thank you for contributing to docmgr!

