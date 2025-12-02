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

// SlugifyTitleForTicket generates a slug from a title while avoiding duplicating
// the ticket identifier. If the title begins with the ticket identifier (a
// common pattern like "MEN-1234: Title"), the ticket prefix and any separators
// are removed before slugifying. When the stripped title becomes empty, the
// ticket identifier itself is slugified as a fallback.
func SlugifyTitleForTicket(ticket, title string) string {
	cleanTicket := strings.TrimSpace(ticket)
	cleanTitle := strings.TrimSpace(title)

	if cleanTitle == "" {
		return Slugify(cleanTicket)
	}

	stripped := stripTicketPrefix(cleanTitle, cleanTicket)
	if stripped == "" {
		if cleanTicket == "" {
			return Slugify(cleanTitle)
		}
		return Slugify(cleanTicket)
	}
	return Slugify(stripped)
}

func stripTicketPrefix(title, ticket string) string {
	if title == "" || ticket == "" {
		return title
	}

	trimmedTicket := strings.TrimSpace(ticket)
	if trimmedTicket == "" || len(title) < len(trimmedTicket) {
		return title
	}

	if !strings.EqualFold(title[:len(trimmedTicket)], trimmedTicket) {
		return title
	}

	remainder := title[len(trimmedTicket):]
	remainder = strings.TrimLeftFunc(remainder, func(r rune) bool {
		switch r {
		case ' ', '\t', '-', '–', '—', ':', '_':
			return true
		default:
			return false
		}
	})
	return strings.TrimSpace(remainder)
}
