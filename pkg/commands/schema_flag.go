package commands

import (
	"os"
	"strings"

	"github.com/go-go-golems/glazed/pkg/cmds/layers"
)

// isSchemaFlagSet checks the raw parsed layers for the presence of the
// --print-template-schema flag, accommodating both bool and string values.
func isSchemaFlagSet(pl *layers.ParsedLayers) bool {
	if pl == nil {
		return false
	}
	if pp, ok := pl.GetParameter(layers.DefaultSlug, "print-template-schema"); ok && pp != nil {
		switch v := pp.Value.(type) {
		case bool:
			return v
		case string:
			s := strings.ToLower(strings.TrimSpace(v))
			return s == "true" || s == "1" || s == "yes" || s == "y"
		}
	}
	return false
}

// isSchemaFlagInArgs performs a last-resort check for the raw flag in os.Args.
func isSchemaFlagInArgs() bool {
	for i, a := range os.Args {
		if a == "--print-template-schema" || strings.HasPrefix(a, "--print-template-schema=") {
			return true
		}
		// Handle form: --print-template-schema true
		if a == "--print-template-schema" && i+1 < len(os.Args) {
			n := strings.ToLower(strings.TrimSpace(os.Args[i+1]))
			if n == "true" || n == "1" || n == "yes" || n == "y" {
				return true
			}
		}
	}
	return false
}


