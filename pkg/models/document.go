// Package models defines the core data structures for docmgr.
//
// The Document type represents a managed markdown document with YAML frontmatter
// containing metadata like ticket ID, topics, owners, and related files. Documents
// are organized into ticket workspaces using a date-based directory structure.
//
// The Vocabulary type defines controlled vocabularies for document metadata fields,
// allowing teams to standardize topics, doc types, and intent values.
//
// Example document structure:
//
//	---
//	Title: API Design for User Service
//	Ticket: MEN-3475
//	DocType: design-doc
//	Topics: [api, architecture]
//	Owners: [alice, bob]
//	RelatedFiles:
//	  - Path: backend/api/user.go
//	    Note: Main API implementation
//	---
//
//	# API Design
//
//	...document content...
package models

import (
	"fmt"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Document represents a managed markdown document with YAML frontmatter metadata.
//
// Documents are stored as markdown files with YAML frontmatter containing structured
// metadata. The frontmatter is parsed and validated against the vocabulary to ensure
// consistent use of topics, doc types, and intent values.
//
// Example:
//
//	doc := &models.Document{
//		Title:   "API Design for User Service",
//		Ticket:  "MEN-3475",
//		DocType: "design-doc",
//		Topics:  []string{"api", "architecture"},
//		Owners:  []string{"alice"},
//		RelatedFiles: []RelatedFile{
//			{Path: "backend/api/user.go", Note: "Main API implementation"},
//		},
//	}
//
// When serialized to a file, this becomes:
//
//	---
//	Title: API Design for User Service
//	Ticket: MEN-3475
//	DocType: design-doc
//	Topics: [api, architecture]
//	Owners: [alice]
//	RelatedFiles:
//	  - Path: backend/api/user.go
//	    Note: Main API implementation
//	---
//	# API Design
//	...content...
type Document struct {
	Title           string       `yaml:"Title" json:"title"`
	Ticket          string       `yaml:"Ticket" json:"ticket"`
	Status          string       `yaml:"Status" json:"status"`
	Topics          []string     `yaml:"Topics" json:"topics"`
	DocType         string       `yaml:"DocType" json:"docType"`
	Intent          string       `yaml:"Intent" json:"intent"`
	Owners          []string     `yaml:"Owners" json:"owners"`
	RelatedFiles    RelatedFiles `yaml:"RelatedFiles" json:"relatedFiles"`
	ExternalSources []string     `yaml:"ExternalSources" json:"externalSources"`
	Summary         string       `yaml:"Summary" json:"summary"`
	LastUpdated     time.Time    `yaml:"LastUpdated" json:"lastUpdated"`
}

// Validate checks that the document has all required fields populated.
//
// Required fields are:
//   - Title: Document title (must be non-empty)
//   - Ticket: Ticket identifier (must be non-empty)
//   - DocType: Document type (must be non-empty)
//
// Returns an error if any required field is missing. The error message
// lists all missing fields.
//
// Example:
//
//	doc := &Document{
//		Title:   "API Design",
//		Ticket:  "MEN-3475",
//		DocType: "design-doc",
//	}
//	if err := doc.Validate(); err != nil {
//		return fmt.Errorf("invalid document: %w", err)
//	}
func (d *Document) Validate() error {
	var missing []string

	if strings.TrimSpace(d.Title) == "" {
		missing = append(missing, "Title")
	}
	if strings.TrimSpace(d.Ticket) == "" {
		missing = append(missing, "Ticket")
	}
	if strings.TrimSpace(d.DocType) == "" {
		missing = append(missing, "DocType")
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required fields: %s", strings.Join(missing, ", "))
	}

	return nil
}

// Vocabulary defines the controlled vocabulary for document metadata fields.
//
// Vocabularies allow teams to standardize the values used for topics, document types,
// and intent fields across all documents. This ensures consistency and enables better
// search and filtering capabilities.
//
// Vocabulary files are typically stored as vocabulary.yaml in the documentation root
// and can be managed via `docmgr vocab` commands.
//
// Example vocabulary.yaml:
//
//	topics:
//	  - slug: api
//	    description: API design and implementation
//	  - slug: architecture
//	    description: System architecture decisions
//	docTypes:
//	  - slug: design-doc
//	    description: Design documents describing system architecture
//	  - slug: playbook
//	    description: Step-by-step procedures and runbooks
//	intent:
//	  - slug: long-term
//	    description: Long-term documentation that should be maintained
//	  - slug: short-term
//	    description: Temporary documentation for active work
type Vocabulary struct {
	Topics   []VocabItem `yaml:"topics" json:"topics"`
	DocTypes []VocabItem `yaml:"docTypes" json:"docTypes"`
	Intent   []VocabItem `yaml:"intent" json:"intent"`
}

// VocabItem represents a vocabulary entry
type VocabItem struct {
	Slug        string `yaml:"slug" json:"slug"`
	Description string `yaml:"description" json:"description"`
}

// ExternalSource represents metadata about an imported source
type ExternalSource struct {
	Type        string    `yaml:"type" json:"type"`
	Path        string    `yaml:"path" json:"path"`
	Repo        string    `yaml:"repo,omitempty" json:"repo,omitempty"`
	LastFetched time.Time `yaml:"lastFetched" json:"lastFetched"`
	SHA         string    `yaml:"sha,omitempty" json:"sha,omitempty"`
	URL         string    `yaml:"url,omitempty" json:"url,omitempty"`
}

// TicketWorkspace represents a ticket's documentation workspace.
//
// A TicketWorkspace contains metadata about a ticket's documentation directory,
// including the ticket ID, filesystem path, and the index document. This type is
// used to represent ticket workspaces when querying or manipulating documentation.
//
// Example:
//
//	workspace := TicketWorkspace{
//		Ticket:   "MEN-3475",
//		Path:     "ttmp/2025/11/18/MEN-3475-add-feature",
//		Document: &indexDoc,
//	}
type TicketWorkspace struct {
	Ticket   string
	Path     string
	Document *Document
}

// TicketDirectory is a deprecated alias for TicketWorkspace.
//
// Deprecated: Use TicketWorkspace instead. This alias is maintained for backward
// compatibility and will be removed in a future version.
type TicketDirectory = TicketWorkspace

// RelatedFile represents a single code file related to a document, with an optional
// note explaining why it's relevant.
//
// RelatedFiles link documentation to the actual code that implements or relates to
// the documented concepts. This enables traceability from documentation to code and
// helps maintain documentation as code evolves.
//
// Example:
//
//	rf := RelatedFile{
//		Path: "backend/api/user.go",
//		Note: "Main API implementation for user endpoints",
//	}
//
// In YAML frontmatter, RelatedFiles can be written in two formats:
//
//	# Legacy format (scalar strings, still supported for backward compatibility):
//	RelatedFiles:
//	  - backend/api/user.go
//	  - frontend/components/User.tsx
//
//	# Current format (structured with notes):
//	RelatedFiles:
//	  - Path: backend/api/user.go
//	    Note: Main API implementation
//	  - Path: frontend/components/User.tsx
//	    Note: Frontend component consuming the API
//
// UnmarshalYAML handles both formats automatically for backward compatibility.
type RelatedFile struct {
	Path string `yaml:"Path" json:"path"`
	Note string `yaml:"Note,omitempty" json:"note,omitempty"`
}

// UnmarshalYAML supports both scalar strings (legacy format) and mapping nodes
// (current format with Path/Note fields). This ensures backward compatibility with
// existing documents that use the simpler scalar string format.
//
// Supported formats:
//   - Scalar string: "path/to/file.go" → RelatedFile{Path: "path/to/file.go"}
//   - Mapping node: {Path: "path/to/file.go", Note: "description"} → RelatedFile{Path: "...", Note: "..."}
//
// The mapping node format supports case-insensitive keys (Path/path, Note/note).
func (rf *RelatedFile) UnmarshalYAML(value *yaml.Node) error {
	if value == nil {
		*rf = RelatedFile{}
		return nil
	}
	switch value.Kind {
	case yaml.ScalarNode:
		rf.Path = strings.TrimSpace(value.Value)
		return nil
	case yaml.MappingNode:
		// Manually decode to support both Path/Note and path/note keys (case-insensitive)
		var path string
		var note string
		// value.Content contains [key, value, key, value, ...]
		for i := 0; i+1 < len(value.Content); i += 2 {
			k := strings.ToLower(strings.TrimSpace(value.Content[i].Value))
			v := strings.TrimSpace(value.Content[i+1].Value)
			switch k {
			case "path":
				path = v
			case "note":
				note = v
			}
		}
		rf.Path = path
		rf.Note = note
		return nil
	case yaml.DocumentNode:
		// unsupported for this type; treat as empty
		*rf = RelatedFile{}
		return nil
	case yaml.SequenceNode:
		// unsupported for this type; treat as empty
		*rf = RelatedFile{}
		return nil
	case yaml.AliasNode:
		// unsupported for this type; treat as empty
		*rf = RelatedFile{}
		return nil
	default:
		// Treat unknown kinds as empty
		*rf = RelatedFile{}
		return nil
	}
}

// RelatedFiles is a list of RelatedFile entries that supports backward-compatible
// YAML decoding from either a sequence of scalar strings (legacy format) or a sequence
// of mapping nodes with Path/Note fields (current format).
//
// Example YAML formats:
//
//	# Legacy format (still supported):
//	RelatedFiles:
//	  - backend/api/user.go
//	  - frontend/components/User.tsx
//
//	# Current format:
//	RelatedFiles:
//	  - Path: backend/api/user.go
//	    Note: Main API implementation
//	  - Path: frontend/components/User.tsx
//	    Note: Frontend component
//
// When marshaling, RelatedFiles always outputs the structured format (mapping nodes)
// for consistency, regardless of the input format.
type RelatedFiles []RelatedFile

func (rfs *RelatedFiles) UnmarshalYAML(value *yaml.Node) error {
	if value == nil {
		*rfs = nil
		return nil
	}
	if value.Kind != yaml.SequenceNode {
		// Treat non-sequence as empty
		*rfs = nil
		return nil
	}
	out := make([]RelatedFile, 0, len(value.Content))
	for _, n := range value.Content {
		switch n.Kind {
		case yaml.ScalarNode:
			if n.Value != "" {
				out = append(out, RelatedFile{Path: n.Value})
			}
		case yaml.MappingNode:
			var rf RelatedFile
			if err := n.Decode(&rf); err != nil {
				// Best-effort: skip invalid entries
				continue
			}
			if rf.Path != "" {
				out = append(out, rf)
			}
		case yaml.DocumentNode:
			// ignore
		case yaml.SequenceNode:
			// ignore nested sequences
		case yaml.AliasNode:
			// ignore aliases
		default:
			// ignore other node kinds
		}
	}
	*rfs = out
	return nil
}

func (rfs RelatedFiles) MarshalYAML() (interface{}, error) {
	// Always marshal as a sequence of objects for consistency
	return []RelatedFile(rfs), nil
}
