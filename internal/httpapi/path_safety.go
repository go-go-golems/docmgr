package httpapi

import (
	"errors"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func resolveFileWithin(rootDir string, rawPath string) (string, string, os.FileInfo, error) {
	rawPath = strings.TrimSpace(rawPath)
	if rawPath == "" {
		return "", "", nil, NewHTTPError(http.StatusBadRequest, "invalid_argument", "missing path", map[string]any{
			"field": "path",
		})
	}
	if strings.ContainsRune(rawPath, 0) {
		return "", "", nil, NewHTTPError(http.StatusBadRequest, "invalid_argument", "invalid path", map[string]any{
			"field": "path",
		})
	}

	rootAbs, err := filepath.Abs(rootDir)
	if err != nil {
		return "", "", nil, err
	}
	rootAbs = filepath.Clean(rootAbs)

	rootEval := rootAbs
	if v, err := filepath.EvalSymlinks(rootAbs); err == nil {
		rootEval = v
	}

	cleaned := filepath.Clean(filepath.FromSlash(rawPath))

	// Convert the input into a stable path *relative to the root*.
	// This keeps the only "variable" component a validated relative path, even if
	// the user provided an absolute path.
	absInput := cleaned
	if !filepath.IsAbs(absInput) {
		absInput = filepath.Join(rootAbs, absInput)
	}
	absInput = filepath.Clean(absInput)

	// Cheap early traversal check before we touch the filesystem.
	relOS, ok := tryRelWithin(rootAbs, absInput)
	if !ok {
		return "", "", nil, NewHTTPError(http.StatusForbidden, "forbidden", "path escapes allowed root", map[string]any{
			"field": "path",
		})
	}
	if relOS == "." {
		return "", "", nil, NewHTTPError(http.StatusBadRequest, "invalid_argument", "path refers to a directory", map[string]any{
			"field": "path",
			"value": rawPath,
		})
	}
	relFS := filepath.ToSlash(relOS)
	if !fs.ValidPath(relFS) {
		return "", "", nil, NewHTTPError(http.StatusBadRequest, "invalid_argument", "invalid path", map[string]any{
			"field": "path",
			"value": rawPath,
		})
	}

	absTarget := filepath.Join(rootAbs, relOS)
	absTarget = filepath.Clean(absTarget)

	absEval, err := filepath.EvalSymlinks(absTarget)
	if err != nil {
		if os.IsNotExist(err) || errors.Is(err, fs.ErrNotExist) {
			return "", "", nil, NewHTTPError(http.StatusNotFound, "not_found", "file not found", map[string]any{
				"field": "path",
				"value": rawPath,
			})
		}
		return "", "", nil, err
	}

	if _, ok := tryRelWithin(rootEval, absEval); !ok {
		return "", "", nil, NewHTTPError(http.StatusForbidden, "forbidden", "path escapes allowed root", map[string]any{
			"field": "path",
		})
	}

	fi, err := fs.Stat(os.DirFS(rootAbs), relFS)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return "", "", nil, NewHTTPError(http.StatusNotFound, "not_found", "file not found", map[string]any{
				"field": "path",
				"value": rawPath,
			})
		}
		return "", "", nil, err
	}

	return absTarget, relFS, fi, nil
}

func tryRelWithin(rootAbs string, targetAbs string) (string, bool) {
	rel, err := filepath.Rel(rootAbs, targetAbs)
	if err != nil {
		return "", false
	}
	if rel == "." {
		return rel, true
	}
	if strings.HasPrefix(rel, ".."+string(filepath.Separator)) || rel == ".." {
		return "", false
	}
	return rel, true
}
