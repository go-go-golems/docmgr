package workspace

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"sort"
	"strings"

	docpkg "github.com/go-go-golems/docmgr/pkg/doc"
	"github.com/pkg/errors"
)

// ExportSQLiteOptions controls the export behavior.
type ExportSQLiteOptions struct {
	OutPath string
	Force   bool
}

// ExportIndexToSQLiteFile exports the current in-memory workspace index into a persistent SQLite file.
//
// It additionally creates and populates a README table containing docmgr's embedded documentation
// (`pkg/doc/*.md`) so that shared DB files are self-describing for debugging and introspection.
//
// NOTE: This requires the workspace index to already be initialized (`InitIndex`).
func (w *Workspace) ExportIndexToSQLiteFile(ctx context.Context, opts ExportSQLiteOptions) error {
	if ctx == nil {
		return errors.New("nil context")
	}
	if opts.OutPath == "" {
		return errors.New("missing OutPath")
	}
	if w.db == nil {
		return errors.New("workspace index not initialized (db is nil); call InitIndex first")
	}

	outAbs, err := filepath.Abs(opts.OutPath)
	if err != nil {
		outAbs = opts.OutPath
	}
	outAbs = filepath.Clean(outAbs)

	// Require existing parent dir (per user preference: do not mkdir).
	parent := filepath.Dir(outAbs)
	if parent == "" || parent == "." {
		parent = "."
	}
	if st, err := os.Stat(parent); err != nil || !st.IsDir() {
		return errors.Errorf("output directory does not exist: %s", parent)
	}

	if _, err := os.Stat(outAbs); err == nil {
		if !opts.Force {
			return errors.Errorf("output file already exists (use --force): %s", outAbs)
		}
		if err := os.Remove(outAbs); err != nil {
			return errors.Wrap(err, "remove existing output file")
		}
	}

	if err := ensureReadmeTable(ctx, w.db); err != nil {
		return err
	}
	if err := populateReadmeTable(ctx, w.db); err != nil {
		return err
	}

	// Use VACUUM INTO for a consistent single-file snapshot.
	lit, err := sqliteQuoteStringLiteral(filepath.ToSlash(outAbs))
	if err != nil {
		return err
	}
	// #nosec G202 -- VACUUM INTO cannot be parameterized; `lit` is a properly quoted SQLite string literal.
	_, err = w.db.ExecContext(ctx, "VACUUM INTO "+lit)
	if err != nil {
		return errors.Wrap(err, "vacuum into output sqlite file")
	}
	return nil
}

func ensureReadmeTable(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS README (
    name TEXT PRIMARY KEY,
    content TEXT NOT NULL
);
`)
	return errors.Wrap(err, "create README table")
}

func populateReadmeTable(ctx context.Context, db *sql.DB) error {
	docs, err := docpkg.ReadEmbeddedMarkdownDocs()
	if err != nil {
		return errors.Wrap(err, "read embedded markdown docs")
	}

	// Make insertion deterministic.
	sort.Slice(docs, func(i, j int) bool { return docs[i].Name < docs[j].Name })

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "begin readme tx")
	}
	defer func() { _ = tx.Rollback() }()

	// Replace content on each export so the file is self-contained.
	if _, err := tx.ExecContext(ctx, `DELETE FROM README`); err != nil {
		return errors.Wrap(err, "clear README table")
	}

	stmt, err := tx.PrepareContext(ctx, `INSERT INTO README(name, content) VALUES(?, ?);`)
	if err != nil {
		return errors.Wrap(err, "prepare README insert")
	}
	defer func() { _ = stmt.Close() }()

	about := strings.TrimSpace(
		"# Docmgr Workspace Export (SQLite)\n\n" +
			"This SQLite file is an exported snapshot of docmgr's Workspace index, intended for debugging and sharing.\n\n" +
			"It contains:\n" +
			"- Workspace index tables (docs, doc_topics, related_files, ...)\n" +
			"- README table with docmgr embedded docs (so this DB is self-describing)\n\n" +
			"Quick queries:\n\n" +
			"  -- List tables\n" +
			"  SELECT name FROM sqlite_master WHERE type='table' ORDER BY name;\n\n" +
			"  -- Show embedded docs available\n" +
			"  SELECT name, length(content) AS bytes FROM README ORDER BY name;\n\n" +
			"  -- Inspect a doc\n" +
			"  SELECT content FROM README WHERE name='docmgr-how-to-use.md';\n",
	)
	if _, err := stmt.ExecContext(ctx, "__about__.md", about); err != nil {
		return errors.Wrap(err, "insert README about doc")
	}

	for _, d := range docs {
		if strings.TrimSpace(d.Content) == "" {
			continue
		}
		if _, err := stmt.ExecContext(ctx, d.Name, d.Content); err != nil {
			return errors.Wrap(err, "insert README doc")
		}
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "commit readme tx")
	}
	return nil
}
