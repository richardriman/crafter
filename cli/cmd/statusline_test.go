package cmd

import (
	"encoding/json"
	"testing"
)

// TestStatuslineInput_DecodeFull verifies that a representative Claude Code
// statusLine payload decodes into every field statuslineInput exposes, and that
// pointer fields carry their real values.
func TestStatuslineInput_DecodeFull(t *testing.T) {
	raw := `{
		"model": {"display_name": "Opus 4.8"},
		"effort": {"level": "high"},
		"context_window": {"used_percentage": 42.5, "context_window_size": 1000000},
		"cost": {"total_cost_usd": 0.42, "total_lines_added": 120, "total_lines_removed": 30},
		"workspace": {"current_dir": "/repo", "project_dir": "/repo/crafter"}
	}`

	var in statuslineInput
	if err := json.Unmarshal([]byte(raw), &in); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if in.Model.DisplayName != "Opus 4.8" {
		t.Errorf("display_name: got %q, want %q", in.Model.DisplayName, "Opus 4.8")
	}
	if in.Effort.Level != "high" {
		t.Errorf("effort.level: got %q, want %q", in.Effort.Level, "high")
	}
	if in.ContextWindow.UsedPercentage == nil {
		t.Fatalf("used_percentage: got nil, want non-nil pointer to 42.5")
	}
	if *in.ContextWindow.UsedPercentage != 42.5 {
		t.Errorf("used_percentage: got %v, want 42.5", *in.ContextWindow.UsedPercentage)
	}
	if in.ContextWindow.ContextWindowSize != 1000000 {
		t.Errorf("context_window_size: got %d, want 1000000", in.ContextWindow.ContextWindowSize)
	}
	if in.Cost.TotalCostUSD == nil {
		t.Fatalf("total_cost_usd: got nil, want non-nil pointer to 0.42")
	}
	if *in.Cost.TotalCostUSD != 0.42 {
		t.Errorf("total_cost_usd: got %v, want 0.42", *in.Cost.TotalCostUSD)
	}
	if in.Cost.TotalLinesAdded != 120 {
		t.Errorf("total_lines_added: got %d, want 120", in.Cost.TotalLinesAdded)
	}
	if in.Cost.TotalLinesRemoved != 30 {
		t.Errorf("total_lines_removed: got %d, want 30", in.Cost.TotalLinesRemoved)
	}
	if in.Workspace.CurrentDir != "/repo" {
		t.Errorf("current_dir: got %q, want %q", in.Workspace.CurrentDir, "/repo")
	}
	if in.Workspace.ProjectDir != "/repo/crafter" {
		t.Errorf("project_dir: got %q, want %q", in.Workspace.ProjectDir, "/repo/crafter")
	}
}

// TestStatuslineInput_DecodeEffortAbsent verifies that an absent effort.level
// decodes to the zero value (empty string), distinguishing "no effort" from any
// present level.
func TestStatuslineInput_DecodeEffortAbsent(t *testing.T) {
	raw := `{"model": {"display_name": "Opus 4.8"}}`

	var in statuslineInput
	if err := json.Unmarshal([]byte(raw), &in); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if in.Effort.Level != "" {
		t.Errorf("absent effort.level: got %q, want empty string", in.Effort.Level)
	}
}

// TestStatuslineInput_DecodeUsedPercentageNullAndAbsent verifies that a null or
// absent context_window.used_percentage decodes to a nil pointer — observably
// distinct from a present value (including a present 0).
func TestStatuslineInput_DecodeUsedPercentageNullAndAbsent(t *testing.T) {
	tests := []struct {
		name string
		raw  string
	}{
		{name: "null", raw: `{"context_window": {"used_percentage": null}}`},
		{name: "absent", raw: `{"context_window": {}}`},
		{name: "section absent", raw: `{}`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var in statuslineInput
			if err := json.Unmarshal([]byte(tc.raw), &in); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if in.ContextWindow.UsedPercentage != nil {
				t.Errorf("used_percentage: got %v, want nil", *in.ContextWindow.UsedPercentage)
			}
		})
	}

	// A present 0 must NOT be nil — it is a real value distinct from null/absent.
	var in statuslineInput
	if err := json.Unmarshal([]byte(`{"context_window": {"used_percentage": 0}}`), &in); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if in.ContextWindow.UsedPercentage == nil {
		t.Fatalf("present used_percentage 0: got nil, want non-nil pointer to 0")
	}
	if *in.ContextWindow.UsedPercentage != 0 {
		t.Errorf("present used_percentage 0: got %v, want 0", *in.ContextWindow.UsedPercentage)
	}
}

// TestStatuslineInput_DecodeTotalCostPositiveZeroAbsent verifies that the
// total_cost_usd pointer distinguishes a positive value, a present zero, and an
// absent field — the three cases the cost section depends on (A6).
func TestStatuslineInput_DecodeTotalCostPositiveZeroAbsent(t *testing.T) {
	// Positive.
	var pos statuslineInput
	if err := json.Unmarshal([]byte(`{"cost": {"total_cost_usd": 0.42}}`), &pos); err != nil {
		t.Fatalf("unmarshal positive: %v", err)
	}
	if pos.Cost.TotalCostUSD == nil {
		t.Fatalf("positive total_cost_usd: got nil, want non-nil")
	}
	if *pos.Cost.TotalCostUSD != 0.42 {
		t.Errorf("positive total_cost_usd: got %v, want 0.42", *pos.Cost.TotalCostUSD)
	}

	// Present zero — non-nil pointer to 0, distinct from absent.
	var zero statuslineInput
	if err := json.Unmarshal([]byte(`{"cost": {"total_cost_usd": 0}}`), &zero); err != nil {
		t.Fatalf("unmarshal zero: %v", err)
	}
	if zero.Cost.TotalCostUSD == nil {
		t.Fatalf("present zero total_cost_usd: got nil, want non-nil pointer to 0")
	}
	if *zero.Cost.TotalCostUSD != 0 {
		t.Errorf("present zero total_cost_usd: got %v, want 0", *zero.Cost.TotalCostUSD)
	}

	// Absent — nil pointer.
	var absent statuslineInput
	if err := json.Unmarshal([]byte(`{"cost": {}}`), &absent); err != nil {
		t.Fatalf("unmarshal absent: %v", err)
	}
	if absent.Cost.TotalCostUSD != nil {
		t.Errorf("absent total_cost_usd: got %v, want nil", *absent.Cost.TotalCostUSD)
	}
}
