package skillbook

import (
	"testing"
)

func TestTokenize_ShortWordsStripped(t *testing.T) {
	tokens := Tokenize("if a or b")
	if len(tokens) != 0 {
		t.Errorf("expected empty token set, got %v", tokens)
	}
}

func TestJaccardSimilarity_EmptyStrings(t *testing.T) {
	a := Tokenize("")
	b := Tokenize("")
	got := JaccardSimilarity(a, b)
	if got != 0.0 {
		t.Errorf("expected 0.0 for empty strings, got %f", got)
	}
}

func TestJaccardSimilarity_IdenticalStrings(t *testing.T) {
	text := "always use descriptive variable names"
	a := Tokenize(text)
	b := Tokenize(text)
	got := JaccardSimilarity(a, b)
	if got != 1.0 {
		t.Errorf("expected 1.0 for identical strings, got %f", got)
	}
}

func TestJaccardSimilarity_KnownExample(t *testing.T) {
	// "prefer short functions over long ones" and "prefer short functions over long functions"
	// share 5 tokens out of 6 union tokens → Jaccard ≈ 0.83.
	a := Tokenize("prefer short functions over long ones")
	b := Tokenize("prefer short functions over long functions")
	got := JaccardSimilarity(a, b)
	if got <= 0.6 {
		t.Errorf("expected similarity > 0.6, got %f", got)
	}
}

func TestFindDuplicate_ReturnsIndexAboveThreshold(t *testing.T) {
	// Jaccard("prefer short functions over long ones",
	//         "prefer short functions over long functions") ≈ 0.83, well above 0.5.
	skills := []Skill{
		{ID: "a1", Agent: "coder", Rule: "prefer short functions over long ones"},
		{ID: "a2", Agent: "coder", Rule: "always use descriptive variable names"},
	}
	idx, found := FindDuplicate(skills, "coder", "prefer short functions over long functions", 0.5)
	if !found {
		t.Fatal("expected duplicate to be found")
	}
	if idx != 0 {
		t.Errorf("expected index 0, got %d", idx)
	}
}

func TestFindDuplicate_ReturnsFalseWhenBelowThreshold(t *testing.T) {
	skills := []Skill{
		{ID: "a1", Agent: "coder", Rule: "write short functions"},
	}
	_, found := FindDuplicate(skills, "coder", "always use descriptive variable names", 0.5)
	if found {
		t.Error("expected no duplicate to be found")
	}
}

func TestFindDuplicate_SkipsDeprecated(t *testing.T) {
	skills := []Skill{
		{ID: "a1", Agent: "coder", Rule: "always use descriptive variable names", Deprecated: true},
	}
	_, found := FindDuplicate(skills, "coder", "use descriptive naming for variables", 0.5)
	if found {
		t.Error("expected deprecated skill to be skipped")
	}
}

func TestFindDuplicate_SkipsWrongAgent(t *testing.T) {
	skills := []Skill{
		{ID: "a1", Agent: "reviewer", Rule: "always use descriptive variable names"},
	}
	_, found := FindDuplicate(skills, "coder", "use descriptive naming for variables", 0.5)
	if found {
		t.Error("expected skill with different agent to be skipped")
	}
}
