package workspace

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// PathTags are ingest-time derived tags stored on each doc row.
//
// Spec: §6 (skip rules + tagging).
type PathTags struct {
	IsIndex        bool
	IsArchivedPath bool
	IsScriptsPath  bool
	IsSourcesPath  bool
	IsControlDoc   bool
}

// DefaultIngestSkipDir is the canonical ingest-time directory skip predicate.
//
// Spec: §6.1 (directories).
// - Always skip `.meta/` entirely.
// - Always skip underscore dirs (`_*/`) entirely (templates, guidelines, etc.).
func DefaultIngestSkipDir(_ string, d fs.DirEntry) bool {
	name := d.Name()
	if name == ".meta" {
		return true
	}
	if name != "." && strings.HasPrefix(name, "_") {
		return true
	}
	return false
}

// ComputePathTags computes path tags for a markdown document at the given path.
//
// docPath is expected to be a filesystem path (typically absolute).
// The returned tags are designed to be stored into the SQLite index (`docs` table).
func ComputePathTags(docPath string) PathTags {
	slash := filepath.ToSlash(filepath.Clean(docPath))
	baseLower := strings.ToLower(filepath.Base(docPath))

	tags := PathTags{
		IsIndex: strings.EqualFold(filepath.Base(docPath), "index.md"),
		// Segment-based checks to avoid matching "myarchive" / "scripts-old", etc.
		IsArchivedPath: containsPathSegment(slash, "archive"),
		IsScriptsPath:  containsPathSegment(slash, "scripts"),
		IsSourcesPath:  containsPathSegment(slash, "sources"),
	}

	// Control docs are ticket-root files:
	// README.md, tasks.md, changelog.md — but only when they live alongside an index.md.
	if isControlDocBase(baseLower) && hasSiblingIndex(docPath) {
		tags.IsControlDoc = true
	}

	return tags
}

func isControlDocBase(baseLower string) bool {
	switch baseLower {
	case "readme.md", "tasks.md", "changelog.md":
		return true
	default:
		return false
	}
}

func hasSiblingIndex(docPath string) bool {
	dir := filepath.Dir(docPath)
	if dir == "" || dir == "." {
		return false
	}
	_, err := os.Stat(filepath.Join(dir, "index.md"))
	return err == nil
}

func containsPathSegment(slashPath string, seg string) bool {
	// Ensure we match whole segments with "/" boundaries.
	if seg == "" {
		return false
	}
	needle := "/" + seg + "/"
	if strings.Contains(slashPath, needle) {
		return true
	}
	// Also match when path ends with "/seg" (unlikely for a file path, but harmless).
	if strings.HasSuffix(slashPath, "/"+seg) {
		return true
	}
	return false
}
