package httpapi

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/go-go-golems/docmgr/internal/workspace"
	"github.com/go-go-golems/docmgr/pkg/commands"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/types"
)

type doctorFinding struct {
	Ticket   string `json:"ticket"`
	Issue    string `json:"issue"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
	Path     string `json:"path"`
}

type doctorRollupItem struct {
	Ticket   string `json:"ticket"`
	Errors   int    `json:"errors"`
	Warnings int    `json:"warnings"`
	Infos    int    `json:"infos"`
	Status   string `json:"status"`
}

type doctorTotals struct {
	Findings       int `json:"findings"`
	Errors         int `json:"errors"`
	Warnings       int `json:"warnings"`
	Infos          int `json:"infos"`
	TicketsChecked int `json:"ticketsChecked"`
}

type doctorResponse struct {
	Ticket   string             `json:"ticket"`
	Totals   doctorTotals       `json:"totals"`
	Rollup   []doctorRollupItem `json:"rollup"`
	Findings []doctorFinding    `json:"findings"`
}

// doctorRowCollector collects glazed rows emitted by the doctor command.
type doctorRowCollector struct {
	rows []types.Row
}

func (c *doctorRowCollector) AddRow(_ context.Context, row types.Row) error {
	c.rows = append(c.rows, row)
	return nil
}

func (c *doctorRowCollector) Close(_ context.Context) error { return nil }

// handleWorkspaceDoctor wraps the 'docmgr doctor' scan (read-only: no --fix)
// and returns its findings as JSON: a per-ticket rollup plus the per-finding
// list, using the same row model as the CLI.
func (s *Server) handleWorkspaceDoctor(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodGet {
		return NewHTTPError(http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", nil)
	}

	ticketParam := strings.TrimSpace(r.URL.Query().Get("ticket"))
	staleAfter := parseIntDefault(r.URL.Query().Get("staleAfter"), 30)
	if staleAfter <= 0 {
		staleAfter = 30
	}

	var root string
	var ticketID string
	if err := s.mgr.WithWorkspace(func(ws *workspace.Workspace) error {
		root = ws.Context().Root
		if ticketParam == "" {
			return nil
		}
		res, err := resolveTicketOrHTTPError(r, ws, ticketParam)
		if err != nil {
			return err
		}
		ticketID = res.TicketID
		return nil
	}); err != nil {
		return err
	}

	cmd, err := commands.NewDoctorCommand()
	if err != nil {
		return err
	}
	defaultSection, ok := cmd.GetDefaultSection()
	if !ok {
		return fmt.Errorf("doctor command missing default section")
	}
	sectionValues, err := values.NewSectionValues(
		defaultSection,
		values.WithFieldValue("root", root),
		values.WithFieldValue("ticket", ticketID),
		values.WithFieldValue("all", ticketID == ""),
		values.WithFieldValue("stale-after", staleAfter),
		values.WithFieldValue("fail-on", "none"),
	)
	if err != nil {
		return err
	}
	parsed := values.New()
	parsed.Set(schema.DefaultSlug, sectionValues)

	collector := &doctorRowCollector{}
	if err := cmd.RunIntoGlazeProcessor(r.Context(), parsed, collector); err != nil {
		return err
	}

	resp := doctorResponse{
		Ticket:   ticketID,
		Rollup:   []doctorRollupItem{},
		Findings: make([]doctorFinding, 0, len(collector.rows)),
	}
	rollup := map[string]*doctorRollupItem{}
	for _, row := range collector.rows {
		f := doctorFinding{
			Ticket:   rowString(row, "ticket"),
			Issue:    rowString(row, "issue"),
			Severity: rowString(row, "severity"),
			Message:  rowString(row, "message"),
			Path:     rowString(row, "path"),
		}
		resp.Findings = append(resp.Findings, f)
		resp.Totals.Findings++

		item, ok := rollup[f.Ticket]
		if !ok {
			item = &doctorRollupItem{Ticket: f.Ticket, Status: "ok"}
			rollup[f.Ticket] = item
		}
		switch f.Severity {
		case "error":
			resp.Totals.Errors++
			item.Errors++
			item.Status = "error"
		case "warning":
			resp.Totals.Warnings++
			item.Warnings++
			if item.Status != "error" {
				item.Status = "warning"
			}
		case "info":
			resp.Totals.Infos++
			item.Infos++
		}
	}

	keys := make([]string, 0, len(rollup))
	for k := range rollup {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		resp.Rollup = append(resp.Rollup, *rollup[k])
	}
	resp.Totals.TicketsChecked = len(resp.Rollup)

	return writeJSON(w, http.StatusOK, resp)
}

func rowString(row types.Row, key string) string {
	v, ok := row.Get(key)
	if !ok || v == nil {
		return ""
	}
	return fmt.Sprint(v)
}
