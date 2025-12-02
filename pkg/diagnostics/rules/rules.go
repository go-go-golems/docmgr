package rules

import (
	"context"
	"sort"

	"github.com/go-go-golems/docmgr/pkg/diagnostics/core"
)

// RuleResult is produced by a rule renderer.
type RuleResult struct {
	Headline string
	Body     string
	Severity core.Severity
	Actions  []Action
}

// Action suggests a follow-up command.
type Action struct {
	Label   string
	Command string
	Args    []string
}

// Rule matches a taxonomy and renders a result.
type Rule interface {
	Match(*core.Taxonomy) (bool, int) // ok, score (0-100)
	Render(context.Context, *core.Taxonomy) (*RuleResult, error)
}

// Registry stores rules and renders them by score.
type Registry struct {
	rules []Rule
}

func NewRegistry() *Registry {
	return &Registry{rules: []Rule{}}
}

// Register adds a rule to the registry.
func (r *Registry) Register(rule Rule) {
	if rule == nil {
		return
	}
	r.rules = append(r.rules, rule)
}

// RenderAll runs all matching rules, sorted by score (desc).
func (r *Registry) RenderAll(ctx context.Context, t *core.Taxonomy) ([]*RuleResult, error) {
	type scored struct {
		res   *RuleResult
		score int
	}
	matches := []scored{}

	for _, rule := range r.rules {
		ok, score := rule.Match(t)
		if !ok {
			continue
		}
		res, err := rule.Render(ctx, t)
		if err != nil {
			return nil, err
		}
		matches = append(matches, scored{res: res, score: score})
	}

	sort.SliceStable(matches, func(i, j int) bool {
		return matches[i].score > matches[j].score
	})

	out := make([]*RuleResult, 0, len(matches))
	for _, m := range matches {
		out = append(out, m.res)
	}
	return out, nil
}
