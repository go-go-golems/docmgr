package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-go-golems/docmgr/pkg/models"
	"gopkg.in/yaml.v3"
)

// LoadVocabulary loads vocabulary from the configured path or defaults.
// Resolution order:
// - .ttmp.yaml 'vocabulary' path (relative to config dir if not absolute)
// - <root>/vocabulary.yaml, where root is from .ttmp.yaml (default 'ttmp')
// - fallback search for 'ttmp/vocabulary.yaml' upwards
// - legacy fallback 'doc/vocabulary.yaml' upwards
func LoadVocabulary() (*models.Vocabulary, error) {
	if path, err := ResolveVocabularyPath(); err == nil {
		if _, err2 := os.Stat(path); err2 == nil {
			return loadVocabularyFromFile(path)
		}
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

// SaveVocabulary saves vocabulary to the resolved vocabulary path, creating directories as needed.
// If no configuration is found, it defaults to '<repoRoot>/ttmp/vocabulary.yaml'.
func SaveVocabulary(vocab *models.Vocabulary, repoRoot string) error {
	// Resolve configured path or default to <repoRoot>/ttmp/vocabulary.yaml
	vocabPath, err := ResolveVocabularyPath()
	if err != nil || vocabPath == "" {
		vocabPath = filepath.Join(repoRoot, "ttmp", "vocabulary.yaml")
	}
	dir := filepath.Dir(vocabPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create vocabulary directory: %w", err)
	}

	data, err := yaml.Marshal(vocab)
	if err != nil {
		return fmt.Errorf("failed to marshal vocabulary: %w", err)
	}
	if err := os.WriteFile(vocabPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write vocabulary file: %w", err)
	}

	return nil
}
