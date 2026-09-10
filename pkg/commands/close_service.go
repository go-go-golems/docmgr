package commands

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"time"

	"github.com/go-go-golems/docmgr/internal/documents"
	"github.com/go-go-golems/docmgr/internal/operations"
)

type CloseRequest struct {
	Ticket string `json:"ticket"`
	Status string `json:"status"`
	Intent string `json:"intent"`
	Entry  string `json:"entry"`
}
type CloseResult struct {
	Receipt   operations.Receipt `json:"receipt"`
	Status    string             `json:"status"`
	Intent    string             `json:"intent"`
	OpenTasks int                `json:"open_tasks"`
	DoneTasks int                `json:"done_tasks"`
}

func CloseTicket(ctx context.Context, dir, id string, req CloseRequest) (CloseResult, error) {
	return closeTicket(ctx, dir, id, req, operations.Store{})
}
func closeTicket(ctx context.Context, dir, id string, req CloseRequest, store operations.Store) (CloseResult, error) {
	result := CloseResult{}
	if req.Status == "" {
		req.Status = "complete"
	}
	if req.Entry == "" {
		req.Entry = "Ticket closed"
	}
	if id == "" {
		unchanged := false
		err := operations.WithLock(ctx, dir, func() error {
			records, err := operations.Records(dir)
			if err != nil {
				return err
			}
			for _, r := range records {
				if r.Receipt.State != "committed" {
					return fmt.Errorf("pending operation %s; retry with --operation-id and original request", r.Receipt.OperationID)
				}
			}
			doc, _, err := documents.ReadDocumentWithFrontmatter(filepath.Join(dir, "index.md"))
			if err != nil {
				return err
			}
			if doc.Ticket != req.Ticket {
				return fmt.Errorf("ticket identity mismatch")
			}
			result.Status = doc.Status
			result.Intent = doc.Intent
			unchanged = doc.Status == req.Status && (req.Intent == "" || doc.Intent == req.Intent)
			return nil
		})
		if err != nil {
			return result, err
		}
		if unchanged {
			result.Receipt.State = "unchanged"
			result.OpenTasks, result.DoneTasks = countTasksInTicket(dir)
			return result, nil
		}
		var nonce [16]byte
		if _, err := rand.Read(nonce[:]); err != nil {
			return result, err
		}
		id = "close-" + hex.EncodeToString(nonce[:])
	}
	receipt, err := store.Execute(ctx, dir, id, req, false, func() ([]operations.Change, error) {
		_, before, err := operations.Read(dir, "index.md")
		if err != nil {
			return nil, err
		}
		doc, body, err := documents.ReadDocumentWithFrontmatter(filepath.Join(dir, "index.md"))
		if err != nil {
			return nil, err
		}
		if doc.Ticket != req.Ticket {
			return nil, fmt.Errorf("ticket identity mismatch")
		}
		result.Status = doc.Status
		result.Intent = doc.Intent
		if doc.Status == req.Status && (req.Intent == "" || doc.Intent == req.Intent) {
			return []operations.Change{}, nil
		}
		now := time.Now()
		if store.Now != nil {
			now = store.Now()
		}
		doc.Status = req.Status
		if req.Intent != "" {
			doc.Intent = req.Intent
		}
		doc.LastUpdated = now
		after, err := documents.SerializeDocument(doc, body)
		if err != nil {
			return nil, err
		}
		old, logHash, err := operations.Read(dir, "changelog.md")
		if err != nil {
			return nil, err
		}
		log := BuildChangelogEntry(string(old), now.Format("2006-01-02"), "", req.Entry, nil)
		// Status is the last projection, so predictable history failures cannot mark
		// the ticket complete. A later failure remains visible in the journal.
		return []operations.Change{{Path: "changelog.md", Before: logHash, After: []byte(log)}, {Path: "index.md", Before: before, After: after}}, nil
	})
	result.Receipt = receipt
	if err == nil {
		doc, _, readErr := documents.ReadDocumentWithFrontmatter(filepath.Join(dir, "index.md"))
		if readErr != nil {
			return result, readErr
		}
		result.Status = doc.Status
		result.Intent = doc.Intent
	}
	result.OpenTasks, result.DoneTasks = countTasksInTicket(dir)
	return result, err
}
