package core

import (
	"errors"
	"fmt"
)

// StageCode identifies the stage where a problem occurred (tool-defined).
type StageCode string

// SymptomCode identifies the symptom of a problem (tool-defined).
type SymptomCode string

// Severity indicates error importance (info|warning|error).
type Severity string

const (
	SeverityInfo    Severity = "info"
	SeverityWarning Severity = "warning"
	SeverityError   Severity = "error"
)

// ContextPayload carries tool-specific details for a taxonomy entry.
// Implementations should be small, typed structs per domain.
type ContextPayload interface {
	Stage() StageCode
	Summary() string
}

// Taxonomy models a structured diagnostic that can be rendered by rules.
type Taxonomy struct {
	Tool     string
	Stage    StageCode
	Symptom  SymptomCode
	Path     string
	Severity Severity
	Context  ContextPayload
	Cause    error
}

func (t *Taxonomy) ContextSummary() string {
	if t == nil || t.Context == nil {
		return ""
	}
	return t.Context.Summary()
}

// taxonomyError wraps a Taxonomy to participate in Go error chains.
type taxonomyError struct {
	taxonomy *Taxonomy
}

func (e *taxonomyError) Error() string {
	t := e.taxonomy
	if t == nil {
		return "taxonomy: <nil>"
	}
	stage := string(t.Stage)
	if stage == "" {
		stage = "unknown-stage"
	}
	symptom := string(t.Symptom)
	if symptom == "" {
		symptom = "unknown-symptom"
	}
	summary := t.ContextSummary()
	if summary == "" {
		return fmt.Sprintf("taxonomy: %s/%s", stage, symptom)
	}
	return fmt.Sprintf("taxonomy: %s/%s: %s", stage, symptom, summary)
}

func (e *taxonomyError) Unwrap() error {
	if e.taxonomy == nil {
		return nil
	}
	return e.taxonomy.Cause
}

// WrapTaxonomy creates an error carrying taxonomy data and optional cause.
func WrapTaxonomy(t *Taxonomy) error {
	return &taxonomyError{taxonomy: t}
}

// WrapWithCause attaches the cause to the taxonomy and returns a wrapped error.
func WrapWithCause(cause error, t *Taxonomy) error {
	if t != nil {
		t.Cause = cause
	}
	return &taxonomyError{taxonomy: t}
}

// AsTaxonomy extracts Taxonomy data from an error chain.
func AsTaxonomy(err error) (*Taxonomy, bool) {
	var te *taxonomyError
	if errors.As(err, &te) && te != nil {
		return te.taxonomy, true
	}
	return nil, false
}
