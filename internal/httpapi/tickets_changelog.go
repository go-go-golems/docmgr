package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-go-golems/docmgr/internal/tickets"
	"github.com/go-go-golems/docmgr/internal/workspace"
	"github.com/go-go-golems/docmgr/pkg/commands"
)

type ticketChangelogResponse struct {
	Ticket  string                    `json:"ticket"`
	Exists  bool                      `json:"exists"`
	Path    string                    `json:"path"`
	Entries []commands.ChangelogEntry `json:"entries"`
}

// handleTicketsChangelog serves GET (parsed date-sectioned entries) and POST
// (append an entry via the 'docmgr changelog update' primitive) for a
// ticket's changelog.md.
func (s *Server) handleTicketsChangelog(w http.ResponseWriter, r *http.Request) error {
	switch r.Method {
	case http.MethodGet:
		return s.handleTicketsChangelogGet(w, r)
	case http.MethodPost:
		return s.handleTicketsChangelogPost(w, r)
	default:
		return NewHTTPError(http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", nil)
	}
}

func (s *Server) handleTicketsChangelogGet(w http.ResponseWriter, r *http.Request) error {
	ticketID := strings.TrimSpace(r.URL.Query().Get("ticket"))
	if ticketID == "" {
		return NewHTTPError(http.StatusBadRequest, "invalid_argument", "missing ticket", map[string]any{"field": "ticket"})
	}

	var resp ticketChangelogResponse
	if err := s.mgr.WithWorkspace(func(ws *workspace.Workspace) error {
		res, err := resolveTicketOrHTTPError(r, ws, ticketID)
		if err != nil {
			return err
		}

		rawPath := filepath.ToSlash(filepath.Join(res.TicketDirRel, "changelog.md"))
		abs, rel, _, err := resolveFileWithin(ws.Context().Root, rawPath)
		if err != nil {
			var he *HTTPError
			if errors.As(err, &he) && he.Status == http.StatusNotFound {
				resp = ticketChangelogResponse{
					Ticket:  ticketID,
					Exists:  false,
					Path:    rawPath,
					Entries: []commands.ChangelogEntry{},
				}
				return nil
			}
			return err
		}

		raw, err := os.ReadFile(abs) // #nosec G304 -- abs went through resolveFileWithin
		if err != nil {
			return err
		}
		resp = ticketChangelogResponse{
			Ticket:  ticketID,
			Exists:  true,
			Path:    rel,
			Entries: commands.ParseChangelogEntries(string(raw)),
		}
		return nil
	}); err != nil {
		return err
	}

	return writeJSON(w, http.StatusOK, resp)
}

type ticketChangelogAppendRequest struct {
	Ticket string `json:"ticket"`
	Title  string `json:"title"`
	Entry  string `json:"entry"`
}

func (s *Server) handleTicketsChangelogPost(w http.ResponseWriter, r *http.Request) error {
	var req ticketChangelogAppendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return NewHTTPError(http.StatusBadRequest, "invalid_argument", "invalid json body", nil)
	}
	req.Ticket = strings.TrimSpace(req.Ticket)
	if req.Ticket == "" {
		return NewHTTPError(http.StatusBadRequest, "invalid_argument", "missing ticket", map[string]any{"field": "ticket"})
	}
	if strings.TrimSpace(req.Entry) == "" {
		return NewHTTPError(http.StatusBadRequest, "invalid_argument", "missing entry", map[string]any{"field": "entry"})
	}

	var resp map[string]any
	if err := s.mgr.WithWorkspace(func(ws *workspace.Workspace) error {
		res, err := resolveTicketOrHTTPError(r, ws, req.Ticket)
		if err != nil {
			return err
		}

		relPath := filepath.ToSlash(filepath.Join(res.TicketDirRel, "changelog.md"))
		absPath := filepath.Join(ws.Context().Root, filepath.FromSlash(res.TicketDirRel), "changelog.md")
		date, err := commands.AppendChangelogEntry(absPath, req.Title, req.Entry, nil)
		if err != nil {
			return err
		}
		resp = map[string]any{
			"ok":     true,
			"ticket": req.Ticket,
			"path":   relPath,
			"date":   date,
		}
		return nil
	}); err != nil {
		return err
	}

	if _, err := s.mgr.Refresh(r.Context()); err != nil {
		return err
	}

	return writeJSON(w, http.StatusOK, resp)
}

// resolveTicketOrHTTPError wraps tickets.Resolve mapping not-found/ambiguous
// to the HTTP error shape used across ticket handlers.
func resolveTicketOrHTTPError(r *http.Request, ws *workspace.Workspace, ticketID string) (tickets.Resolution, error) {
	res, err := tickets.Resolve(r.Context(), ws, ticketID)
	if err != nil {
		if errors.Is(err, tickets.ErrNotFound) {
			return tickets.Resolution{}, NewHTTPError(http.StatusNotFound, "not_found", "ticket not found", map[string]any{"field": "ticket", "value": ticketID})
		}
		if errors.Is(err, tickets.ErrAmbiguous) {
			return tickets.Resolution{}, NewHTTPError(http.StatusBadRequest, "ambiguous", "ticket is ambiguous", map[string]any{"field": "ticket", "value": ticketID})
		}
		return tickets.Resolution{}, err
	}
	return res, nil
}
