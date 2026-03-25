package skillbook

import (
	"strings"
	"unicode"
)

// Tokenize converts text into a set of tokens.
// It lowercases all characters, strips non-alphanumeric characters (hyphens
// are kept), splits on whitespace, and discards tokens with length <= 2.
func Tokenize(text string) map[string]struct{} {
	lower := strings.ToLower(text)

	var b strings.Builder
	for _, r := range lower {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || unicode.IsSpace(r) {
			b.WriteRune(r)
		} else {
			b.WriteRune(' ')
		}
	}

	tokens := make(map[string]struct{})
	for _, tok := range strings.Fields(b.String()) {
		if len(tok) > 2 {
			tokens[tok] = struct{}{}
		}
	}
	return tokens
}

// JaccardSimilarity computes |intersection| / |union| for two token sets.
// Returns 0.0 if both sets are empty.
func JaccardSimilarity(a, b map[string]struct{}) float64 {
	if len(a) == 0 && len(b) == 0 {
		return 0.0
	}

	intersection := 0
	for tok := range a {
		if _, ok := b[tok]; ok {
			intersection++
		}
	}

	union := len(a) + len(b) - intersection
	return float64(intersection) / float64(union)
}

// FindDuplicate searches skills for the first non-deprecated entry matching
// agent whose rule has Jaccard similarity > threshold with rule.
// Returns the index in skills and true if found, or 0 and false otherwise.
func FindDuplicate(skills []Skill, agent, rule string, threshold float64) (int, bool) {
	ruleTokens := Tokenize(rule)
	for i, s := range skills {
		if s.Deprecated || s.Agent != agent {
			continue
		}
		if JaccardSimilarity(ruleTokens, Tokenize(s.Rule)) > threshold {
			return i, true
		}
	}
	return -1, false
}
