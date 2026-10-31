package canteen

import (
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// NormalizeName builds the comparison key used by the canteen unique
// constraints and by prefix/substring search. Full-width characters are folded
// to half-width, whitespace/punctuation/symbols are dropped and letters are
// lower-cased, so "麻辣烫 ！" and "麻辣烫!" share one key.
func NormalizeName(name string) string {
	folded := norm.NFKC.String(strings.TrimSpace(name))
	if folded == "" {
		return ""
	}

	var builder strings.Builder
	builder.Grow(len(folded))
	for _, r := range folded {
		switch {
		case unicode.IsSpace(r):
			continue
		case unicode.IsPunct(r):
			continue
		case unicode.IsSymbol(r):
			continue
		default:
			builder.WriteRune(unicode.ToLower(r))
		}
	}
	return builder.String()
}

// NormalizeCategories trims, de-duplicates and caps store categories while
// preserving the caller order.
func NormalizeCategories(categories []string) []string {
	if len(categories) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(categories))
	result := make([]string, 0, len(categories))
	for _, category := range categories {
		category = strings.TrimSpace(category)
		if category == "" {
			continue
		}
		if _, ok := seen[category]; ok {
			continue
		}
		seen[category] = struct{}{}
		result = append(result, category)
		if len(result) == MaxStoreCategories {
			break
		}
	}
	return result
}

// NameDistance returns the rune-level Levenshtein distance between two
// normalized names. It is only used to suggest an existing dish when search
// found nothing exact; it never merges entities automatically.
func NameDistance(left, right string) int {
	a := []rune(NormalizeName(left))
	b := []rune(NormalizeName(right))
	if len(a) == 0 || len(b) == 0 {
		return len(a) + len(b)
	}
	previous := make([]int, len(b)+1)
	current := make([]int, len(b)+1)
	for j := 0; j <= len(b); j++ {
		previous[j] = j
	}
	for i := 1; i <= len(a); i++ {
		current[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			current[j] = minInt(
				current[j-1]+1,
				previous[j]+1,
				previous[j-1]+cost,
			)
		}
		previous, current = current, previous
	}
	return previous[len(b)]
}

func minInt(values ...int) int {
	minimum := values[0]
	for _, value := range values[1:] {
		if value < minimum {
			minimum = value
		}
	}
	return minimum
}
