package tickets

import (
	"context"
	"errors"
	"path/filepath"
	"regexp"
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

func Resolve(ctx context.Context, ws *workspace.Workspace, ticketID string) (Resolution, error) {
	ticketID = strings.TrimSpace(ticketID)
	if ticketID == "" {
		return Resolution{}, errors.New("missing ticket id")
	}
	if ws == nil {
		return Resolution{}, errors.New("nil workspace")
	}

	res, err := ws.QueryDocs(ctx, workspace.DocQuery{
		Scope: workspace.Scope{Kind: workspace.ScopeRepo},
		Filters: workspace.DocFilters{
			Ticket:  ticketID,
			DocType: "index",
		},
		Options: workspace.DocQueryOptions{
			IncludeBody:         false,
			IncludeErrors:       false,
			IncludeDiagnostics:  false,
			IncludeArchivedPath: true,
			IncludeScriptsPath:  true,
			IncludeControlDocs:  true,
			OrderBy:             workspace.OrderByPath,
		},
	})
	if err != nil {
		return Resolution{}, err
	}

	var matches []workspace.DocHandle
	for _, h := range res.Docs {
		if h.Doc == nil {
			continue
		}
		if strings.TrimSpace(h.Doc.Ticket) != ticketID {
			continue
		}
		if strings.TrimSpace(h.Doc.DocType) != "index" {
			continue
		}
		matches = append(matches, h)
	}

	if len(matches) == 0 {
		return Resolution{}, ErrNotFound
	}
	if len(matches) > 1 {
		return Resolution{}, ErrAmbiguous
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
