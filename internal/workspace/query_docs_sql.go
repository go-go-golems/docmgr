package workspace

import (
	"context"
	"path/filepath"
	"sort"
	"strings"

	"github.com/go-go-golems/docmgr/internal/paths"
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
	var joins []string

	// Visibility defaults: hide tagged categories unless explicitly included.
	if !q.Options.IncludeArchivedPath {
		where = append(where, "d.is_archived_path = 0")
	}
	if !q.Options.IncludeScriptsPath {
		where = append(where, "d.is_scripts_path = 0")
	}
	if !q.Options.IncludeSourcesPath {
		where = append(where, "d.is_sources_path = 0")
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
		n := w.resolver.NormalizeNoFS(strings.TrimSpace(q.Scope.DocPath))
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
	if strings.TrimSpace(q.Filters.Intent) != "" {
		where = append(where, "d.intent = ?")
		args = append(args, strings.TrimSpace(q.Filters.Intent))
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

	// OwnersAny: OR semantics (any owner matches).
	if owners := normalizeLowerList(q.Filters.OwnersAny); len(owners) > 0 {
		clause, cargs := existsInClause(
			`SELECT 1 FROM doc_owners o WHERE o.doc_id = d.doc_id AND `,
			"o.owner_lower",
			toAny(owners),
		)
		where = append(where, clause)
		args = append(args, cargs...)
	}

	// RelatedFile: OR semantics (any file matches).
	if len(q.Filters.RelatedFile) > 0 {
		keys := buildQueryPathKeySet(w, q.Filters.RelatedFile)
		suffixes := buildQueryFileSuffixPatterns(q.Filters.RelatedFile)
		if len(keys) > 0 || len(suffixes) > 0 {
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

	// TextQuery: FTS-backed search (no compatibility guarantees).
	if strings.TrimSpace(q.Filters.TextQuery) != "" {
		joins = append(joins, "JOIN docs_fts ON docs_fts.rowid = d.doc_id")
		where = append(where, "docs_fts MATCH ?")
		args = append(args, strings.TrimSpace(q.Filters.TextQuery))
	}

	// Order.
	orderExpr := ""
	switch q.Options.OrderBy {
	case OrderByPath:
		orderExpr = "d.path"
	case OrderByLastUpdated:
		orderExpr = "d.last_updated"
	case OrderByRank:
		if strings.TrimSpace(q.Filters.TextQuery) == "" {
			return compiledSQL{}, errors.New("OrderByRank requires a non-empty TextQuery")
		}
		orderExpr = "bm25(docs_fts)"
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
  d.what_for,
  d.when_to_use,
  d.parse_ok,
  d.parse_err,
  d.body
FROM docs d
`)
	if len(joins) > 0 {
		sql.WriteString(strings.Join(joins, "\n"))
		sql.WriteString("\n")
	}
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

// buildQueryPathKeySet resolves query inputs to absolute paths via the one
// resolver; reverse lookup matches exactly on related_files.norm_abs.
func buildQueryPathKeySet(w *Workspace, rawPaths []string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, raw := range rawPaths {
		k := queryPathAbsKey(w.resolver, raw)
		if k == "" {
			continue
		}
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		out = append(out, k)
	}
	return out
}

// buildQueryDirPrefixes builds case-sensitive GLOB patterns matching files
// under the queried directory: the resolved absolute prefix plus (for
// relative, non-anchored queries) a whole-segment infix pattern.
func buildQueryDirPrefixes(w *Workspace, rawDirs []string) []string {
	seen := map[string]struct{}{}
	add := func(p string) {
		if p == "" {
			return
		}
		if _, ok := seen[p]; ok {
			return
		}
		seen[p] = struct{}{}
	}
	for _, raw := range rawDirs {
		if abs := queryPathAbsKey(w.resolver, raw); abs != "" {
			add(escapeGlob(strings.TrimRight(abs, "/")) + "/*")
		}
		if rel := querySuffixRel(raw); rel != "" {
			add("*/" + escapeGlob(rel) + "/*")
		}
	}
	out := make([]string, 0, len(seen))
	for p := range seen {
		out = append(out, p)
	}
	sortStrings(out)
	return out
}

// buildQueryFileSuffixPatterns builds case-sensitive whole-segment suffix
// GLOB patterns for relative (non-anchored, non-absolute) queries: the query's
// full path must equal a trailing run of the stored absolute path's segments.
// Substring containment is intentionally not supported ("api.go" must not
// match "chatapi.go").
func buildQueryFileSuffixPatterns(rawPaths []string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, raw := range rawPaths {
		rel := querySuffixRel(raw)
		if rel == "" {
			continue
		}
		p := "*/" + escapeGlob(rel)
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		out = append(out, p)
	}
	return out
}

// querySuffixRel returns the cleaned relative form of a query usable for
// whole-segment suffix matching, or "" when the query is absolute/anchored.
func querySuffixRel(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if paths.IsAnchored(raw) {
		return ""
	}
	if filepath.IsAbs(filepath.FromSlash(raw)) {
		return ""
	}
	rel := filepath.ToSlash(filepath.Clean(filepath.FromSlash(raw)))
	rel = strings.Trim(rel, "/")
	if rel == "" || rel == "." || rel == ".." || strings.HasPrefix(rel, "../") {
		return ""
	}
	return rel
}

// queryPathAbsKey resolves a query input to its absolute path key without
// touching the filesystem. Query filters are lookup keys only; existence-based
// anchor selection belongs to indexing persisted RelatedFiles, not to HTTP/CLI
// search inputs.
func queryPathAbsKey(resolver *paths.Resolver, raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" || resolver == nil {
		return ""
	}
	n := resolver.ResolveNoFS(raw)
	return filepath.ToSlash(strings.TrimSpace(n.Abs))
}

func relatedFileExistsClause(keys []string, suffixPatterns []string) (string, []any) {
	var ors []string
	var args []any
	if len(keys) > 0 {
		ors = append(ors, "rf.norm_abs IN ("+makePlaceholders(len(keys))+")")
		for _, k := range keys {
			args = append(args, k)
		}
	}
	for _, p := range suffixPatterns {
		// GLOB is case-sensitive (unlike LIKE).
		ors = append(ors, "rf.norm_abs GLOB ?")
		args = append(args, p)
	}
	if len(ors) == 0 {
		return "1=0", nil
	}
	return "EXISTS (SELECT 1 FROM related_files rf WHERE rf.doc_id = d.doc_id AND (" + strings.Join(ors, " OR ") + "))", args
}

func relatedDirExistsClause(prefixes []string) (string, []any) {
	var ors []string
	var args []any
	for _, p := range prefixes {
		ors = append(ors, "rf.norm_abs GLOB ?")
		args = append(args, p)
	}
	if len(ors) == 0 {
		return "1=0", nil
	}
	return "EXISTS (SELECT 1 FROM related_files rf WHERE rf.doc_id = d.doc_id AND (" + strings.Join(ors, " OR ") + "))", args
}

// escapeGlob escapes SQLite GLOB metacharacters so pattern parts derived from
// user paths match literally.
func escapeGlob(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch r {
		case '*', '?', '[':
			b.WriteRune('[')
			b.WriteRune(r)
			b.WriteRune(']')
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

func sortStrings(values []string) {
	sort.Strings(values)
}
