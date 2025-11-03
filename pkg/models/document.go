package models

import (
	"time"
)

// Document represents a managed document with metadata
type Document struct {
	Title           string    `yaml:"Title" json:"title"`
	Ticket          string    `yaml:"Ticket" json:"ticket"`
	Status          string    `yaml:"Status" json:"status"`
	Topics          []string  `yaml:"Topics" json:"topics"`
	DocType         string    `yaml:"DocType" json:"docType"`
	Intent          string    `yaml:"Intent" json:"intent"`
	Owners          []string  `yaml:"Owners" json:"owners"`
	RelatedFiles    []string  `yaml:"RelatedFiles" json:"relatedFiles"`
	ExternalSources []string  `yaml:"ExternalSources" json:"externalSources"`
	Summary         string    `yaml:"Summary" json:"summary"`
	LastUpdated     time.Time `yaml:"LastUpdated" json:"lastUpdated"`
}

// Vocabulary defines the allowed values for various fields
type Vocabulary struct {
	Topics      []VocabItem `yaml:"topics" json:"topics"`
	DocTypes    []VocabItem `yaml:"docTypes" json:"docTypes"`
	SourceTypes []VocabItem `yaml:"sourceTypes" json:"sourceTypes"`
	Lifecycle   []VocabItem `yaml:"lifecycle" json:"lifecycle"`
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
