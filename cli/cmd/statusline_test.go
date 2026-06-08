package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
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

// TestRunStatusline_PipePath_FullPayload verifies the pipe path (stdin is a
// regular file, not a tty) with a complete valid payload: runStatusline must
// read the payload, not error, and not panic. This is the happy-path complement
// to the degraded cases in TestRunStatusline_NeverBreaksStatusBar.
func TestRunStatusline_PipePath_FullPayload(t *testing.T) {
	payload := `{
		"model": {"display_name": "Sonnet 4.6"},
		"effort": {"level": "normal"},
		"context_window": {"used_percentage": 12.3, "context_window_size": 200000},
		"cost": {"total_cost_usd": 0.05, "total_lines_added": 10, "total_lines_removed": 2},
		"workspace": {"current_dir": "` + t.TempDir() + `", "project_dir": ""}
	}`
	withStdin(t, payload, func() {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("runStatusline panicked: %v", r)
			}
		}()
		if err := runStatusline(statuslineCmd, nil); err != nil {
			t.Errorf("pipe path with full payload returned non-nil error: %v", err)
		}
	})
}

// TestRunStatusline_PipePath_BlocksOnNeverEOF demonstrates that the pipe path
// (stdin is a pipe, not a tty) causes runStatusline to block when the write end
// of the pipe is never closed — that is, io.ReadAll is entered and does not
// return until EOF.
//
// This is the strongest deterministic test of the tty guard achievable without
// a real terminal file descriptor (which is not allocatable under go test) or a
// production seam (which is out of scope for this step). It proves:
//
//   - The pipe path (ModeCharDevice == 0) always enters the blocking io.ReadAll.
//   - Therefore the tty guard (ModeCharDevice set) is the necessary and
//     sufficient mechanism that prevents blocking on an interactive terminal —
//     its absence would cause exactly the hang this test observes.
//
// Limitation: because go test cannot allocate a real tty fd, we cannot
// directly trigger the ModeCharDevice branch and assert it returns promptly.
// The guard logic (statusline.go) has been reviewed by inspection; this test
// provides the complementary behavioral proof from the pipe side.
func TestRunStatusline_PipePath_BlocksOnNeverEOF(t *testing.T) {
	// Switch to a temp dir so the goroutine that unblocks after pw.Close() does
	// not resolve the live repo's git/plan state and emit an incidental panel
	// line to stdout. This is the same hygiene TestRunStatusline_NeverBreaksStatusBar
	// already applies for its chdir cases.
	t.Chdir(t.TempDir())

	pr, pw, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	// pw is intentionally never written to or closed for the duration of the
	// test: io.ReadAll on pr will block until the write end closes.
	defer pw.Close()
	defer pr.Close()

	orig := os.Stdin
	os.Stdin = pr
	defer func() { os.Stdin = orig }()

	done := make(chan error, 1)
	go func() {
		done <- runStatusline(statuslineCmd, nil)
	}()

	const deadline = 80 * time.Millisecond
	select {
	case err := <-done:
		// runStatusline returned before EOF — this means it did NOT block on
		// the pipe. That would indicate the tty guard was triggered via some
		// unexpected mechanism, which would require investigation.
		t.Errorf("runStatusline returned before stdin EOF (unexpected early return, err=%v); "+
			"expected it to block on the never-EOF pipe — if the tty guard is triggering "+
			"on a pipe fd, the guard logic has changed", err)
	case <-time.After(deadline):
		// Expected path: the goroutine is still blocked inside io.ReadAll,
		// confirming the pipe path reads stdin and that the tty guard is what
		// prevents this block for real character devices.
	}

	// Unblock the goroutine so it does not leak: close the write end so
	// io.ReadAll gets EOF and the goroutine can exit cleanly before the test
	// completes.
	pw.Close()
	select {
	case <-done:
		// goroutine exited cleanly
	case <-time.After(2 * time.Second):
		t.Error("goroutine did not exit after write-end close — possible goroutine leak")
	}
}
