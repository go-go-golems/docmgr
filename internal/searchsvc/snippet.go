package searchsvc

import "strings"

// ExtractSnippet extracts a snippet of text around a query match.
//
// NOTE: This intentionally preserves the previous docmgr behavior for snippet generation.
func ExtractSnippet(content, query string, contextLen int) string {
	if query == "" {
		if len(content) <= contextLen {
			return content
		}
		return content[:contextLen] + "..."
	}

	queryLower := strings.ToLower(query)
	contentLower := strings.ToLower(content)

	idx := strings.Index(contentLower, queryLower)
	if idx == -1 {
		if len(content) <= contextLen {
			return content
		}
		return content[:contextLen] + "..."
	}

	start := idx - contextLen
	if start < 0 {
		start = 0
	}
	end := idx + len(query) + contextLen
	if end > len(content) {
		end = len(content)
	}

	snippet := content[start:end]
	if start > 0 {
		snippet = "..." + snippet
	}
	if end < len(content) {
		snippet = snippet + "..."
	}

	return snippet
}
