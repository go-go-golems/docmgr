package httpapi

import (
	"errors"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/go-go-golems/docmgr/internal/documents"
	"github.com/go-go-golems/docmgr/internal/workspace"
	"github.com/go-go-golems/docmgr/pkg/diagnostics/core"
	"github.com/go-go-golems/docmgr/pkg/models"
)

const (
	maxServedTextBytes = 2 * 1024 * 1024
)

type fileStats struct {
	SizeBytes int64  `json:"sizeBytes"`
	ModTime   string `json:"modTime"`
}

type docGetResponse struct {
	Path         string               `json:"path"`
	Doc          *models.Document     `json:"doc,omitempty"`
	RelatedFiles []models.RelatedFile `json:"relatedFiles"`
	Body         string               `json:"body"`
	Stats        fileStats            `json:"stats"`
	Diagnostic   *core.Taxonomy       `json:"diagnostic,omitempty"`
}

type fileGetResponse struct {
	Path        string    `json:"path"`
	Root        string    `json:"root"`
	Language    string    `json:"language"`
	ContentType string    `json:"contentType"`
	Truncated   bool      `json:"truncated"`
	Content     string    `json:"content"`
	Stats       fileStats `json:"stats"`
}

func (s *Server) handleDocsGet(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodGet {
		return NewHTTPError(http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", nil)
	}

	rawPath := strings.TrimSpace(r.URL.Query().Get("path"))
	if rawPath == "" {
		return NewHTTPError(http.StatusBadRequest, "invalid_argument", "missing path", map[string]any{
			"field": "path",
		})
	}

	var resp docGetResponse
	if err := s.mgr.WithWorkspace(func(ws *workspace.Workspace) error {
		ctx := ws.Context()
		abs, rel, fi, err := resolveFileWithin(ctx.Root, rawPath)
		if err != nil {
			return err
		}
		if fi.IsDir() {
			return NewHTTPError(http.StatusBadRequest, "invalid_argument", "path is a directory", map[string]any{
				"field": "path",
				"value": rawPath,
			})
		}

		doc, body, err := documents.ReadDocumentWithFrontmatter(abs)
		var diag *core.Taxonomy
		if err != nil {
			if t, ok := core.AsTaxonomy(err); ok {
				diag = t
			}

			raw, readErr := os.ReadFile(abs)
			if readErr != nil {
				return readErr
			}
			_, b, _, splitErr := documents.SplitFrontmatter(raw)
			if splitErr != nil {
				body = string(raw)
			} else {
				body = string(b)
			}
			doc = nil
		}

		resp = docGetResponse{
			Path: rel,
			Doc:  doc,
			RelatedFiles: func() []models.RelatedFile {
				if doc == nil {
					return []models.RelatedFile{}
				}
				return append([]models.RelatedFile{}, doc.RelatedFiles...)
			}(),
			Body: body,
			Stats: fileStats{
				SizeBytes: fi.Size(),
				ModTime:   fi.ModTime().Format(time.RFC3339Nano),
			},
			Diagnostic: diag,
		}
		return nil
	}); err != nil {
		return err
	}

	return writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleFilesGet(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodGet {
		return NewHTTPError(http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", nil)
	}

	rawPath := strings.TrimSpace(r.URL.Query().Get("path"))
	if rawPath == "" {
		return NewHTTPError(http.StatusBadRequest, "invalid_argument", "missing path", map[string]any{
			"field": "path",
		})
	}
	rootParam := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("root")))
	if rootParam == "" {
		rootParam = "repo"
	}
	if rootParam != "repo" && rootParam != "docs" {
		return NewHTTPError(http.StatusBadRequest, "invalid_argument", "invalid root", map[string]any{
			"field": "root",
			"value": rootParam,
		})
	}

	var resp fileGetResponse
	if err := s.mgr.WithWorkspace(func(ws *workspace.Workspace) error {
		wctx := ws.Context()

		var rootDir string
		switch rootParam {
		case "docs":
			rootDir = wctx.Root
		case "repo":
			rootDir = wctx.RepoRoot
		}

		abs, rel, _, err := resolveFileWithin(rootDir, rawPath)
		if err != nil {
			if rootParam == "repo" {
				if fallbackAbs, fallbackRel, ok := resolveViaWorkspace(ws, rawPath); ok {
					abs, rel, err = fallbackAbs, fallbackRel, nil
				}
			}
			if err != nil {
				return err
			}
		}

		content, truncated, err := readTextFile(abs, maxServedTextBytes)
		if err != nil {
			var he *HTTPError
			if errors.As(err, &he) {
				return he
			}
			return err
		}

		fi, err := os.Stat(abs)
		if err != nil {
			return err
		}

		ct := mime.TypeByExtension(filepath.Ext(abs))
		if ct == "" {
			ct = "text/plain; charset=utf-8"
		}
		resp = fileGetResponse{
			Path:        rel,
			Root:        rootParam,
			Language:    inferLanguage(abs),
			ContentType: ct,
			Truncated:   truncated,
			Content:     content,
			Stats: fileStats{
				SizeBytes: fi.Size(),
				ModTime:   fi.ModTime().Format(time.RFC3339Nano),
			},
		}
		return nil
	}); err != nil {
		return err
	}

	return writeJSON(w, http.StatusOK, resp)
}

func resolveViaWorkspace(ws *workspace.Workspace, raw string) (string, string, bool) {
	if ws == nil {
		return "", "", false
	}
	n := ws.Resolver().Normalize(raw)
	if strings.TrimSpace(n.Abs) == "" || !n.Exists {
		return "", "", false
	}
	abs := filepath.FromSlash(n.Abs)
	resAbs, resRel, _, err := resolveFileWithin(ws.Context().RepoRoot, abs)
	if err != nil {
		return "", "", false
	}
	return resAbs, resRel, true
}

func inferLanguage(path string) string {
	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(path)), ".")
	switch ext {
	case "go":
		return "go"
	case "ts", "tsx":
		return "typescript"
	case "js", "jsx":
		return "javascript"
	case "json":
		return "json"
	case "yaml", "yml":
		return "yaml"
	case "md":
		return "markdown"
	case "sh", "bash":
		return "bash"
	case "py":
		return "python"
	case "rb":
		return "ruby"
	case "rs":
		return "rust"
	case "java":
		return "java"
	case "c", "h":
		return "c"
	case "cpp", "cc", "cxx", "hpp":
		return "cpp"
	case "html", "htm":
		return "html"
	case "css":
		return "css"
	case "toml":
		return "toml"
	case "sql":
		return "sql"
	default:
		return ""
	}
}

func readTextFile(path string, maxBytes int64) (string, bool, error) {
	if maxBytes <= 0 {
		maxBytes = maxServedTextBytes
	}

	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", false, NewHTTPError(http.StatusNotFound, "not_found", "file not found", map[string]any{
				"field": "path",
			})
		}
		return "", false, err
	}
	defer func() { _ = f.Close() }()

	lr := &io.LimitedReader{R: f, N: maxBytes + 1}
	b, err := io.ReadAll(lr)
	if err != nil {
		return "", false, err
	}

	truncated := int64(len(b)) > maxBytes
	if truncated {
		b = b[:maxBytes]
	}

	if bytesContainNUL(b) {
		return "", false, NewHTTPError(http.StatusUnsupportedMediaType, "unsupported_media_type", "binary file not supported", map[string]any{
			"field": "path",
		})
	}
	if !utf8.Valid(b) {
		return "", false, NewHTTPError(http.StatusUnsupportedMediaType, "unsupported_media_type", "file is not valid UTF-8", map[string]any{
			"field": "path",
		})
	}

	return string(b), truncated, nil
}

func bytesContainNUL(b []byte) bool {
	for _, c := range b {
		if c == 0 {
			return true
		}
	}
	return false
}
