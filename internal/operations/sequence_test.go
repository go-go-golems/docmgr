package operations

import (
	"context"
	"testing"
	"time"
)

func TestSequenceOrdersCheckpointsAcrossClockRollback(t *testing.T) {
	dir := t.TempDir()
	plan := func() ([]Change, error) { return []Change{}, nil }
	first, err := (Store{Now: func() time.Time { return time.Unix(200, 0) }}).Execute(context.Background(), dir, "z-first", true, false, plan)
	checkTest(t, err)
	last, err := (Store{Now: func() time.Time { return time.Unix(100, 0) }}).Execute(context.Background(), dir, "a-last", true, false, plan)
	checkTest(t, err)
	if last.Sequence <= first.Sequence {
		t.Fatal("sequence did not advance")
	}
	records, err := Records(dir)
	checkTest(t, err)
	if records[len(records)-1].Receipt.OperationID != "a-last" {
		t.Fatal("clock rollback lost latest record")
	}
}
