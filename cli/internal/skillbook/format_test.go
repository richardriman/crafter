package skillbook

import (
	"strings"
	"testing"
)

func makeSkill(id, agent, rule, confidence string, appliedCount int, deprecated bool) Skill {
	return Skill{
		ID:           id,
		Agent:        agent,
		Rule:         rule,
		Confidence:   confidence,
		AppliedCount: appliedCount,
		Deprecated:   deprecated,
	}
}

func TestFormatMarkdown_NoMatchingSkills_ReturnsEmpty(t *testing.T) {
	skills := []Skill{
		makeSkill("a1", "reviewer", "some rule", "high", 0, false),
	}
	got := FormatMarkdown(skills, "coder", 10)
	if got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestFormatMarkdown_SortedByConfidenceThenAppliedCount(t *testing.T) {
	skills := []Skill{
		makeSkill("low1", "coder", "low confidence rule", "low", 5, false),
		makeSkill("med1", "coder", "medium confidence rule", "medium", 1, false),
		makeSkill("high1", "coder", "high confidence rule", "high", 0, false),
		makeSkill("med2", "coder", "another medium rule", "medium", 3, false),
	}
	result := FormatMarkdown(skills, "coder", 10)

	// high must appear before medium, medium before low
	highPos := strings.Index(result, "high confidence rule")
	med2Pos := strings.Index(result, "another medium rule")
	med1Pos := strings.Index(result, "medium confidence rule")
	lowPos := strings.Index(result, "low confidence rule")

	if highPos == -1 || med2Pos == -1 || med1Pos == -1 || lowPos == -1 {
		t.Fatal("one or more expected rules missing from output")
	}

	if !(highPos < med2Pos) {
		t.Error("high confidence should appear before medium (higher appliedCount)")
	}
	if !(med2Pos < med1Pos) {
		t.Error("medium with appliedCount=3 should appear before medium with appliedCount=1")
	}
	if !(med1Pos < lowPos) {
		t.Error("medium confidence should appear before low confidence")
	}
}

func TestFormatMarkdown_LimitTopN(t *testing.T) {
	skills := []Skill{
		makeSkill("s1", "coder", "rule one", "high", 0, false),
		makeSkill("s2", "coder", "rule two", "high", 0, false),
		makeSkill("s3", "coder", "rule three", "medium", 0, false),
		makeSkill("s4", "coder", "rule four", "medium", 0, false),
		makeSkill("s5", "coder", "rule five", "low", 0, false),
	}
	result := FormatMarkdown(skills, "coder", 2)

	lines := []string{}
	for _, line := range strings.Split(result, "\n") {
		if strings.HasPrefix(line, "- ") {
			lines = append(lines, line)
		}
	}
	if len(lines) != 2 {
		t.Errorf("expected 2 skill lines with limit=2, got %d", len(lines))
	}

	// Content check: only the top-2 high-confidence rules should appear.
	if !strings.Contains(result, "rule one") {
		t.Error("expected 'rule one' (high confidence) to appear in limited output")
	}
	if !strings.Contains(result, "rule two") {
		t.Error("expected 'rule two' (high confidence) to appear in limited output")
	}
	// Lower-ranked rules must not appear.
	if strings.Contains(result, "rule three") || strings.Contains(result, "rule four") || strings.Contains(result, "rule five") {
		t.Error("expected rules beyond top-2 limit to be absent from output")
	}
}

func TestFormatMarkdown_DeprecatedSkillsExcluded(t *testing.T) {
	skills := []Skill{
		makeSkill("s1", "coder", "active rule", "high", 0, false),
		makeSkill("s2", "coder", "deprecated rule", "high", 0, true),
	}
	result := FormatMarkdown(skills, "coder", 10)

	if strings.Contains(result, "deprecated rule") {
		t.Error("deprecated rule should not appear in output")
	}
	if !strings.Contains(result, "active rule") {
		t.Error("active rule should appear in output")
	}
}

func TestFormatMarkdown_ConfidencePrefixes(t *testing.T) {
	skills := []Skill{
		makeSkill("h1", "coder", "high rule", "high", 0, false),
		makeSkill("m1", "coder", "medium rule", "medium", 0, false),
		makeSkill("l1", "coder", "low rule", "low", 0, false),
	}
	result := FormatMarkdown(skills, "coder", 10)

	if !strings.Contains(result, "**IMPORTANT:** high rule") {
		t.Errorf("expected IMPORTANT prefix for high confidence, result:\n%s", result)
	}
	if !strings.Contains(result, "**Guideline:** medium rule") {
		t.Errorf("expected Guideline prefix for medium confidence, result:\n%s", result)
	}
	if !strings.Contains(result, "**Consider:** low rule") {
		t.Errorf("expected Consider prefix for low confidence, result:\n%s", result)
	}
}

func TestIncrementApplied_IncrementsOnlySelectedSkills(t *testing.T) {
	sb := &Skillbook{
		Version: 1,
		Skills: []Skill{
			makeSkill("s1", "coder", "high rule", "high", 0, false),
			makeSkill("s2", "coder", "low rule", "low", 0, false),
			makeSkill("s3", "reviewer", "reviewer rule", "high", 0, false),
		},
	}

	// With limit=1, only the top skill (s1, high confidence) should be selected.
	ids := IncrementApplied(sb, "coder", 1)

	if len(ids) != 1 || ids[0] != "s1" {
		t.Errorf("expected [s1], got %v", ids)
	}

	// s1 must be incremented.
	if sb.Skills[0].AppliedCount != 1 {
		t.Errorf("s1 appliedCount: expected 1, got %d", sb.Skills[0].AppliedCount)
	}
	// s2 must not be touched.
	if sb.Skills[1].AppliedCount != 0 {
		t.Errorf("s2 appliedCount: expected 0, got %d", sb.Skills[1].AppliedCount)
	}
	// s3 (different agent) must not be touched.
	if sb.Skills[2].AppliedCount != 0 {
		t.Errorf("s3 appliedCount: expected 0, got %d", sb.Skills[2].AppliedCount)
	}
}

func TestIncrementApplied_NoMatchReturnsEmpty(t *testing.T) {
	sb := &Skillbook{
		Version: 1,
		Skills:  []Skill{},
	}
	ids := IncrementApplied(sb, "coder", 10)
	if len(ids) != 0 {
		t.Errorf("expected empty slice, got %v", ids)
	}
}
