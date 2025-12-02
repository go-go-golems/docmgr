package commands

import (
	"bytes"
	"testing"
)

func TestNormalizeDelimitersAddsClosingAndBody(t *testing.T) {
	raw := []byte(`---
Title: T
Ticket: TEST-1
DocType: reference
Summary: needs quotes: here
Body line should be treated as body`)

	fixed, err := normalizeDelimiters(raw)
	if err != nil {
		t.Fatalf("normalizeDelimiters error: %v", err)
	}
	if !bytes.HasPrefix(fixed, []byte("---\nTitle: T\n")) {
		t.Fatalf("unexpected prefix:\n%s", string(fixed))
	}
	if !bytes.Contains(fixed, []byte("\n---\n")) || !bytes.Contains(fixed, []byte("Body line")) {
		t.Fatalf("expected closing delimiter and body separation, got:\n%s", string(fixed))
	}
}

func TestGenerateFixesHandlesStrayDelimiterAndTrailingLines(t *testing.T) {
	raw := []byte(`---
Title: T
Ticket: TEST-2
DocType: reference
Summary: broken delimiters
----
Body is here`)

	fixes, fixed, err := generateFixes(raw)
	if err != nil {
		t.Fatalf("generateFixes error: %v", err)
	}
	if len(fixes) == 0 {
		t.Fatalf("expected fixes")
	}
	if bytes.Contains(fixed, []byte("----")) {
		t.Fatalf("expected stray delimiter removed, got:\n%s", string(fixed))
	}
	if !bytes.Contains(fixed, []byte("Body is here")) {
		t.Fatalf("expected body retained, got:\n%s", string(fixed))
	}
}
