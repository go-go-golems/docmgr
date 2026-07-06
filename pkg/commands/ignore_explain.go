package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-go-golems/docmgr/internal/workspace"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
)

type IgnoreExplainCommand struct {
	*cmds.CommandDescription
}

type IgnoreExplainSettings struct {
	Root  string `glazed:"root"`
	Path  string `glazed:"path"`
	IsDir bool   `glazed:"is-dir"`
	Trace bool   `glazed:"trace"`
}

func NewIgnoreExplainCommand() (*IgnoreExplainCommand, error) {
	return &IgnoreExplainCommand{
		CommandDescription: cmds.NewCommandDescription(
			"explain",
			cmds.WithShort("Explain whether docmgr ignore policy ignores a path"),
			cmds.WithLong(`Explains the workspace-owned docmgr ignore decision for a path.

The command uses the same ignore matcher that workspace indexing uses: built-in
excludes plus repository/docs/nested .docmgrignore files. Use this when doctor,
list, or search appear to include or exclude a surprising path.

Examples:
  docmgr ignore explain ttmp/2026/06/08/TICKET--slug/scripts/node_modules/pkg/README.md
  docmgr ignore explain --root ttmp 2026/06/08/TICKET--slug/scripts/node_modules --is-dir
  docmgr ignore explain --trace ttmp/2026/06/08/TICKET--slug/reference/01-plan.md`),
			cmds.WithArguments(
				fields.New(
					"path",
					fields.TypeString,
					fields.WithHelp("Path to explain (absolute, repo-relative, or docs-root-relative)"),
					fields.WithRequired(true),
				),
			),
			cmds.WithFlags(
				fields.New(
					"root",
					fields.TypeString,
					fields.WithHelp("Root directory for docs"),
					fields.WithDefault("ttmp"),
				),
				fields.New(
					"is-dir",
					fields.TypeBool,
					fields.WithHelp("Treat the path as a directory"),
					fields.WithDefault(false),
				),
				fields.New(
					"trace",
					fields.TypeBool,
					fields.WithHelp("Emit one row per matcher source instead of only the final decision"),
					fields.WithDefault(false),
				),
			),
		),
	}, nil
}

func (c *IgnoreExplainCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedValues *values.Values,
	gp middlewares.Processor,
) error {
	settings := &IgnoreExplainSettings{}
	if err := parsedValues.DecodeSectionInto(schema.DefaultSlug, settings); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}
	settings.Path = strings.TrimSpace(settings.Path)
	if settings.Path == "" {
		return fmt.Errorf("path is required")
	}

	ws, err := workspace.DiscoverWorkspace(ctx, workspace.DiscoverOptions{RootOverride: settings.Root})
	if err != nil {
		return fmt.Errorf("failed to discover workspace: %w", err)
	}
	matcher := ws.IgnoreMatcher()
	if matcher == nil {
		return fmt.Errorf("workspace ignore matcher is not initialized")
	}
	decision := matcher.Match(settings.Path, settings.IsDir)

	if settings.Trace {
		for i, step := range decision.Trace {
			row := types.NewRow(
				types.MRP("path", decision.Path),
				types.MRP("is_dir", decision.IsDir),
				types.MRP("final_ignored", decision.Ignored),
				types.MRP("trace_index", i),
				types.MRP("source_kind", string(step.SourceKind)),
				types.MRP("source", step.SourceName),
				types.MRP("matched", step.Matched),
				types.MRP("ignored", step.Ignored),
				types.MRP("pattern", step.Pattern),
				types.MRP("pattern_file", step.File),
				types.MRP("pattern_line", step.Line),
				types.MRP("pattern_column", step.Column),
			)
			if err := gp.AddRow(ctx, row); err != nil {
				return fmt.Errorf("failed to emit trace row: %w", err)
			}
		}
		return nil
	}

	row := types.NewRow(
		types.MRP("path", decision.Path),
		types.MRP("is_dir", decision.IsDir),
		types.MRP("ignored", decision.Ignored),
		types.MRP("matched", decision.Matched),
		types.MRP("source_kind", string(decision.SourceKind)),
		types.MRP("source", decision.SourceName),
		types.MRP("pattern", decision.Pattern),
		types.MRP("pattern_file", decision.PatternFile),
		types.MRP("pattern_line", decision.PatternLine),
		types.MRP("pattern_column", decision.PatternColumn),
		types.MRP("docs_root", matcher.DocsRoot()),
		types.MRP("repo_root", matcher.RepoRoot()),
	)
	if err := gp.AddRow(ctx, row); err != nil {
		return fmt.Errorf("failed to emit ignore decision row: %w", err)
	}
	return nil
}

var _ cmds.GlazeCommand = &IgnoreExplainCommand{}

func (c *IgnoreExplainCommand) Run(
	ctx context.Context,
	parsedValues *values.Values,
) error {
	collector := &relateRowCollector{}
	if err := c.RunIntoGlazeProcessor(ctx, parsedValues, collector); err != nil {
		return err
	}

	for _, row := range collector.rows {
		if idx, ok := row.Get("trace_index"); ok {
			if fmt.Sprint(idx) == "0" {
				path, _ := row.Get("path")
				isDir, _ := row.Get("is_dir")
				finalIgnored, _ := row.Get("final_ignored")
				fmt.Printf("Path: %v (dir: %v)\n", path, isDir)
				fmt.Printf("Ignored: %v\n", finalIgnored)
				fmt.Println("Trace:")
			}
			source, _ := row.Get("source")
			sourceKind, _ := row.Get("source_kind")
			matched, _ := row.Get("matched")
			ignored, _ := row.Get("ignored")
			line := fmt.Sprintf("- source=%v kind=%v matched=%v ignored=%v", source, sourceKind, matched, ignored)
			if pattern := ignoreExplainField(row, "pattern"); pattern != "" {
				line += fmt.Sprintf(" pattern=%q", pattern)
			}
			if file := ignoreExplainField(row, "pattern_file"); file != "" {
				lineNo, _ := row.Get("pattern_line")
				line += fmt.Sprintf(" (%s:%v)", file, lineNo)
			}
			fmt.Println(line)
			continue
		}

		path, _ := row.Get("path")
		isDir, _ := row.Get("is_dir")
		ignored, _ := row.Get("ignored")
		matched, _ := row.Get("matched")
		fmt.Printf("Path: %v (dir: %v)\n", path, isDir)
		fmt.Printf("Ignored: %v\n", ignored)
		fmt.Printf("Matched: %v\n", matched)
		if source := ignoreExplainField(row, "source"); source != "" {
			sourceKind, _ := row.Get("source_kind")
			fmt.Printf("Source: %s (kind: %v)\n", source, sourceKind)
		}
		if pattern := ignoreExplainField(row, "pattern"); pattern != "" {
			fmt.Printf("Pattern: %q", pattern)
			if file := ignoreExplainField(row, "pattern_file"); file != "" {
				lineNo, _ := row.Get("pattern_line")
				fmt.Printf(" (%s:%v)", file, lineNo)
			}
			fmt.Println()
		}
		if docsRoot := ignoreExplainField(row, "docs_root"); docsRoot != "" {
			fmt.Printf("Docs root: %s\n", docsRoot)
		}
		if repoRoot := ignoreExplainField(row, "repo_root"); repoRoot != "" {
			fmt.Printf("Repo root: %s\n", repoRoot)
		}
	}
	return nil
}

func ignoreExplainField(row types.Row, key string) string {
	v, ok := row.Get(key)
	if !ok || v == nil {
		return ""
	}
	s := strings.TrimSpace(fmt.Sprint(v))
	if s == "<nil>" {
		return ""
	}
	return s
}

var _ cmds.BareCommand = &IgnoreExplainCommand{}
