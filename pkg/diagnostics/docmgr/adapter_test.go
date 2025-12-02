package docmgr

import (
	"context"
	"strings"
	"testing"

	"github.com/go-go-golems/docmgr/pkg/diagnostics/docmgrctx"
)

func TestRendererCollectsAndOutputsJSON(t *testing.T) {
	renderer := NewRenderer(WithCollector(), WithoutText())
	ctx := ContextWithRenderer(context.Background(), renderer)

	tax := docmgrctx.NewVocabularyUnknownTaxonomy("doc.md", "Topics", "custom", []string{"chat"})
	RenderTaxonomy(ctx, tax)

	results := renderer.Results()
	if len(results) == 0 {
		t.Fatalf("expected collected results, got none")
	}

	data, err := renderer.JSON()
	if err != nil {
		t.Fatalf("failed to marshal diagnostics JSON: %v", err)
	}
	if !strings.Contains(string(data), "Topics") {
		t.Fatalf("expected JSON to mention field name, got: %s", string(data))
	}
}
