package render

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-go-golems/docmgr/pkg/diagnostics/rules"
)

// RenderToText formats rule results as human-readable text blocks.
func RenderToText(results []*rules.RuleResult) string {
	if len(results) == 0 {
		return ""
	}

	var b strings.Builder
	for i, res := range results {
		fmt.Fprintf(&b, "%d) [%s] %s\n", i+1, res.Severity, res.Headline)
		if res.Body != "" {
			b.WriteString(res.Body)
			if !strings.HasSuffix(res.Body, "\n") {
				b.WriteString("\n")
			}
		}
		if len(res.Actions) > 0 {
			b.WriteString("Actions:\n")
			for _, a := range res.Actions {
				line := a.Label
				if a.Command != "" {
					args := strings.Join(a.Args, " ")
					if args != "" {
						line = fmt.Sprintf("- %s: %s %s", a.Label, a.Command, args)
					} else {
						line = fmt.Sprintf("- %s: %s", a.Label, a.Command)
					}
				} else {
					line = "- " + line
				}
				b.WriteString(line + "\n")
			}
		}
		if i < len(results)-1 {
			b.WriteString("\n")
		}
	}
	return b.String()
}

// RenderToJSON marshals rule results to pretty JSON for CI/IDE consumption.
func RenderToJSON(results []*rules.RuleResult) ([]byte, error) {
	return json.MarshalIndent(results, "", "  ")
}
