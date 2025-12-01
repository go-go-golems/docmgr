// Package utils provides utility functions for docmgr.
//
// The Slugify function converts arbitrary strings into filesystem-friendly slugs
// suitable for use in filenames and directory names.
package utils

import (
	"strings"
)

// Slugify converts an arbitrary string into a filesystem-friendly slug:
// - lowercases
// - replaces any non [a-z0-9] with '-'
// - collapses consecutive '-'
// - trims leading/trailing '-'
func Slugify(input string) string {
	s := strings.ToLower(input)
	var b strings.Builder
	prevHyphen := false
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			prevHyphen = false
			continue
		}
		if !prevHyphen {
			b.WriteByte('-')
			prevHyphen = true
		}
	}
	out := b.String()
	out = strings.Trim(out, "-")
	if out == "" {
		return "document"
	}
	return out
}

// StripTicketFromTitle removes common ticket identifier patterns from the beginning of a title
// before slugifying. This prevents duplicate ticket identifiers in directory names.
// Handles patterns like "TICKET:", "TICKET -", "TICKET ".
func StripTicketFromTitle(title, ticket string) string {
	if ticket == "" {
		return title
	}
	title = strings.TrimSpace(title)
	patterns := []string{
		ticket + ":",
		ticket + " -",
		ticket + " ",
	}
	for _, pattern := range patterns {
		if strings.HasPrefix(title, pattern) {
			cleaned := strings.TrimSpace(strings.TrimPrefix(title, pattern))
			if cleaned != "" {
				return cleaned
			}
		}
	}
	return title
}
