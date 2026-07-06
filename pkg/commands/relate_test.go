package commands

import (
	"strings"
	"testing"
)

func TestParseFileNotes(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name    string
		input   []string
		want    map[string]string
		wantErr string
	}{
		{
			name:  "parses colon and equals delimiters",
			input: []string{"a/b.go:note one", "c/d.ts=note two"},
			want:  map[string]string{"a/b.go": "note one", "c/d.ts": "note two"},
		},
		{
			name:  "keeps commas inside a single note",
			input: []string{"pkg/commands/relate.go:note (sections 4.4, 8.1)"},
			want:  map[string]string{"pkg/commands/relate.go": "note (sections 4.4, 8.1)"},
		},
		{
			name:  "skips empty values",
			input: []string{"", "   ", "a/b.go:note"},
			want:  map[string]string{"a/b.go": "note"},
		},
		{
			name:    "errors on missing delimiter",
			input:   []string{"a/b.go note without delimiter"},
			wantErr: "malformed --file-note value \"a/b.go note without delimiter\"",
		},
		{
			name:    "errors on empty path",
			input:   []string{":note"},
			wantErr: "empty path",
		},
		{
			name:    "errors on empty note",
			input:   []string{"a/b.go:"},
			wantErr: "non-empty note for a/b.go",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := parseFileNotes(tc.input)
			if tc.wantErr != "" {
				if err == nil {
					t.Fatalf("parseFileNotes() = %v, expected error containing %q", got, tc.wantErr)
				}
				if !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("parseFileNotes() error = %q, expected it to contain %q", err.Error(), tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseFileNotes() unexpected error: %v", err)
			}
			if len(got) != len(tc.want) {
				t.Fatalf("parseFileNotes() = %v, want %v", got, tc.want)
			}
			for k, v := range tc.want {
				if got[k] != v {
					t.Fatalf("parseFileNotes()[%q] = %q, want %q", k, got[k], v)
				}
			}
		})
	}
}

func TestAppendNote(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name      string
		existing  string
		addition  string
		want      string
		didChange bool
	}{
		{
			name:      "adds note when existing empty",
			existing:  "",
			addition:  "added note",
			want:      "added note",
			didChange: true,
		},
		{
			name:      "trims addition and avoids duplicates",
			existing:  "first note",
			addition:  "  first note  ",
			want:      "first note",
			didChange: false,
		},
		{
			name:      "appends with newline",
			existing:  "first note",
			addition:  "second note",
			want:      "first note\nsecond note",
			didChange: true,
		},
		{
			name:      "preserves trailing newline",
			existing:  "first note\n",
			addition:  "second note",
			want:      "first note\nsecond note",
			didChange: true,
		},
		{
			name:      "skips empty addition",
			existing:  "first note",
			addition:  "   ",
			want:      "first note",
			didChange: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, changed := appendNote(tc.existing, tc.addition)
			if got != tc.want {
				t.Fatalf("appendNote() = %q, want %q", got, tc.want)
			}
			if changed != tc.didChange {
				t.Fatalf("appendNote() change = %v, want %v", changed, tc.didChange)
			}
		})
	}
}
