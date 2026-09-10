package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"path/filepath"

	"github.com/go-go-golems/docmgr/internal/operations"
	"github.com/go-go-golems/docmgr/internal/workspace"
	"github.com/go-go-golems/docmgr/pkg/commands"
)

func (s *Server) handleTicketMilestone(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPost {
		return NewHTTPError(405, "method_not_allowed", "method not allowed", nil)
	}
	var req commands.MilestoneRequest
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 65536))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		return NewHTTPError(400, "invalid_argument", err.Error(), nil)
	}
	var extra any
	if err := dec.Decode(&extra); err != io.EOF {
		return NewHTTPError(400, "invalid_argument", "expected one JSON object", nil)
	}
	if req.Ticket == "" {
		return NewHTTPError(400, "invalid_argument", "missing ticket", nil)
	}
	dry := r.URL.Query().Get("dry_run") == "true"
	var receipt operations.Receipt
	err := s.mgr.WithWorkspace(func(ws *workspace.Workspace) error {
		res, err := resolveTicketOrHTTPError(r, ws, req.Ticket)
		if err != nil {
			return err
		}
		receipt, err = commands.RecordMilestone(r.Context(), filepath.Join(ws.Context().Root, res.TicketDirRel), req, dry)
		return err
	})
	if err != nil {
		var he *HTTPError
		if errors.As(err, &he) {
			return err
		}
		details := map[string]any{"receipt": receipt}
		var partial *operations.ApplyError
		if errors.As(err, &partial) {
			details["applied"] = partial.Applied
		}
		return NewHTTPError(409, "operation_conflict", err.Error(), details)
	}
	warnings := []string{}
	if !dry {
		if _, err := s.mgr.Refresh(r.Context()); err != nil {
			warnings = append(warnings, "operation committed; index refresh failed: "+err.Error())
		}
	}
	return writeJSON(w, 200, map[string]any{"receipt": receipt, "warnings": warnings})
}
func (s *Server) handleTicketResume(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodGet {
		return NewHTTPError(405, "method_not_allowed", "method not allowed", nil)
	}
	var view commands.ResumeView
	err := s.mgr.WithWorkspace(func(ws *workspace.Workspace) error {
		res, err := resolveTicketOrHTTPError(r, ws, r.URL.Query().Get("ticket"))
		if err != nil {
			return err
		}
		view, err = commands.TicketResume(r.Context(), filepath.Join(ws.Context().Root, res.TicketDirRel))
		return err
	})
	if err != nil {
		return err
	}
	return writeJSON(w, 200, view)
}
