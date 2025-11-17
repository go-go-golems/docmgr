package commands

import "testing"

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
