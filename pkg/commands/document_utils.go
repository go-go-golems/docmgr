package commands

import (
	"github.com/go-go-golems/docmgr/internal/documents"
	"github.com/go-go-golems/docmgr/pkg/models"
)

func readDocumentFrontmatter(path string) (*models.Document, error) {
	doc, _, err := documents.ReadDocumentWithFrontmatter(path)
	return doc, err
}

func readDocumentWithContent(path string) (*models.Document, string, error) {
	return documents.ReadDocumentWithFrontmatter(path)
}
