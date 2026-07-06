package common

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// repeatableStringSliceValue is a pflag value with cobra StringArray semantics:
// each flag occurrence contributes exactly one element and values are never
// split on commas. It still reports itself as a stringSlice and renders as CSV
// so glazed's GetStringSlice-based flag gathering keeps working unchanged.
type repeatableStringSliceValue struct {
	values  []string
	changed bool
}

func (s *repeatableStringSliceValue) Set(val string) error {
	if !s.changed {
		s.values = []string{val}
		s.changed = true
	} else {
		s.values = append(s.values, val)
	}
	return nil
}

func (s *repeatableStringSliceValue) Type() string { return "stringSlice" }

func (s *repeatableStringSliceValue) String() string {
	if len(s.values) == 0 {
		return "[]"
	}
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	_ = w.Write(s.values)
	w.Flush()
	return "[" + strings.TrimSuffix(buf.String(), "\n") + "]"
}

// DisableFlagCommaSplitting swaps the named stringList-backed flags registered
// by glazed for values that treat each repeated flag occurrence as a single
// element (no comma splitting), e.g. --file-note "a.go:note (sections 4.4, 8.1)".
func DisableFlagCommaSplitting(cmd *cobra.Command, names ...string) error {
	for _, name := range names {
		f := cmd.Flags().Lookup(name)
		if f == nil {
			return fmt.Errorf("flag %q not found on command %s", name, cmd.Name())
		}
		f.Value = &repeatableStringSliceValue{}
		f.DefValue = "[]"
	}
	return nil
}
