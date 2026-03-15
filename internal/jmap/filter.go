package jmap

import (
	"fmt"
	"strings"
	"time"
)

// knownKeywords are the recognized filter prefixes. Order matters for parsing:
// we check these when splitting the query into tokens.
var knownKeywords = []string{"from:", "to:", "subject:", "in:", "before:", "after:", "has:"}

// ParseFilterQuery parses a human-friendly search query into a SearchFilter.
//
// Supported syntax:
//
//	from:user@example.com      → SearchFilter.From
//	to:recipient@example.com   → SearchFilter.To
//	subject:meeting notes      → SearchFilter.Subject (value runs until next keyword or end)
//	in:INBOX                   → SearchFilter.InMailbox (mailbox name, resolved to ID later)
//	before:2025-01-01          → SearchFilter.Before
//	after:2025-01-01           → SearchFilter.After
//	has:attachment              → SearchFilter.HasAttachment = true
//	<anything else>            → SearchFilter.Text (free text search)
func ParseFilterQuery(query string) (SearchFilter, error) {
	var filter SearchFilter

	query = strings.TrimSpace(query)
	if query == "" {
		return filter, nil
	}

	// We tokenize the query by splitting on spaces, then reassemble values
	// for keywords whose values may contain spaces (like subject:).
	tokens := strings.Fields(query)
	var freeText []string

	for i := 0; i < len(tokens); i++ {
		token := tokens[i]
		lower := strings.ToLower(token)

		switch {
		case strings.HasPrefix(lower, "from:"):
			filter.From = token[len("from:"):]

		case strings.HasPrefix(lower, "to:"):
			filter.To = token[len("to:"):]

		case strings.HasPrefix(lower, "subject:"):
			// subject: value continues until the next known keyword or end
			value := token[len("subject:"):]
			var parts []string
			if value != "" {
				parts = append(parts, value)
			}
			for i+1 < len(tokens) && !isKeyword(tokens[i+1]) {
				i++
				parts = append(parts, tokens[i])
			}
			filter.Subject = strings.Join(parts, " ")

		case strings.HasPrefix(lower, "in:"):
			filter.InMailbox = token[len("in:"):]

		case strings.HasPrefix(lower, "before:"):
			dateStr := token[len("before:"):]
			t, err := time.Parse("2006-01-02", dateStr)
			if err != nil {
				return filter, fmt.Errorf("invalid date format for before: %q (expected YYYY-MM-DD)", dateStr)
			}
			filter.Before = &t

		case strings.HasPrefix(lower, "after:"):
			dateStr := token[len("after:"):]
			t, err := time.Parse("2006-01-02", dateStr)
			if err != nil {
				return filter, fmt.Errorf("invalid date format for after: %q (expected YYYY-MM-DD)", dateStr)
			}
			filter.After = &t

		case strings.HasPrefix(lower, "has:"):
			value := strings.ToLower(token[len("has:"):])
			if value == "attachment" {
				filter.HasAttachment = true
			}

		default:
			freeText = append(freeText, token)
		}
	}

	if len(freeText) > 0 {
		filter.Text = strings.Join(freeText, " ")
	}

	return filter, nil
}

// isKeyword reports whether the token starts with a known filter keyword prefix.
func isKeyword(token string) bool {
	lower := strings.ToLower(token)
	for _, kw := range knownKeywords {
		if strings.HasPrefix(lower, kw) {
			return true
		}
	}
	return false
}
