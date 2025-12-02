package utils

import "testing"

func TestSlugifyTitleForTicket(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		ticket string
		title  string
		slug   string
	}{
		{
			name:   "title begins with ticket and colon",
			ticket: "TEST-9999",
			title:  "TEST-9999: Test ticket with ticket in title",
			slug:   "test-ticket-with-ticket-in-title",
		},
		{
			name:   "title begins with ticket and dash",
			ticket: "TEST-1234",
			title:  "TEST-1234 - Another test",
			slug:   "another-test",
		},
		{
			name:   "title begins with ticket and unicode dash",
			ticket: "MEN-4242",
			title:  "MEN-4242 — Unicode dash title",
			slug:   "unicode-dash-title",
		},
		{
			name:   "title without ticket prefix",
			ticket: "MEN-5678",
			title:  "Secondary ticket — WebSocket reconnection plan",
			slug:   "secondary-ticket-websocket-reconnection-plan",
		},
		{
			name:   "title equals ticket",
			ticket: "MEN-0001",
			title:  "MEN-0001",
			slug:   "men-0001",
		},
		{
			name:   "empty title falls back to ticket",
			ticket: "DOC-42",
			title:  "",
			slug:   "doc-42",
		},
		{
			name:   "empty ticket falls back to title",
			ticket: "",
			title:  "Just a title",
			slug:   "just-a-title",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := SlugifyTitleForTicket(tc.ticket, tc.title)
			if got != tc.slug {
				t.Fatalf("expected slug %q, got %q", tc.slug, got)
			}
		})
	}
}
