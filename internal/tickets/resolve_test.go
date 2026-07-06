package tickets

import (
	"errors"
	"strings"
	"testing"
)

func TestMatchTicketRef(t *testing.T) {
	candidates := []Candidate{
		{ID: "MEN-4242", DirBase: "MEN-4242--normalize-chat-api-paths"},
		{ID: "MEN-5678", DirBase: "MEN-5678--secondary-ticket"},
		{ID: "DOCMGR-200", DirBase: "DOCMGR-200--improve-docmgr"},
	}

	cases := []struct {
		name    string
		ref     string
		want    string
		wantErr error
	}{
		{name: "exact ID", ref: "MEN-4242", want: "MEN-4242"},
		{name: "case-insensitive exact", ref: "men-4242", want: "MEN-4242"},
		{name: "unique prefix", ref: "DOCMGR", want: "DOCMGR-200"},
		{name: "ambiguous prefix", ref: "MEN-", wantErr: ErrAmbiguous},
		{name: "directory basename", ref: "MEN-4242--normalize-chat-api-paths", want: "MEN-4242"},
		{name: "directory path", ref: "2026/07/05/MEN-5678--secondary-ticket/", want: "MEN-5678"},
		{name: "slug with unknown suffix", ref: "DOCMGR-200--something-else", want: "DOCMGR-200"},
		{name: "unique substring", ref: "5678", want: "MEN-5678"},
		{name: "not found", ref: "NOPE-1", wantErr: ErrNotFound},
		{name: "empty", ref: "  ", wantErr: nil, want: ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := MatchTicketRef(tc.ref, candidates)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("MatchTicketRef(%q) error = %v, want %v", tc.ref, err, tc.wantErr)
				}
				return
			}
			if tc.want == "" {
				if err == nil {
					t.Fatalf("MatchTicketRef(%q) expected an error, got %q", tc.ref, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("MatchTicketRef(%q) unexpected error: %v", tc.ref, err)
			}
			if got != tc.want {
				t.Fatalf("MatchTicketRef(%q) = %q, want %q", tc.ref, got, tc.want)
			}
		})
	}
}

func TestMatchTicketRefAmbiguityListsCandidates(t *testing.T) {
	candidates := []Candidate{
		{ID: "TEST-100", DirBase: "TEST-100--a"},
		{ID: "TEST-200", DirBase: "TEST-200--b"},
	}
	_, err := MatchTicketRef("TEST", candidates)
	if !errors.Is(err, ErrAmbiguous) {
		t.Fatalf("expected ErrAmbiguous, got %v", err)
	}
	if !strings.Contains(err.Error(), "TEST-100") || !strings.Contains(err.Error(), "TEST-200") {
		t.Fatalf("ambiguity error should list candidates, got: %v", err)
	}
}
