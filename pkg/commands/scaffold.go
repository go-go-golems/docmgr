package commands

import (
    "fmt"
    "os"
    "path/filepath"
)

// writeFileIfNotExists writes content to a file only if it doesn't exist,
// unless force is true. Returns an error if file exists and force is false.
func writeFileIfNotExists(path string, content []byte, force bool) error {
    if !force {
        if _, err := os.Stat(path); err == nil {
            // File exists, skip writing
            return nil
        }
    }
    return os.WriteFile(path, content, 0644)
}

// scaffoldTemplatesAndGuidelines creates the _templates/ and _guidelines/ directories
// at the root level and populates them with template and guideline files
func scaffoldTemplatesAndGuidelines(root string, force bool) error {
    templatesDir := filepath.Join(root, "_templates")
    guidelinesDir := filepath.Join(root, "_guidelines")

    // Create directories
    if err := os.MkdirAll(templatesDir, 0755); err != nil {
        return fmt.Errorf("failed to create templates directory: %w", err)
    }
    if err := os.MkdirAll(guidelinesDir, 0755); err != nil {
        return fmt.Errorf("failed to create guidelines directory: %w", err)
    }

    // Write template files
    for docType, template := range TemplateContent {
        templatePath := filepath.Join(templatesDir, fmt.Sprintf("%s.md", docType))
        if err := writeFileIfNotExists(templatePath, []byte(template), force); err != nil {
            return fmt.Errorf("failed to write template %s: %w", docType, err)
        }
    }

    // Write guideline files
    for docType, guideline := range GuidelineContent {
        guidelinePath := filepath.Join(guidelinesDir, fmt.Sprintf("%s.md", docType))
        if err := writeFileIfNotExists(guidelinePath, []byte(guideline), force); err != nil {
            return fmt.Errorf("failed to write guideline %s: %w", docType, err)
        }
    }

    return nil
}


