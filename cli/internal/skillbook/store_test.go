package skillbook

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestNewSkillbook_DefaultValues(t *testing.T) {
	sb := NewSkillbook()
	if sb.Version != 1 {
		t.Errorf("expected version 1, got %d", sb.Version)
	}
	if sb.Skills == nil {
		t.Error("expected non-nil skills slice")
	}
	if len(sb.Skills) != 0 {
		t.Errorf("expected empty skills slice, got length %d", len(sb.Skills))
	}
}

func TestLoad_NonExistentFile_ReturnsEmptySkillbook(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "does_not_exist.json")

	sb, err := Load(path)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if sb == nil {
		t.Fatal("expected non-nil skillbook")
	}
	if sb.Version != 1 {
		t.Errorf("expected version 1, got %d", sb.Version)
	}
	if len(sb.Skills) != 0 {
		t.Errorf("expected empty skills, got %d", len(sb.Skills))
	}
}

func TestSaveAndLoad_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "skillbook.json")

	original := &Skillbook{
		Version: 1,
		Skills: []Skill{
			{
				ID:           "abc123",
				Agent:        "coder",
				Rule:         "always use descriptive variable names",
				Rationale:    "improves readability",
				SourceTask:   "task-1",
				Confidence:   "high",
				AddedAt:      "2026-01-01T00:00:00Z",
				AppliedCount: 3,
				Deprecated:   false,
			},
		},
	}

	if err := Save(path, original); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.Version != original.Version {
		t.Errorf("version: expected %d, got %d", original.Version, loaded.Version)
	}
	if len(loaded.Skills) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(loaded.Skills))
	}

	s := loaded.Skills[0]
	orig := original.Skills[0]

	// Safety net: catch any newly added fields that lack an individual assertion below.
	if !reflect.DeepEqual(s, orig) {
		t.Errorf("loaded skill does not deep-equal original:\n  got:  %+v\n  want: %+v", s, orig)
	}

	if s.ID != orig.ID {
		t.Errorf("ID: expected %q, got %q", orig.ID, s.ID)
	}
	if s.Agent != orig.Agent {
		t.Errorf("Agent: expected %q, got %q", orig.Agent, s.Agent)
	}
	if s.Rule != orig.Rule {
		t.Errorf("Rule: expected %q, got %q", orig.Rule, s.Rule)
	}
	if s.Rationale != orig.Rationale {
		t.Errorf("Rationale: expected %q, got %q", orig.Rationale, s.Rationale)
	}
	if s.SourceTask != orig.SourceTask {
		t.Errorf("SourceTask: expected %q, got %q", orig.SourceTask, s.SourceTask)
	}
	if s.Confidence != orig.Confidence {
		t.Errorf("Confidence: expected %q, got %q", orig.Confidence, s.Confidence)
	}
	if s.AddedAt != orig.AddedAt {
		t.Errorf("AddedAt: expected %q, got %q", orig.AddedAt, s.AddedAt)
	}
	if s.AppliedCount != orig.AppliedCount {
		t.Errorf("AppliedCount: expected %d, got %d", orig.AppliedCount, s.AppliedCount)
	}
	if s.Deprecated != orig.Deprecated {
		t.Errorf("Deprecated: expected %v, got %v", orig.Deprecated, s.Deprecated)
	}
}

func TestSave_TmpFileCleanedUp(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "skillbook.json")

	sb := NewSkillbook()
	if err := Save(path, sb); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	tmpPath := path + ".tmp"
	if _, err := os.Stat(tmpPath); !os.IsNotExist(err) {
		t.Errorf("expected .tmp file to be removed after Save, but it still exists")
	}
}
