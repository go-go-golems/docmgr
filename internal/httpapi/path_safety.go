package httpapi

import (
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

	absTarget := cleaned
	if !filepath.IsAbs(absTarget) {
		absTarget = filepath.Join(rootAbs, absTarget)
	}
	absTarget = filepath.Clean(absTarget)

	// Cheap early traversal check before we touch the filesystem.
	if _, ok := tryRelWithin(rootAbs, absTarget); !ok {
		return "", "", nil, NewHTTPError(http.StatusForbidden, "forbidden", "path escapes allowed root", map[string]any{
			"field": "path",
		})
	}

	fi, err := os.Lstat(absTarget)
	if err != nil {
		if os.IsNotExist(err) {
			return "", "", nil, NewHTTPError(http.StatusNotFound, "not_found", "file not found", map[string]any{
				"field": "path",
				"value": rawPath,
			})
		}
		return "", "", nil, err
	}

	absEval, err := filepath.EvalSymlinks(absTarget)
	if err != nil {
		return "", "", nil, err
	}

	if _, ok := tryRelWithin(rootEval, absEval); !ok {
		return "", "", nil, NewHTTPError(http.StatusForbidden, "forbidden", "path escapes allowed root", map[string]any{
			"field": "path",
		})
	}

	rel, _ := filepath.Rel(rootAbs, absTarget)
	rel = filepath.ToSlash(rel)

	return absTarget, rel, fi, nil
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
