package docmgr

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/go-go-golems/docmgr/pkg/diagnostics/core"
	"github.com/go-go-golems/docmgr/pkg/diagnostics/docmgrrules"
	"github.com/go-go-golems/docmgr/pkg/diagnostics/render"
	"github.com/go-go-golems/docmgr/pkg/diagnostics/rules"
)

type rendererKey struct{}

// Renderer renders diagnostics and can optionally collect results for JSON output.
type Renderer struct {
	registry   *rules.Registry
	writer     io.Writer
	renderText bool
	collect    bool
	results    []*rules.RuleResult
}

type Option func(*Renderer)

var defaultRenderer = NewRenderer()

// NewRenderer builds a renderer configured with docmgr rules and output options.
func NewRenderer(opts ...Option) *Renderer {
	r := &Renderer{
		registry:   docmgrrules.DefaultRegistry(),
		writer:     os.Stderr,
		renderText: true,
		collect:    false,
		results:    []*rules.RuleResult{},
	}
	for _, opt := range opts {
		if opt != nil {
			opt(r)
		}
	}
	return r
}

// WithWriter overrides the text output destination (defaults to stderr).
func WithWriter(w io.Writer) Option {
	return func(r *Renderer) {
		if w != nil {
			r.writer = w
		}
	}
}

// WithCollector enables result collection for JSON output.
func WithCollector() Option {
	return func(r *Renderer) {
		r.collect = true
	}
}

// WithoutText disables text rendering (useful for tests/CI-only JSON).
func WithoutText() Option {
	return func(r *Renderer) {
		r.renderText = false
	}
}

// WithRegistry supplies a custom registry (defaults to docmgrrules.DefaultRegistry()).
func WithRegistry(reg *rules.Registry) Option {
	return func(r *Renderer) {
		if reg != nil {
			r.registry = reg
		}
	}
}

// ContextWithRenderer attaches a renderer to the context for downstream RenderTaxonomy calls.
func ContextWithRenderer(ctx context.Context, r *Renderer) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, rendererKey{}, r)
}

func rendererFromContext(ctx context.Context) *Renderer {
	if ctx == nil {
		return nil
	}
	if r, ok := ctx.Value(rendererKey{}).(*Renderer); ok {
		return r
	}
	return nil
}

// Render renders the taxonomy, collecting results and printing text based on renderer options.
func (r *Renderer) Render(ctx context.Context, tax *core.Taxonomy) {
	if tax == nil || r == nil {
		return
	}
	reg := r.registry
	if reg == nil {
		reg = docmgrrules.DefaultRegistry()
	}
	results, err := reg.RenderAll(ctx, tax)
	if err != nil {
		return
	}
	if r.collect {
		r.results = append(r.results, results...)
	}
	if r.renderText {
		out := render.RenderToText(results)
		if strings.TrimSpace(out) != "" {
			fmt.Fprintln(r.writer, out)
		}
	}
}

// Results returns collected rule results (empty slice if none).
func (r *Renderer) Results() []*rules.RuleResult {
	return r.results
}

// JSON renders collected results into pretty JSON.
func (r *Renderer) JSON() ([]byte, error) {
	return render.RenderToJSON(r.results)
}

// RenderTaxonomy renders diagnostics using a renderer attached to the context, falling back to
// the default renderer when no renderer is present.
func RenderTaxonomy(ctx context.Context, tax *core.Taxonomy) {
	if tax == nil {
		return
	}
	if r := rendererFromContext(ctx); r != nil {
		r.Render(ctx, tax)
		return
	}
	defaultRenderer.Render(ctx, tax)
}
