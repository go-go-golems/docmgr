package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-go-golems/docmgr/internal/documents"
	"github.com/go-go-golems/docmgr/internal/operations"
	"github.com/go-go-golems/docmgr/internal/paths"
	"github.com/go-go-golems/docmgr/internal/tasksmd"
)

type EvidenceRef struct {
	Kind     string `json:"kind"`
	Path     string `json:"path"`
	Revision string `json:"revision"`
	Claim    string `json:"claim"`
}
type MilestoneRequest struct {
	Ticket      string            `json:"ticket"`
	OperationID string            `json:"operation_id"`
	Summary     string            `json:"summary"`
	Phase       string            `json:"phase"`
	Next        string            `json:"next"`
	TaskIDs     []string          `json:"task_ids"`
	Evidence    []EvidenceRef     `json:"evidence"`
	Expected    map[string]string `json:"expected"`
}

// RecordMilestone checks references, but never decides whether evidence proves a
// claim. Evidence revisions are SHA256 of ticket-local files, not Git revisions.
func RecordMilestone(ctx context.Context, dir string, req MilestoneRequest, dry bool) (operations.Receipt, error) {
	return recordMilestone(ctx, dir, req, dry, operations.Store{})
}
func recordMilestone(ctx context.Context, dir string, req MilestoneRequest, dry bool, store operations.Store) (operations.Receipt, error) {
	if strings.TrimSpace(req.Ticket) == "" || strings.TrimSpace(req.Summary) == "" || len(req.Summary) > 8192 || len(req.Next) > 4096 || len(req.TaskIDs) > 128 || len(req.Evidence) > 64 {
		return operations.Receipt{}, fmt.Errorf("invalid milestone request bounds")
	}
	switch req.Phase {
	case "start", "implement", "validate", "checkpoint", "resume", "close":
	default:
		return operations.Receipt{}, fmt.Errorf("invalid phase")
	}
	return store.Execute(ctx, dir, req.OperationID, req, dry, func() ([]operations.Change, error) {
		doc, _, err := documents.ReadDocumentWithFrontmatter(filepath.Join(dir, "index.md"))
		if err != nil {
			return nil, err
		}
		if doc.Ticket != req.Ticket {
			return nil, fmt.Errorf("ticket identity mismatch")
		}
		for name, want := range req.Expected {
			if name != "index.md" && name != "tasks.md" && name != "changelog.md" {
				return nil, fmt.Errorf("invalid expected path")
			}
			_, got, err := operations.Read(dir, name)
			if err != nil {
				return nil, err
			}
			if got != want {
				return nil, fmt.Errorf("revision conflict: %s", name)
			}
		}
		for _, e := range req.Evidence {
			if strings.TrimSpace(e.Claim) == "" || strings.TrimSpace(e.Kind) == "" || len(e.Claim) > 4096 || len(e.Path) > 4096 {
				return nil, fmt.Errorf("invalid evidence")
			}
			_, got, err := readEvidence(dir, e.Path)
			if err != nil {
				return nil, err
			}
			if got == "missing" || got != e.Revision {
				return nil, fmt.Errorf("evidence revision mismatch: %s", e.Path)
			}
		}
		changes := []operations.Change{}
		if len(req.TaskIDs) > 0 {
			raw, before, err := operations.Read(dir, "tasks.md")
			if err != nil {
				return nil, err
			}
			lines := strings.Split(strings.TrimSuffix(string(raw), "\n"), "\n")
			_, tasks := tasksmd.Parse(lines)
			stable := map[string]bool{}
			for _, task := range tasks {
				if task.StableID != "" {
					if stable[task.StableID] {
						return nil, fmt.Errorf("duplicate stable task ID: %s", task.StableID)
					}
					stable[task.StableID] = true
				}
			}
			for _, id := range req.TaskIDs {
				if !stable[id] {
					return nil, fmt.Errorf("unknown stable task ID: %s", id)
				}
			}
			next, err := tasksmd.ToggleCheckedByRefs(lines, req.TaskIDs, true)
			if err != nil {
				return nil, err
			}
			after := []byte(strings.Join(next, "\n") + "\n")
			if operations.Hash(after) != before {
				changes = append(changes, operations.Change{Path: "tasks.md", Before: before, After: after})
			}
		}
		old, before, err := operations.Read(dir, "changelog.md")
		if err != nil {
			return nil, err
		}
		now := time.Now()
		if store.Now != nil {
			now = store.Now()
		}
		entry := req.Summary + "\n\nOperation: `" + req.OperationID + "`"
		after := BuildChangelogEntry(string(old), now.Format("2006-01-02"), "", entry, nil)
		changes = append(changes, operations.Change{Path: "changelog.md", Before: before, After: []byte(after)})
		return changes, nil
	})
}
func readEvidence(dir, raw string) ([]byte, string, error) {
	if raw == "" || filepath.IsAbs(raw) {
		return nil, "", fmt.Errorf("evidence must be ticket-local")
	}
	if !strings.Contains(raw, "://") {
		raw = "doc://" + raw
	}
	resolver := paths.NewResolver(paths.ResolverOptions{DocPath: filepath.Join(dir, "index.md")})
	n := resolver.ResolveNoFS(raw)
	rel, err := filepath.Rel(dir, n.Abs)
	if err != nil || !filepath.IsLocal(rel) {
		return nil, "", fmt.Errorf("evidence escapes ticket: %s", raw)
	}
	return operations.Read(dir, rel)
}

type ResumeView struct {
	SchemaVersion int                 `json:"schema_version"`
	Ticket        string              `json:"ticket"`
	Status        string              `json:"status"`
	Phase         string              `json:"phase"`
	Next          string              `json:"next"`
	Latest        *operations.Receipt `json:"latest_milestone"`
	Remaining     []tasksmd.Item      `json:"remaining"`
	Revisions     map[string]string   `json:"revisions"`
	Documents     []string            `json:"documents"`
	Evidence      []EvidenceRef       `json:"evidence"`
	Conflicts     []string            `json:"conflicts"`
}

func TicketResume(ctx context.Context, dir string) (ResumeView, error) {
	view := ResumeView{SchemaVersion: 1, Phase: "start", Remaining: []tasksmd.Item{}, Revisions: map[string]string{}, Documents: []string{}, Evidence: []EvidenceRef{}, Conflicts: []string{}}
	err := operations.WithLock(ctx, dir, func() error {
		doc, _, err := documents.ReadDocumentWithFrontmatter(filepath.Join(dir, "index.md"))
		if err != nil {
			return err
		}
		view.Ticket = doc.Ticket
		view.Status = doc.Status
		for _, name := range []string{"index.md", "tasks.md", "changelog.md"} {
			_, hash, err := operations.Read(dir, name)
			if err != nil {
				return err
			}
			view.Revisions[name] = hash
		}
		raw, _, err := operations.Read(dir, "tasks.md")
		if err != nil {
			return err
		}
		parsed, _ := tasksmd.Parse(strings.Split(string(raw), "\n"))
		for _, section := range parsed.Sections {
			for _, task := range section.Items {
				if !task.Checked {
					view.Remaining = append(view.Remaining, task)
				}
			}
		}
		docs, err := filepath.Glob(filepath.Join(dir, "*", "*.md"))
		if err != nil {
			return err
		}
		if len(docs) > 512 {
			return fmt.Errorf("resume document limit exceeded (512)")
		}
		for _, p := range docs {
			rel, err := filepath.Rel(dir, p)
			if err != nil {
				return err
			}
			view.Documents = append(view.Documents, filepath.ToSlash(rel))
		}
		records, err := operations.Records(dir)
		if err != nil {
			return err
		}
		for _, record := range records {
			if record.Receipt.State != "committed" {
				view.Conflicts = append(view.Conflicts, "pending operation: "+record.Receipt.OperationID)
				continue
			}
			var req MilestoneRequest
			if err := json.Unmarshal(record.Request, &req); err != nil {
				return err
			}
			// Close receipts share the store, but are not milestone checkpoints.
			if req.Phase != "" {
				receipt := record.Receipt
				view.Latest = &receipt
				view.Phase = req.Phase
				view.Next = req.Next
				view.Evidence = req.Evidence
			}
		}
		if view.Latest != nil {
			for name, want := range view.Latest.After {
				if got := view.Revisions[name]; got != want {
					view.Conflicts = append(view.Conflicts, "projection changed since milestone: "+name)
				}
			}
			for _, e := range view.Evidence {
				_, hash, err := readEvidence(dir, e.Path)
				if err != nil || hash != e.Revision {
					view.Conflicts = append(view.Conflicts, "evidence changed or unavailable: "+e.Path)
				}
			}
		}
		return nil
	})
	return view, err
}
