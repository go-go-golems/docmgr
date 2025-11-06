package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var numericPrefixRe = regexp.MustCompile(`^(\d{2,3})-`)

// hasNumericPrefix returns true if the filename begins with NN- or NNN-.
func hasNumericPrefix(name string) bool {
	return numericPrefixRe.MatchString(name)
}

// stripNumericPrefix removes a leading NN-/NNN- and returns the remainder.
// It also returns the parsed prefix number (or 0 if none) and the width used.
func stripNumericPrefix(name string) (string, int, int) {
	m := numericPrefixRe.FindStringSubmatch(name)
	if len(m) == 2 {
		width := len(m[1])
		num := 0
		if n, err := strconv.Atoi(m[1]); err == nil {
			num = n
		}
		return name[len(m[0]):], num, width
	}
	return name, 0, 0
}

// nextPrefixForDir scans a directory for .md files with numeric prefixes and
// returns the next prefix string (e.g., "01-" or "100-") along with the next
// integer and width selected (2 unless next >= 100, then 3).
func nextPrefixForDir(dir string) (string, int, int, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", 0, 0, err
	}
	maxNum := 0
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".md") {
			continue
		}
		if m := numericPrefixRe.FindStringSubmatch(name); len(m) == 2 {
			if n, err := strconv.Atoi(m[1]); err == nil {
				if n > maxNum {
					maxNum = n
				}
			}
		}
	}
	next := maxNum + 1
	width := 2
	if next >= 100 {
		width = 3
	}
	prefix := fmt.Sprintf("%0*d-", width, next)
	return prefix, next, width, nil
}

// buildPrefixedDocPath computes a prefixed filename for slug.md inside dir.
// If a collision occurs, it increments the numeric part until a free name is found.
func buildPrefixedDocPath(dir string, slug string) (string, error) {
	// Ensure slug has no leading numeric prefix already
	clean := slug
	if hasNumericPrefix(slug) {
		// strip to avoid double-prefixing
		b, _, _ := stripNumericPrefix(slug)
		clean = b
	}

	prefix, next, width, err := nextPrefixForDir(dir)
	if err != nil {
		return "", err
	}

	// Try up to a reasonable number of attempts
	for attempts := 0; attempts < 1000; attempts++ {
		name := fmt.Sprintf("%s%s.md", prefix, clean)
		path := filepath.Join(dir, name)
		if _, err := os.Stat(path); err != nil {
			if os.IsNotExist(err) {
				return path, nil
			}
			return "", err
		}
		// Collision: increment next and recompute prefix/width
		next++
		if next >= 100 {
			width = 3
		}
		prefix = fmt.Sprintf("%0*d-", width, next)
	}
	return "", fmt.Errorf("could not find free filename in %s for slug %s", dir, slug)
}

// (helper removed as unused)
