package skillbook

import (
	"fmt"
	"sort"
	"strings"
)

// confidenceRank maps confidence strings to numeric rank for sorting.
func confidenceRank(c string) int {
	switch c {
	case "high":
		return 3
	case "medium":
		return 2
	default: // "low" or unknown
		return 1
	}
}

// confidenceLabel returns the markdown label for a confidence level.
func confidenceLabel(c string) string {
	switch c {
	case "high":
		return "IMPORTANT"
	case "medium":
		return "Guideline"
	default:
		return "Consider"
	}
}

// selectSkills filters skills by agent (non-deprecated), sorts by confidence
// descending then appliedCount descending, and returns the top limit entries.
func selectSkills(skills []Skill, agent string, limit int) []Skill {
	var filtered []Skill
	for _, s := range skills {
		if s.Agent == agent && !s.Deprecated {
			filtered = append(filtered, s)
		}
	}

	sort.SliceStable(filtered, func(i, j int) bool {
		ri := confidenceRank(filtered[i].Confidence)
		rj := confidenceRank(filtered[j].Confidence)
		if ri != rj {
			return ri > rj
		}
		return filtered[i].AppliedCount > filtered[j].AppliedCount
	})

	if limit > 0 && len(filtered) > limit {
		filtered = filtered[:limit]
	}
	return filtered
}

// FormatMarkdown formats the top skills for the given agent as a markdown block.
// Returns empty string if no skills match (so the caller knows to skip injection).
func FormatMarkdown(skills []Skill, agent string, limit int) string {
	selected := selectSkills(skills, agent, limit)
	if len(selected) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("## Learned Guidelines (from project skillbook)\n\n")
	for _, s := range selected {
		label := confidenceLabel(s.Confidence)
		fmt.Fprintf(&sb, "- **%s:** %s\n", label, s.Rule)
	}
	return sb.String()
}

// IncrementApplied finds the same skills that FormatMarkdown would select,
// increments their appliedCount in the skillbook, and returns their IDs.
func IncrementApplied(sb *Skillbook, agent string, limit int) []string {
	selected := selectSkills(sb.Skills, agent, limit)
	if len(selected) == 0 {
		return nil
	}

	// Build a set of selected IDs.
	ids := make(map[string]struct{}, len(selected))
	for _, s := range selected {
		ids[s.ID] = struct{}{}
	}

	var result []string
	for i := range sb.Skills {
		if _, ok := ids[sb.Skills[i].ID]; ok {
			sb.Skills[i].AppliedCount++
			result = append(result, sb.Skills[i].ID)
		}
	}
	return result
}
