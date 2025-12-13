package workspace

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

type compiledSQL struct {
	SQL  string
	Args []any
}

func compileDocQuery(ctx context.Context, w *Workspace, q DocQuery) (compiledSQL, error) {
	if q.Options.IncludeErrors {
		return compileDocQueryWithParseFilter(ctx, w, q, nil)
	}
	one := 1
	return compileDocQueryWithParseFilter(ctx, w, q, &one)
}

func compileDocQueryWithParseFilter(ctx context.Context, w *Workspace, q DocQuery, parseOKFilter *int) (compiledSQL, error) {
	_ = ctx // reserved for future compilation-time diagnostics

	var where []string
	var args []any

	// Visibility defaults: hide tagged categories unless explicitly included.
	if !q.Options.IncludeArchivedPath {
		where = append(where, "d.is_archived_path = 0")
	}
	if !q.Options.IncludeScriptsPath {
		where = append(where, "d.is_scripts_path = 0")
	}
	if !q.Options.IncludeControlDocs {
		where = append(where, "d.is_control_doc = 0")
	}

	// Parse state filter (optional).
	if parseOKFilter != nil {
		where = append(where, "d.parse_ok = ?")
		args = append(args, *parseOKFilter)
	}

	// Scope.
	switch q.Scope.Kind {
	case ScopeRepo:
	case ScopeTicket:
		where = append(where, "d.ticket_id = ?")
		args = append(args, strings.TrimSpace(q.Scope.TicketID))
	case ScopeDoc:
		n := w.resolver.Normalize(strings.TrimSpace(q.Scope.DocPath))
		if strings.TrimSpace(n.Abs) == "" {
			return compiledSQL{}, errors.Errorf("failed to normalize DocPath: %q", q.Scope.DocPath)
		}
		abs := filepath.ToSlash(filepath.Clean(n.Abs))
		where = append(where, "d.path = ?")
		args = append(args, abs)
	default:
		return compiledSQL{}, errors.Errorf("unknown scope kind: %d", q.Scope.Kind)
	}

	// Filters.
	if strings.TrimSpace(q.Filters.Ticket) != "" && q.Scope.Kind != ScopeTicket {
		where = append(where, "d.ticket_id = ?")
		args = append(args, strings.TrimSpace(q.Filters.Ticket))
	}
	if strings.TrimSpace(q.Filters.DocType) != "" {
		where = append(where, "d.doc_type = ?")
		args = append(args, strings.TrimSpace(q.Filters.DocType))
	}
	if strings.TrimSpace(q.Filters.Status) != "" {
		where = append(where, "d.status = ?")
		args = append(args, strings.TrimSpace(q.Filters.Status))
	}

	// TopicsAny: OR semantics (any topic matches).
	if topics := normalizeLowerList(q.Filters.TopicsAny); len(topics) > 0 {
		clause, cargs := existsInClause(
			`SELECT 1 FROM doc_topics t WHERE t.doc_id = d.doc_id AND `,
			"t.topic_lower",
			toAny(topics),
		)
		where = append(where, clause)
		args = append(args, cargs...)
	}

	// RelatedFile: OR semantics (any file matches).
	if len(q.Filters.RelatedFile) > 0 {
		keys := buildQueryPathKeySet(w, q.Filters.RelatedFile)
		suffixes := buildQueryFileSuffixPatterns(q.Filters.RelatedFile)
		if len(keys) > 0 {
			sub, subArgs := relatedFileExistsClause(keys, suffixes)
			where = append(where, sub)
			args = append(args, subArgs...)
		}
	}

	// RelatedDir: OR semantics (any dir matches).
	if len(q.Filters.RelatedDir) > 0 {
		prefixes := buildQueryDirPrefixes(w, q.Filters.RelatedDir)
		if len(prefixes) > 0 {
			sub, subArgs := relatedDirExistsClause(prefixes)
			where = append(where, sub)
			args = append(args, subArgs...)
		}
	}

	// Order.
	orderExpr := ""
	switch q.Options.OrderBy {
	case OrderByPath:
		orderExpr = "d.path"
	case OrderByLastUpdated:
		orderExpr = "d.last_updated"
	default:
		return compiledSQL{}, errors.Errorf("unsupported OrderBy: %q", q.Options.OrderBy)
	}
	orderDir := "ASC"
	if q.Options.Reverse {
		orderDir = "DESC"
	}

	sql := strings.Builder{}
	sql.WriteString(`SELECT
  d.doc_id,
  d.path,
  d.ticket_id,
  d.doc_type,
  d.status,
  d.intent,
  d.title,
  d.last_updated,
  d.parse_ok,
  d.parse_err,
  d.body
FROM docs d
`)
	if len(where) > 0 {
		sql.WriteString("WHERE ")
		sql.WriteString(strings.Join(where, " AND "))
		sql.WriteString("\n")
	}
	sql.WriteString("ORDER BY ")
	sql.WriteString(orderExpr)
	sql.WriteString(" ")
	sql.WriteString(orderDir)
	sql.WriteString(";\n")

	return compiledSQL{SQL: sql.String(), Args: args}, nil
}

func normalizeLowerList(values []string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, v := range values {
		v = strings.ToLower(strings.TrimSpace(v))
		if v == "" {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}

func toAny(values []string) []any {
	out := make([]any, 0, len(values))
	for _, v := range values {
		out = append(out, v)
	}
	return out
}

func existsInClause(prefixSQL string, col string, values []any) (string, []any) {
	// prefixSQL should end with the part before "<col> IN (...)"
	if len(values) == 0 {
		return "1=1", nil
	}
	in := makePlaceholders(len(values))
	clause := "EXISTS (" + prefixSQL + col + " IN (" + in + "))"
	return clause, values
}

func makePlaceholders(n int) string {
	if n <= 0 {
		return ""
	}
	if n == 1 {
		return "?"
	}
	return strings.Repeat("?,", n-1) + "?"
}

func buildQueryPathKeySet(w *Workspace, rawPaths []string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, raw := range rawPaths {
		for _, k := range queryPathKeys(w.resolver, raw) {
			k = filepath.ToSlash(strings.TrimSpace(k))
			if k == "" {
				continue
			}
			if _, ok := seen[k]; ok {
				continue
			}
			seen[k] = struct{}{}
			out = append(out, k)
		}
	}
	return out
}

func buildQueryDirPrefixes(w *Workspace, rawDirs []string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, raw := range rawDirs {
		keys := queryPathKeys(w.resolver, raw)
		for _, k := range keys {
			k = strings.Trim(filepath.ToSlash(strings.TrimSpace(k)), "/")
			if k == "" {
				continue
			}
			p := k + "/%"
			if _, ok := seen[p]; ok {
				continue
			}
			seen[p] = struct{}{}
			out = append(out, p)
		}
	}
	return out
}

func buildQueryFileSuffixPatterns(rawPaths []string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, raw := range rawPaths {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		// Only enable suffix matching for basename-like queries. This intentionally trades
		// precision for UX in "I only know the filename" scenarios (see scenario tests).
		if strings.Contains(raw, "/") || strings.Contains(raw, "\\") {
			continue
		}
		base := filepath.ToSlash(filepath.Clean(raw))
		if base == "" || base == "." || base == "/" {
			continue
		}
		p := "%/" + base
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		out = append(out, p)
	}
	return out
}

func relatedFileExistsClause(keys []string, suffixPatterns []string) (string, []any) {
	// We match against multiple persisted representations.
	cols := []string{
		"rf.norm_canonical",
		"rf.norm_repo_rel",
		"rf.norm_docs_rel",
		"rf.norm_doc_rel",
		"rf.norm_abs",
		"rf.norm_clean",
		"rf.raw_path",
	}
	in := makePlaceholders(len(keys))
	var parts []string
	var args []any
	for _, col := range cols {
		var ors []string
		ors = append(ors, col+" IN ("+in+")")
		for range suffixPatterns {
			ors = append(ors, col+" LIKE ?")
		}
		parts = append(parts, "("+strings.Join(ors, " OR ")+")")
		for _, k := range keys {
			args = append(args, k)
		}
		for _, p := range suffixPatterns {
			args = append(args, p)
		}
	}
	return "EXISTS (SELECT 1 FROM related_files rf WHERE rf.doc_id = d.doc_id AND (" + strings.Join(parts, " OR ") + "))", args
}

func relatedDirExistsClause(prefixes []string) (string, []any) {
	cols := []string{
		"rf.norm_canonical",
		"rf.norm_repo_rel",
		"rf.norm_docs_rel",
		"rf.norm_doc_rel",
		"rf.norm_abs",
		"rf.norm_clean",
		"rf.raw_path",
	}
	in := makePlaceholders(len(prefixes))
	_ = in // not used, but keep the pattern consistent with file clause
	var parts []string
	var args []any
	for _, col := range cols {
		// OR together LIKE patterns for each prefix.
		var likes []string
		for range prefixes {
			likes = append(likes, col+" LIKE ?")
		}
		parts = append(parts, "("+strings.Join(likes, " OR ")+")")
		for _, p := range prefixes {
			args = append(args, p)
		}
	}
	return "EXISTS (SELECT 1 FROM related_files rf WHERE rf.doc_id = d.doc_id AND (" + strings.Join(parts, " OR ") + "))", args
}
