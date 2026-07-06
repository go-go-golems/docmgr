package tickets

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/go-go-golems/docmgr/internal/workspace"
	"github.com/go-go-golems/docmgr/pkg/models"
)

var (
	ErrNotFound   = errors.New("ticket not found")
	ErrAmbiguous  = errors.New("ticket ambiguous")
	datePathRegex = regexp.MustCompile(`^(\d{4})/(\d{2})/(\d{2})/`)
)

type Resolution struct {
	TicketID string

	TicketDirRel string
	TicketDirAbs string

	IndexPathRel string
	IndexPathAbs string

	IndexDoc *models.Document

	CreatedAt string // YYYY-MM-DD if inferred from path, else empty.
}

// Candidate is one known ticket, used for forgiving reference matching.
type Candidate struct {
	ID      string
	DirBase string // basename of the ticket directory (e.g. "MEN-4242--fix-chat-paths")
}

// ResolveTicketID resolves a user-provided ticket reference to a canonical ticket ID.
//
// Matching precedence (deterministic; first stage with matches wins):
//  1. exact ID match (case-sensitive, then case-insensitive)
//  2. unique ID prefix match
//  3. directory-name match (the "<ID>--slug" form agents paste): exact match on the
//     ticket directory basename, or the reference stripped at its first "--" matched
//     against IDs (exact, then unique prefix)
//  4. unique substring match against IDs
//
// On ambiguity the error lists the candidate IDs; on no match it suggests
// `docmgr ticket list`.
func ResolveTicketID(ctx context.Context, ws *workspace.Workspace, ref string) (string, error) {
	candidates, err := listTicketCandidates(ctx, ws)
	if err != nil {
		return "", err
	}
	return MatchTicketRef(ref, candidates)
}

// MatchTicketRef applies the forgiving matching precedence documented on
// ResolveTicketID against an in-memory candidate list.
func MatchTicketRef(ref string, candidates []Candidate) (string, error) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return "", errors.New("missing ticket id")
	}

	ids := make([]string, 0, len(candidates))
	for _, c := range candidates {
		ids = append(ids, c.ID)
	}

	// Stage 1: exact ID match.
	for _, id := range ids {
		if id == ref {
			return id, nil
		}
	}
	if matched := collectMatches(ids, func(id string) bool { return strings.EqualFold(id, ref) }); len(matched) > 0 {
		return uniqueOrAmbiguous(ref, matched)
	}

	// Stage 2: unique ID prefix match.
	lowerRef := strings.ToLower(ref)
	if matched := collectMatches(ids, func(id string) bool {
		return strings.HasPrefix(strings.ToLower(id), lowerRef)
	}); len(matched) > 0 {
		return uniqueOrAmbiguous(ref, matched)
	}

	// Stage 3: directory-name match ("<ID>--slug" or a full/relative ticket dir path).
	base := filepath.Base(filepath.ToSlash(strings.Trim(ref, "/")))
	if matched := collectCandidateMatches(candidates, func(c Candidate) bool {
		return c.DirBase != "" && strings.EqualFold(c.DirBase, base)
	}); len(matched) > 0 {
		return uniqueOrAmbiguous(ref, matched)
	}
	if idx := strings.Index(base, "--"); idx > 0 {
		head := base[:idx]
		lowerHead := strings.ToLower(head)
		if matched := collectMatches(ids, func(id string) bool { return strings.EqualFold(id, head) }); len(matched) > 0 {
			return uniqueOrAmbiguous(ref, matched)
		}
		if matched := collectMatches(ids, func(id string) bool {
			return strings.HasPrefix(strings.ToLower(id), lowerHead)
		}); len(matched) > 0 {
			return uniqueOrAmbiguous(ref, matched)
		}
	}

	// Stage 4: unique substring match against IDs (last resort).
	if matched := collectMatches(ids, func(id string) bool {
		return strings.Contains(strings.ToLower(id), lowerRef)
	}); len(matched) > 0 {
		return uniqueOrAmbiguous(ref, matched)
	}

	return "", fmt.Errorf("%w: %q (run 'docmgr ticket list' to see available tickets)", ErrNotFound, ref)
}

func collectMatches(ids []string, pred func(string) bool) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, id := range ids {
		if !pred(id) {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

func collectCandidateMatches(candidates []Candidate, pred func(Candidate) bool) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, c := range candidates {
		if !pred(c) {
			continue
		}
		if _, ok := seen[c.ID]; ok {
			continue
		}
		seen[c.ID] = struct{}{}
		out = append(out, c.ID)
	}
	return out
}

func uniqueOrAmbiguous(ref string, matched []string) (string, error) {
	if len(matched) == 1 {
		return matched[0], nil
	}
	sorted := append([]string{}, matched...)
	sort.Strings(sorted)
	return "", fmt.Errorf("%w: %q matches multiple tickets: %s", ErrAmbiguous, ref, strings.Join(sorted, ", "))
}

func listTicketCandidates(ctx context.Context, ws *workspace.Workspace) ([]Candidate, error) {
	handles, err := queryIndexDocs(ctx, ws)
	if err != nil {
		return nil, err
	}
	var out []Candidate
	for _, h := range handles {
		id := strings.TrimSpace(h.Doc.Ticket)
		if id == "" {
			continue
		}
		out = append(out, Candidate{
			ID:      id,
			DirBase: filepath.Base(filepath.Dir(filepath.Clean(h.Path))),
		})
	}
	return out, nil
}

func queryIndexDocs(ctx context.Context, ws *workspace.Workspace) ([]workspace.DocHandle, error) {
	res, err := ws.QueryDocs(ctx, workspace.DocQuery{
		Scope: workspace.Scope{Kind: workspace.ScopeRepo},
		Filters: workspace.DocFilters{
			DocType: "index",
		},
		Options: workspace.DocQueryOptions{
			IncludeBody:         false,
			IncludeErrors:       false,
			IncludeDiagnostics:  false,
			IncludeArchivedPath: true,
			IncludeScriptsPath:  true,
			IncludeSourcesPath:  true,
			IncludeControlDocs:  true,
			OrderBy:             workspace.OrderByPath,
		},
	})
	if err != nil {
		return nil, err
	}
	var out []workspace.DocHandle
	for _, h := range res.Docs {
		if h.Doc == nil {
			continue
		}
		if strings.TrimSpace(h.Doc.DocType) != "index" {
			continue
		}
		out = append(out, h)
	}
	return out, nil
}

// Resolve resolves a (possibly imprecise) ticket reference to its workspace
// location. See ResolveTicketID for the accepted reference forms.
func Resolve(ctx context.Context, ws *workspace.Workspace, ticketRef string) (Resolution, error) {
	ticketRef = strings.TrimSpace(ticketRef)
	if ticketRef == "" {
		return Resolution{}, errors.New("missing ticket id")
	}
	if ws == nil {
		return Resolution{}, errors.New("nil workspace")
	}

	handles, err := queryIndexDocs(ctx, ws)
	if err != nil {
		return Resolution{}, err
	}

	var candidates []Candidate
	for _, h := range handles {
		id := strings.TrimSpace(h.Doc.Ticket)
		if id == "" {
			continue
		}
		candidates = append(candidates, Candidate{
			ID:      id,
			DirBase: filepath.Base(filepath.Dir(filepath.Clean(h.Path))),
		})
	}

	ticketID, err := MatchTicketRef(ticketRef, candidates)
	if err != nil {
		return Resolution{}, err
	}

	var matches []workspace.DocHandle
	for _, h := range handles {
		if strings.TrimSpace(h.Doc.Ticket) != ticketID {
			continue
		}
		matches = append(matches, h)
	}

	if len(matches) == 0 {
		return Resolution{}, fmt.Errorf("%w: %q (run 'docmgr ticket list' to see available tickets)", ErrNotFound, ticketRef)
	}
	if len(matches) > 1 {
		paths := make([]string, 0, len(matches))
		for _, m := range matches {
			paths = append(paths, m.Path)
		}
		sort.Strings(paths)
		return Resolution{}, fmt.Errorf("%w: %q has multiple index docs: %s", ErrAmbiguous, ticketID, strings.Join(paths, ", "))
	}

	indexAbs := filepath.Clean(matches[0].Path)
	ticketDirAbs := filepath.Dir(indexAbs)
	indexRel := func() string {
		if rel, err := filepath.Rel(ws.Context().Root, indexAbs); err == nil {
			return filepath.ToSlash(rel)
		}
		return filepath.ToSlash(indexAbs)
	}()
	ticketDirRel := filepath.ToSlash(filepath.Dir(indexRel))

	createdAt := ""
	if m := datePathRegex.FindStringSubmatch(ticketDirRel + "/"); len(m) == 4 {
		createdAt = m[1] + "-" + m[2] + "-" + m[3]
	}

	docCopy := *matches[0].Doc

	return Resolution{
		TicketID:     ticketID,
		TicketDirRel: ticketDirRel,
		TicketDirAbs: ticketDirAbs,
		IndexPathRel: indexRel,
		IndexPathAbs: indexAbs,
		IndexDoc:     &docCopy,
		CreatedAt:    createdAt,
	}, nil
}
