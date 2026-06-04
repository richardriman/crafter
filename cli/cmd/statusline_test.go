package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
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

// withStdin redirects os.Stdin to a temp file pre-loaded with content for the
// duration of fn, then restores it. A nil pointer is restored even if fn panics
// so a failing case cannot corrupt sibling tests.
func withStdin(t *testing.T, content string, fn func()) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "stdin")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("writing stdin fixture: %v", err)
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("opening stdin fixture: %v", err)
	}
	defer f.Close()

	orig := os.Stdin
	os.Stdin = f
	defer func() { os.Stdin = orig }()

	fn()
}

// TestRunStatusline_NeverBreaksStatusBar is the durable guard for Step 3.2: the
// statusline command must never break the status bar. It drives the command's
// inner run function (runStatusline) — the exact boundary whose return value
// decides the process exit code (cmd.Execute calls os.Exit(1) only on a non-nil
// error) — across the three degraded inputs and asserts that none returns an
// error (→ exit 0) and none panics. Output degrades to whatever data exists,
// possibly the empty string.
func TestRunStatusline_NeverBreaksStatusBar(t *testing.T) {
	// nonGitDir is a workspace with neither a .git repo nor a .crafter context,
	// used for the non-Crafter / non-git directory case.
	nonGitDir := t.TempDir()

	tests := []struct {
		name string
		// stdin is the payload fed to runStatusline.
		stdin string
		// chdir, when true, points the process working directory at a fresh
		// non-git, non-Crafter temp dir for the subtest. The malformed-JSON and
		// empty-stdin payloads carry no workspace.current_dir, so runStatusline
		// falls back to os.Getwd(); without this the fallback would resolve to
		// the live Crafter repo and render a full panel instead of exercising
		// the degraded path these cases name.
		chdir bool
	}{
		{
			name:  "malformed JSON",
			stdin: "garbage",
			chdir: true,
		},
		{
			name:  "empty stdin",
			stdin: "",
			chdir: true,
		},
		{
			name: "non-Crafter / non-git directory",
			// Valid JSON pointing current_dir at a directory with no .git and no
			// .crafter — the renderer must degrade rather than error.
			stdin: `{"workspace": {"current_dir": "` + nonGitDir + `"}}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.chdir {
				// t.Chdir (Go 1.24+) switches the process cwd and auto-restores it
				// when the subtest ends, so the os.Getwd() fallback degrades to a
				// non-git, non-Crafter dir regardless of where the suite runs.
				t.Chdir(t.TempDir())
			}
			withStdin(t, tc.stdin, func() {
				defer func() {
					if r := recover(); r != nil {
						t.Fatalf("runStatusline panicked: %v", r)
					}
				}()
				// A nil error propagates up through rootCmd.Execute, so cmd.Execute
				// never reaches os.Exit(1): the process exits 0.
				if err := runStatusline(statuslineCmd, nil); err != nil {
					t.Errorf("runStatusline returned a non-nil error (would break the status bar): %v", err)
				}
			})
		})
	}
}
