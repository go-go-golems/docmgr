package docmgr

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/go-go-golems/docmgr/pkg/diagnostics/core"
	"github.com/go-go-golems/docmgr/pkg/diagnostics/docmgrrules"
	"github.com/go-go-golems/docmgr/pkg/diagnostics/render"
)

// RenderTaxonomy renders diagnostics using the default docmgr rules/renderer.
// Best effort: failures are ignored to avoid masking the original command error.
func RenderTaxonomy(ctx context.Context, tax *core.Taxonomy) {
	if tax == nil {
		return
	}
	reg := docmgrrules.DefaultRegistry()
	results, err := reg.RenderAll(ctx, tax)
	if err != nil {
		return
	}
	out := render.RenderToText(results)
	if strings.TrimSpace(out) != "" {
		fmt.Fprintln(os.Stderr, out)
	}
}
