// Package matching turns raw ingredient strings into Ingredient Terms and
// ranks Recipes by how much of each one a pantry covers.
//
// docs/ingredient-matching-algorithm.md specifies the algorithm and is the
// source of truth for it; this package is one of its two implementations. It
// deliberately depends on nothing outside the standard library — no database,
// no HTTP, no filesystem — so that scoring lives in one testable place.
package matching

import (
	"sort"
	"strings"
	"unicode"
)

// A Term is an Ingredient Term: the set of tokens one raw string normalizes
// to. Order is never significant, so the tokens are held sorted and
// deduplicated, which also gives a Term a stable key for deduplicating Terms
// against one another.
type Term struct {
	tokens []string
}

// Normalize turns one raw string — a Pantry Item as the user typed it, or an
// Ingredient's name as the Author wrote it — into an Ingredient Term, by the
// steps of section 1 of the algorithm document, in that order. The result may
// be the empty Term; that is a legal outcome.
func Normalize(raw string) Term {
	lowered := strings.ToLower(raw)

	fields := strings.FieldsFunc(lowered, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})

	seen := make(map[string]struct{}, len(fields))
	tokens := make([]string, 0, len(fields))
	for _, field := range fields {
		field = strings.Map(func(r rune) rune {
			if unicode.IsDigit(r) {
				return -1
			}
			return r
		}, field)

		if field == "" {
			continue
		}
		if isStopword(field) {
			continue
		}
		if _, duplicate := seen[field]; duplicate {
			continue
		}
		seen[field] = struct{}{}
		tokens = append(tokens, field)
	}
	sort.Strings(tokens)

	return Term{tokens: tokens}
}

// Tokens returns the Term's tokens, sorted and deduplicated.
func (t Term) Tokens() []string {
	return t.tokens
}

// IsEmpty reports whether normalization left no tokens at all, which happens
// for a string that is only punctuation, digits, measurements or descriptors.
func (t Term) IsEmpty() bool {
	return len(t.tokens) == 0
}

// key identifies the Term's token set, so two Ingredients that normalize the
// same way can be counted once.
func (t Term) key() string {
	return strings.Join(t.tokens, "\x00")
}
