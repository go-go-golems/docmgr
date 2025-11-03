package models

import (
    "strings"
    "time"

    "gopkg.in/yaml.v3"
)

// Document represents a managed document with metadata
type Document struct {
	Title           string    `yaml:"Title" json:"title"`
	Ticket          string    `yaml:"Ticket" json:"ticket"`
	Status          string    `yaml:"Status" json:"status"`
	Topics          []string  `yaml:"Topics" json:"topics"`
	DocType         string    `yaml:"DocType" json:"docType"`
	Intent          string    `yaml:"Intent" json:"intent"`
    Owners          []string      `yaml:"Owners" json:"owners"`
    RelatedFiles    RelatedFiles  `yaml:"RelatedFiles" json:"relatedFiles"`
	ExternalSources []string  `yaml:"ExternalSources" json:"externalSources"`
	Summary         string    `yaml:"Summary" json:"summary"`
	LastUpdated     time.Time `yaml:"LastUpdated" json:"lastUpdated"`
}

// Vocabulary defines the allowed values for various fields
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

// TicketDirectory represents a ticket's documentation workspace
type TicketDirectory struct {
	Ticket   string
	Path     string
	Document *Document
}

// RelatedFile represents a single related file with an optional rationale note.
type RelatedFile struct {
    Path string `yaml:"Path" json:"path"`
    Note string `yaml:"Note,omitempty" json:"note,omitempty"`
}

// UnmarshalYAML supports both scalar strings (legacy) and mapping nodes.
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
        type alias RelatedFile
        var tmp alias
        if err := value.Decode(&tmp); err != nil {
            return err
        }
        *rf = RelatedFile(tmp)
        return nil
    default:
        // Treat unknown kinds as empty
        *rf = RelatedFile{}
        return nil
    }
}

// RelatedFiles is a list that supports backward-compatible YAML decoding from either
// a sequence of scalars (paths) or a sequence of maps (with Path/Note fields).
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
