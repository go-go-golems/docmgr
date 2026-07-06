package httpapi

import (
	"errors"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-go-golems/docmgr/internal/workspace"
)

// maxRawFileBytes caps GET /files/raw responses (~20MB): large enough for
// screenshots and design assets, small enough to keep the local server snappy.
const maxRawFileBytes int64 = 20 * 1024 * 1024

// handleFilesRaw streams a file's bytes with a sniffed content type. Unlike
// /files/get it has no JSON envelope and serves binary content (images, ...),
// so relative images inside rendered markdown can load. Path resolution goes
// through the same traversal-safe resolveFileWithin as /files/get.
func (s *Server) handleFilesRaw(w http.ResponseWriter, r *http.Request) error {
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

	return s.mgr.WithWorkspace(func(ws *workspace.Workspace) error {
		wctx := ws.Context()

		var rootDir string
		switch rootParam {
		case "docs":
			rootDir = wctx.Root
		default:
			rootDir = wctx.RepoRoot
		}

		abs, rel, fi, err := resolveFileWithin(rootDir, rawPath)
		if err != nil {
			return err
		}
		if fi.IsDir() {
			return NewHTTPError(http.StatusBadRequest, "invalid_argument", "path is a directory", map[string]any{
				"field": "path",
				"value": rawPath,
			})
		}
		if fi.Size() > maxRawFileBytes {
			return NewHTTPError(http.StatusRequestEntityTooLarge, "too_large", "file exceeds raw serving size cap", map[string]any{
				"field":     "path",
				"sizeBytes": fi.Size(),
				"maxBytes":  maxRawFileBytes,
			})
		}

		rootAbs, err := filepath.Abs(rootDir)
		if err != nil {
			return err
		}
		fsys := os.DirFS(filepath.Clean(rootAbs))
		f, err := fsys.Open(rel)
		if err != nil {
			return err
		}
		defer func() { _ = f.Close() }()

		// Sniff the content type from the first bytes when the extension is
		// not enough (e.g. extension-less assets).
		ct := mime.TypeByExtension(filepath.Ext(abs))
		head := make([]byte, 512)
		n, readErr := io.ReadFull(f, head)
		if readErr != nil && !errors.Is(readErr, io.EOF) && !errors.Is(readErr, io.ErrUnexpectedEOF) {
			return readErr
		}
		head = head[:n]
		if ct == "" {
			ct = http.DetectContentType(head)
		}

		w.Header().Set("Content-Type", ct)
		w.Header().Set("Content-Length", strconv.FormatInt(fi.Size(), 10))
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(head); err != nil {
			return nil // client went away; response already committed
		}
		_, _ = io.Copy(w, f)
		return nil
	})
}
