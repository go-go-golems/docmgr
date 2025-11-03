package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/docmgr/docmgr/pkg/models"
	"gopkg.in/yaml.v3"
)

// LoadVocabulary loads vocabulary from doc/vocabulary.yaml
// Searches for the file starting from current directory and walking up to repo root
func LoadVocabulary() (*models.Vocabulary, error) {
	// Start from current directory and walk up to find doc/vocabulary.yaml
	dir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}

	for {
		vocabPath := filepath.Join(dir, "doc", "vocabulary.yaml")
		if _, err := os.Stat(vocabPath); err == nil {
			return loadVocabularyFromFile(vocabPath)
		}

		// Check if we've reached repo root (has .git or go.mod)
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root
			break
		}
		dir = parent
	}

	// Not found, return empty vocabulary
	return &models.Vocabulary{
		Topics:   []models.VocabItem{},
		DocTypes: []models.VocabItem{},
		Intent:   []models.VocabItem{},
	}, nil
}

// LoadVocabularyFromPath loads vocabulary from a specific file path
func loadVocabularyFromFile(path string) (*models.Vocabulary, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read vocabulary file: %w", err)
	}

	var vocab models.Vocabulary
	if err := yaml.Unmarshal(data, &vocab); err != nil {
		return nil, fmt.Errorf("failed to parse vocabulary file: %w", err)
	}

	return &vocab, nil
}

// SaveVocabulary saves vocabulary to doc/vocabulary.yaml
// Creates doc/ directory if it doesn't exist
func SaveVocabulary(vocab *models.Vocabulary, rootDir string) error {
	docDir := filepath.Join(rootDir, "doc")
	if err := os.MkdirAll(docDir, 0755); err != nil {
		return fmt.Errorf("failed to create doc directory: %w", err)
	}

	vocabPath := filepath.Join(docDir, "vocabulary.yaml")
	data, err := yaml.Marshal(vocab)
	if err != nil {
		return fmt.Errorf("failed to marshal vocabulary: %w", err)
	}

	if err := os.WriteFile(vocabPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write vocabulary file: %w", err)
	}

	return nil
}

// FindVocabularyPath finds the path to doc/vocabulary.yaml starting from current directory
func FindVocabularyPath() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	for {
		vocabPath := filepath.Join(dir, "doc", "vocabulary.yaml")
		if _, err := os.Stat(vocabPath); err == nil {
			return vocabPath, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("vocabulary.yaml not found (searched for doc/vocabulary.yaml)")
}

