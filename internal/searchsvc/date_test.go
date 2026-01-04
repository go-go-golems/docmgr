package searchsvc

import (
	"testing"
	"time"
)

func TestParseDate_RelativeDatesNormalizeToStartOfDay(t *testing.T) {
	t.Parallel()

	for _, input := range []string{
		"1 day ago",
		"2 weeks ago",
		"3 months ago",
		"4 years ago",
	} {
		got, err := ParseDate(input)
		if err != nil {
			t.Fatalf("ParseDate(%q) error: %v", input, err)
		}
		if got.IsZero() {
			t.Fatalf("ParseDate(%q) returned zero time", input)
		}
		if got.Hour() != 0 || got.Minute() != 0 || got.Second() != 0 || got.Nanosecond() != 0 {
			t.Fatalf("ParseDate(%q) expected start-of-day; got %s", input, got.Format(time.RFC3339Nano))
		}
	}
}

func TestParseDate_ThisWeekIsStartOfDay(t *testing.T) {
	t.Parallel()

	got, err := ParseDate("this week")
	if err != nil {
		t.Fatalf("ParseDate(this week) error: %v", err)
	}
	if got.Hour() != 0 || got.Minute() != 0 || got.Second() != 0 || got.Nanosecond() != 0 {
		t.Fatalf("expected start-of-day; got %s", got.Format(time.RFC3339Nano))
	}
}
