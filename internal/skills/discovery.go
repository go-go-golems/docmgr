package skills

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-go-golems/docmgr/internal/paths"
	"github.com/go-go-golems/docmgr/internal/tickets"
	"github.com/go-go-golems/docmgr/internal/workspace"
	"github.com/pkg/errors"
)

// PlanHandle is a parsed plan with context for list/show.
type PlanHandle struct {
	Plan        *Plan
	Path        string
	DisplayPath string
	TicketID    string
	SourceFiles []SourceFile
}

// SourceFile captures file sources for matching.
type SourceFile struct {
	Path       string
	Output     string
	Normalized paths.NormalizedPath
}

// DiscoverOptions configures plan discovery.
type DiscoverOptions struct {
	TicketID          string
	IncludeWorkspace  bool
	IncludeAllTickets bool
}

// DiscoverPlans finds skill.yaml plans under workspace and ticket scopes.
func DiscoverPlans(ctx context.Context, ws *workspace.Workspace, opts DiscoverOptions) ([]PlanHandle, error) {
	if ws == nil {
		return nil, errors.New("workspace is required")
	}

	var planPaths []struct {
		Path     string
		TicketID string
	}

	if opts.IncludeWorkspace {
		workspacePlans, err := scanSkillPlans(filepath.Join(ws.Context().Root, "skills"))
		if err != nil {
			return nil, err
		}
		for _, p := range workspacePlans {
			planPaths = append(planPaths, struct {
				Path     string
				TicketID string
			}{Path: p})
		}
	}

	if strings.TrimSpace(opts.TicketID) != "" {
		res, err := tickets.Resolve(ctx, ws, strings.TrimSpace(opts.TicketID))
		if err != nil {
			return nil, errors.Wrap(err, "failed to resolve ticket")
		}

		ticketPlans, err := scanSkillPlans(filepath.Join(res.TicketDirAbs, "skills"))
		if err != nil {
			return nil, err
		}
		for _, p := range ticketPlans {
			planPaths = append(planPaths, struct {
				Path     string
				TicketID string
			}{Path: p, TicketID: res.TicketID})
		}
	}

	if strings.TrimSpace(opts.TicketID) == "" && opts.IncludeAllTickets {
		if ws.DB() == nil {
			return nil, errors.New("workspace index not initialized; call InitIndex before scanning tickets")
		}
		ticketPlans, err := scanAllTicketPlans(ctx, ws)
		if err != nil {
			return nil, err
		}
		planPaths = append(planPaths, ticketPlans...)
	}

	resolver := ws.Resolver()

	var handles []PlanHandle
	for _, planPath := range planPaths {
		plan, err := LoadPlan(planPath.Path)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to load skill plan %s", planPath.Path)
		}

		norm := resolver.Normalize(planPath.Path)
		displayPath := norm.Canonical
		if strings.TrimSpace(displayPath) == "" {
			displayPath = filepath.ToSlash(planPath.Path)
		}

		handle := PlanHandle{
			Plan:        plan,
			Path:        planPath.Path,
			DisplayPath: displayPath,
			TicketID:    strings.TrimSpace(planPath.TicketID),
		}

		for _, source := range plan.Sources {
			if strings.ToLower(strings.TrimSpace(source.Type)) != "file" {
				continue
			}
			normalized := resolver.NormalizeNoFS(source.Path)
			handle.SourceFiles = append(handle.SourceFiles, SourceFile{
				Path:       source.Path,
				Output:     source.Output,
				Normalized: normalized,
			})
		}

		handles = append(handles, handle)
	}

	return handles, nil
}

func scanAllTicketPlans(ctx context.Context, ws *workspace.Workspace) ([]struct {
	Path     string
	TicketID string
}, error) {
	res, err := ws.QueryDocs(ctx, workspace.DocQuery{
		Scope: workspace.Scope{Kind: workspace.ScopeRepo},
		Filters: workspace.DocFilters{
			DocType: "index",
		},
		Options: workspace.DocQueryOptions{
			IncludeErrors:       false,
			IncludeArchivedPath: true,
			IncludeScriptsPath:  true,
			IncludeControlDocs:  true,
			OrderBy:             workspace.OrderByPath,
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to query ticket index docs")
	}

	seen := map[string]struct{}{}
	var planPaths []struct {
		Path     string
		TicketID string
	}
	for _, handle := range res.Docs {
		if handle.Doc == nil {
			continue
		}
		ticketID := strings.TrimSpace(handle.Doc.Ticket)
		if ticketID == "" {
			continue
		}
		ticketDir := filepath.Clean(filepath.Dir(filepath.FromSlash(handle.Path)))
		if _, ok := seen[ticketDir]; ok {
			continue
		}
		seen[ticketDir] = struct{}{}

		ticketPlans, err := scanSkillPlans(filepath.Join(ticketDir, "skills"))
		if err != nil {
			return nil, err
		}
		for _, plan := range ticketPlans {
			planPaths = append(planPaths, struct {
				Path     string
				TicketID string
			}{Path: plan, TicketID: ticketID})
		}
	}

	return planPaths, nil
}

func scanSkillPlans(base string) ([]string, error) {
	if strings.TrimSpace(base) == "" {
		return nil, nil
	}
	if _, err := os.Stat(base); err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, errors.Wrap(err, "failed to stat skills directory")
	}

	var plans []string
	walkErr := filepath.WalkDir(base, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if strings.EqualFold(d.Name(), "skill.yaml") {
			plans = append(plans, path)
		}
		return nil
	})
	if walkErr != nil {
		return nil, errors.Wrap(walkErr, "failed to scan skills directory")
	}

	return plans, nil
}
