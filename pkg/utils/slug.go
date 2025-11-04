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
