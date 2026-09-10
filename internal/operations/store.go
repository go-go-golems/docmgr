// Package operations implements bounded, recoverable ticket projections.
// It provides cooperative exclusion and process-crash recovery, not simultaneous
// multi-file visibility, power-loss durability, or isolation from external editors.
package operations

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"time"

	"github.com/go-go-golems/docmgr/internal/documents"
)

const MaxBytes = 4 << 20
const directory = ".docmgr-operations"

var idPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]{0,63}$`)

type Change struct {
	Path   string `json:"path"`
	Before string `json:"before"`
	After  []byte `json:"after"`
}
type Receipt struct {
	OperationID  string            `json:"operation_id"`
	State        string            `json:"state"`
	CreatedAt    time.Time         `json:"created_at"`
	ChangedPaths []string          `json:"changed_paths"`
	Before       map[string]string `json:"before"`
	After        map[string]string `json:"after"`
}
type Record struct {
	Version int             `json:"schema_version"`
	Digest  string          `json:"digest"`
	Request json.RawMessage `json:"request"`
	Receipt Receipt         `json:"receipt"`
	Changes []Change        `json:"changes"`
}

// ApplyError retains the journal identity and effects visible at the failure.
// Retrying the same request resumes unless a projection has conflicting bytes.
type ApplyError struct {
	Receipt Receipt
	Applied []string
	Cause   error
}

func (e *ApplyError) Error() string {
	return fmt.Sprintf("operation %s incomplete; applied=%v; retry same request: %v", e.Receipt.OperationID, e.Applied, e.Cause)
}
func (e *ApplyError) Unwrap() error { return e.Cause }

type Store struct {
	Now func() time.Time
	// BeforeWrite supports deterministic fault injection; nil in production.
	BeforeWrite func(string) error
}

func Hash(data []byte) string { h := sha256.Sum256(data); return hex.EncodeToString(h[:]) }

// Read reads only bounded regular files beneath dir. Missing files have a
// distinct revision from existing empty files. os.Root confines symlink targets.
func Read(dir, name string) ([]byte, string, error) {
	root, err := os.OpenRoot(dir)
	if err != nil {
		return nil, "", err
	}
	defer func() { _ = root.Close() }()
	info, err := root.Lstat(name)
	if os.IsNotExist(err) {
		return nil, "missing", nil
	}
	if err != nil {
		return nil, "", err
	}
	if !info.Mode().IsRegular() || info.Size() > MaxBytes {
		return nil, "", fmt.Errorf("not a bounded regular file: %s", name)
	}
	f, err := root.Open(name)
	if err != nil {
		return nil, "", err
	}
	defer func() { _ = f.Close() }()
	data, err := io.ReadAll(io.LimitReader(f, MaxBytes+1))
	if err != nil {
		return nil, "", err
	}
	if len(data) > MaxBytes {
		return nil, "", fmt.Errorf("file exceeds limit: %s", name)
	}
	return data, Hash(data), nil
}

func requestHash(raw []byte) string {
	var b bytes.Buffer
	if err := json.Compact(&b, raw); err != nil {
		return ""
	}
	return Hash(b.Bytes())
}

func validProjection(name string) bool {
	return name == "index.md" || name == "tasks.md" || name == "changelog.md"
}

// WithLock serializes cooperating operations, automatically releasing on process
// exit. It does not create any files, including for dry-runs and resume queries.
func WithLock(ctx context.Context, dir string, fn func() error) error {
	if ctx == nil {
		return fmt.Errorf("nil context")
	}
	release, err := lock(ctx, dir)
	if err != nil {
		return err
	}
	defer release()
	if err := ctx.Err(); err != nil {
		return err
	}
	return fn()
}

func Records(dir string) ([]Record, error) {
	root, err := os.OpenRoot(dir)
	if err != nil {
		return nil, err
	}
	defer func() { _ = root.Close() }()
	info, err := root.Lstat(directory)
	if os.IsNotExist(err) {
		return []Record{}, nil
	}
	if err != nil {
		return nil, err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return nil, fmt.Errorf("invalid operation directory")
	}
	f, err := root.Open(directory)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	entries, err := f.ReadDir(1025)
	if err != nil && err != io.EOF {
		return nil, err
	}
	if len(entries) > 1024 {
		return nil, fmt.Errorf("operation record limit exceeded (1024)")
	}
	out := []Record{}
	totalBytes := 0
	for _, entry := range entries {
		if filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		data, _, err := Read(dir, filepath.Join(directory, entry.Name()))
		if err != nil {
			return nil, err
		}
		totalBytes += len(data)
		if totalBytes > 16*MaxBytes {
			return nil, fmt.Errorf("operation history exceeds 64 MiB")
		}
		var rec Record
		if err := json.Unmarshal(data, &rec); err != nil {
			return nil, err
		}
		if err := validateRecord(rec); err != nil {
			return nil, err
		}
		if entry.Name() != rec.Receipt.OperationID+".json" {
			return nil, fmt.Errorf("operation filename identity mismatch")
		}
		out = append(out, rec)
	}
	sort.Slice(out, func(i, j int) bool {
		a, b := out[i].Receipt, out[j].Receipt
		if a.CreatedAt.Equal(b.CreatedAt) {
			return a.OperationID < b.OperationID
		}
		return a.CreatedAt.Before(b.CreatedAt)
	})
	return out, nil
}
func validateRecord(rec Record) error {
	if rec.Version != 1 || !idPattern.MatchString(rec.Receipt.OperationID) || requestHash(rec.Request) != rec.Digest {
		return fmt.Errorf("invalid operation record")
	}
	if rec.Receipt.State != "prepared" && rec.Receipt.State != "committed" {
		return fmt.Errorf("invalid operation state")
	}
	seen := map[string]bool{}
	total := 0
	for _, c := range rec.Changes {
		total += len(c.After)
		if !validProjection(c.Path) || seen[c.Path] || rec.Receipt.Before[c.Path] != c.Before || rec.Receipt.After[c.Path] != Hash(c.After) {
			return fmt.Errorf("invalid projection: %s", c.Path)
		}
		seen[c.Path] = true
	}
	if total > MaxBytes || len(rec.Changes) > 3 {
		return fmt.Errorf("projection bounds exceeded")
	}
	return nil
}

// Execute computes a plan under the ticket lock. The journal is prepared before
// any projection is written. Replays use the original plan and receipt, not a
// newly generated timestamp. Pending operations block unrelated new operations.
func (s Store) Execute(ctx context.Context, dir, id string, request any, dry bool, plan func() ([]Change, error)) (Receipt, error) {
	var result Receipt
	err := WithLock(ctx, dir, func() error {
		if !idPattern.MatchString(id) {
			return fmt.Errorf("invalid operation ID")
		}
		payload, err := json.Marshal(request)
		if err != nil {
			return err
		}
		if len(payload) > 65536 {
			return fmt.Errorf("request exceeds 64 KiB")
		}
		records, err := Records(dir)
		if err != nil {
			return err
		}
		var found *Record
		for i := range records {
			r := &records[i]
			if r.Receipt.OperationID == id {
				found = r
			}
		}
		if found != nil {
			if found.Digest != Hash(payload) {
				return fmt.Errorf("operation ID already used for a different request")
			}
			result = found.Receipt
			if result.State == "committed" {
				return nil
			}
			if dry {
				result.State = "dry-run"
				return s.check(dir, *found, true)
			}
			result, err = s.apply(ctx, dir, *found)
			return err
		}
		if len(records) >= 1024 {
			return fmt.Errorf("operation record limit reached (1024)")
		}
		for _, r := range records {
			if r.Receipt.State != "committed" {
				return fmt.Errorf("pending operation %s must be recovered first", r.Receipt.OperationID)
			}
		}
		changes, err := plan()
		if err != nil {
			return err
		}
		now := time.Now().UTC()
		if s.Now != nil {
			now = s.Now().UTC()
		}
		rec := Record{Version: 1, Digest: Hash(payload), Request: payload, Changes: changes, Receipt: Receipt{OperationID: id, State: "prepared", CreatedAt: now, ChangedPaths: []string{}, Before: map[string]string{}, After: map[string]string{}}}
		for _, c := range changes {
			rec.Receipt.ChangedPaths = append(rec.Receipt.ChangedPaths, c.Path)
			rec.Receipt.Before[c.Path] = c.Before
			rec.Receipt.After[c.Path] = Hash(c.After)
		}
		if err := validateRecord(rec); err != nil {
			return err
		}
		if err := s.check(dir, rec, false); err != nil {
			return err
		}
		result = rec.Receipt
		if dry {
			result.State = "dry-run"
			return nil
		}
		if err := s.save(dir, rec); err != nil {
			return err
		}
		result, err = s.apply(ctx, dir, rec)
		return err
	})
	return result, err
}
func (s Store) check(dir string, rec Record, recovering bool) error {
	for _, c := range rec.Changes {
		_, current, err := Read(dir, c.Path)
		if err != nil {
			return err
		}
		if current != c.Before && (!recovering || current != Hash(c.After)) {
			return fmt.Errorf("revision conflict: %s", c.Path)
		}
	}
	return nil
}
func (s Store) apply(ctx context.Context, dir string, rec Record) (Receipt, error) {
	fail := func(err error) (Receipt, error) {
		// Include projections completed by an earlier interrupted attempt too.
		visible := []string{}
		for _, c := range rec.Changes {
			_, hash, readErr := Read(dir, c.Path)
			if readErr == nil && hash == Hash(c.After) {
				visible = append(visible, c.Path)
			}
		}
		return rec.Receipt, &ApplyError{Receipt: rec.Receipt, Applied: visible, Cause: err}
	}
	if err := s.check(dir, rec, true); err != nil {
		return fail(err)
	}
	for _, c := range rec.Changes {
		if err := ctx.Err(); err != nil {
			return fail(err)
		}
		_, current, err := Read(dir, c.Path)
		if err != nil {
			return fail(err)
		}
		if current == Hash(c.After) {
			continue
		}
		if current != c.Before {
			return fail(fmt.Errorf("revision conflict: %s", c.Path))
		}
		if s.BeforeWrite != nil {
			if err := s.BeforeWrite(c.Path); err != nil {
				return fail(err)
			}
		}
		if err := s.write(dir, c.Path, c.After); err != nil {
			return fail(err)
		}
	}
	rec.Receipt.State = "committed"
	if err := s.save(dir, rec); err != nil {
		rec.Receipt.State = "prepared"
		return fail(err)
	}
	return rec.Receipt, nil
}
func (s Store) save(dir string, rec Record) error {
	data, err := json.MarshalIndent(rec, "", "  ")
	if err != nil {
		return err
	}
	if len(data)+1 > MaxBytes {
		return fmt.Errorf("journal exceeds 4 MiB")
	}
	if s.BeforeWrite != nil {
		if err := s.BeforeWrite("journal:" + rec.Receipt.State); err != nil {
			return err
		}
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		return err
	}
	defer func() { _ = root.Close() }()
	if err := root.Mkdir(directory, 0755); err != nil && !os.IsExist(err) {
		return err
	}
	return s.write(dir, filepath.Join(directory, rec.Receipt.OperationID+".json"), append(data, '\n'))
}
func (s Store) write(dir, name string, data []byte) error {
	// Resolve parents under os.Root and reject symlink components before using the
	// existing atomic writer. Uncooperative parent replacement is outside the lock contract.
	root, err := os.OpenRoot(dir)
	if err != nil {
		return err
	}
	defer func() { _ = root.Close() }()
	parent := filepath.Dir(name)
	if parent != "." {
		info, err := root.Lstat(parent)
		if err != nil {
			return err
		}
		if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("unsafe projection parent")
		}
	}
	_, err = documents.WriteFileIfChanged(filepath.Join(dir, name), data)
	return err
}
