package common

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestDisableFlagCommaSplittingKeepsCommaValuesIntact(t *testing.T) {
	cmd := &cobra.Command{Use: "test", Run: func(cmd *cobra.Command, args []string) {}}
	cmd.Flags().StringSlice("file-note", []string{}, "repeatable path:note")

	if err := DisableFlagCommaSplitting(cmd, "file-note"); err != nil {
		t.Fatalf("DisableFlagCommaSplitting: %v", err)
	}

	cmd.SetArgs([]string{
		"--file-note", "pkg/a.go:note (sections 4.4, 8.1)",
		"--file-note", "pkg/b.go:plain note",
	})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	got, err := cmd.Flags().GetStringSlice("file-note")
	if err != nil {
		t.Fatalf("GetStringSlice: %v", err)
	}
	want := []string{"pkg/a.go:note (sections 4.4, 8.1)", "pkg/b.go:plain note"}
	if len(got) != len(want) {
		t.Fatalf("GetStringSlice() = %#v, want %#v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("GetStringSlice()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestDisableFlagCommaSplittingUnknownFlag(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	if err := DisableFlagCommaSplitting(cmd, "missing"); err == nil {
		t.Fatal("expected error for unknown flag")
	}
}
