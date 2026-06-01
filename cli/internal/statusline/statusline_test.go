package statusline

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// writeFile writes content to path, creating parent directories as needed.
func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("creating directories for %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("writing %s: %v", path, err)
	}
}

// writePlanFile writes a minimal task file whose ## Plan section has the given
// planBody, and returns the file path.
func writePlanFile(t *testing.T, planBody string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "task.md")
	content := "## Plan\n" + planBody + "\n"
	writeFile(t, path, content)
	return path
}

// ---------------------------------------------------------------------------
// renderSegment / parsePlan — plan states
// ---------------------------------------------------------------------------

func TestRenderSegment_States(t *testing.T) {
	tests := []struct {
		name string
		info planInfo
		want string
	}{
		{
			name: "draft → awaiting approval",
			info: planInfo{state: planStateDraft},
			want: "crafter · plan: awaiting approval",
		},
		{
			name: "none (pending) → planning",
			info: planInfo{state: planStateNone},
			want: "crafter · planning",
		},
		{
			name: "approved no phases no steps → crafter only",
			info: planInfo{state: planStateApproved},
			want: "crafter",
		},
		{
			name: "approved with phases and steps executing",
			info: planInfo{
				state:        planStateApproved,
				totalPhases:  3,
				currentPhase: 2,
				doneSteps:    7,
				totalSteps:   12,
			},
			// 7/12 = 58.3% → round → 58%; floor(58/10) = 5 filled glyphs.
			want: "crafter · Phase 2/3 · 7/12 [█████░░░░░] 58%",
		},
		{
			name: "approved all done → 100% full bar",
			info: planInfo{
				state:        planStateApproved,
				totalPhases:  2,
				currentPhase: 2,
				doneSteps:    4,
				totalSteps:   4,
			},
			want: "crafter · Phase 2/2 · 4/4 [██████████] 100%",
		},
		{
			name: "approved 0 done → empty bar",
			info: planInfo{
				state:        planStateApproved,
				totalPhases:  2,
				currentPhase: 1,
				doneSteps:    0,
				totalSteps:   5,
			},
			want: "crafter · Phase 1/2 · 0/5 [░░░░░░░░░░] 0%",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := renderSegment(tc.info)
			if got != tc.want {
				t.Errorf("renderSegment mismatch:\n  got  %q\n  want %q", got, tc.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Percent / bar math
// ---------------------------------------------------------------------------

func TestPercentAndBar(t *testing.T) {
	tests := []struct {
		name        string
		done, total int
		wantPct     int
		wantFilled  int
		wantSegment string
	}{
		{
			name:        "7/12 → 58% 5 filled",
			done:        7,
			total:       12,
			wantPct:     58,
			wantFilled:  5,
			wantSegment: "crafter · Phase 1/1 · 7/12 [█████░░░░░] 58%",
		},
		{
			name:        "1/3 → 33% 3 filled",
			done:        1,
			total:       3,
			wantPct:     33,
			wantFilled:  3,
			wantSegment: "crafter · Phase 1/1 · 1/3 [███░░░░░░░] 33%",
		},
		{
			name:        "all done → 100% 10 filled",
			done:        3,
			total:       3,
			wantPct:     100,
			wantFilled:  10,
			wantSegment: "crafter · Phase 1/1 · 3/3 [██████████] 100%",
		},
		{
			name:        "0 done → 0% 0 filled",
			done:        0,
			total:       6,
			wantPct:     0,
			wantFilled:  0,
			wantSegment: "crafter · Phase 1/1 · 0/6 [░░░░░░░░░░] 0%",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			info := planInfo{
				state:        planStateApproved,
				totalPhases:  1,
				currentPhase: 1,
				doneSteps:    tc.done,
				totalSteps:   tc.total,
			}
			got := renderSegment(info)
			if got != tc.wantSegment {
				t.Errorf("renderSegment mismatch:\n  got  %q\n  want %q", got, tc.wantSegment)
			}
			// Spot-check bar internals via string counting.
			barStart := strings.Index(got, "[")
			barEnd := strings.Index(got, "]")
			if barStart < 0 || barEnd < 0 {
				t.Fatalf("bar brackets not found in %q", got)
			}
			barContent := got[barStart+1 : barEnd]
			var filledCount int
			for _, r := range barContent {
				if string(r) == glyphFilled {
					filledCount++
				}
			}
			if filledCount != tc.wantFilled {
				t.Errorf("filled glyphs: got %d, want %d in %q", filledCount, tc.wantFilled, got)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// parsePlan — state parsing from file content
// ---------------------------------------------------------------------------

func TestParsePlan_States(t *testing.T) {
	tests := []struct {
		name      string
		planBody  string
		wantState planState
	}{
		{
			name:      "approved status",
			planBody:  "**Plan status:** approved\n",
			wantState: planStateApproved,
		},
		{
			name:      "draft status",
			planBody:  "**Plan status:** draft\n",
			wantState: planStateDraft,
		},
		{
			name:      "pending marker",
			planBody:  "_(pending)_\n",
			wantState: planStateNone,
		},
		{
			name:      "empty plan body",
			planBody:  "",
			wantState: planStateNone,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			path := writePlanFile(t, tc.planBody)
			info := parsePlan(path)
			if info.state != tc.wantState {
				t.Errorf("state: got %v, want %v", info.state, tc.wantState)
			}
		})
	}
}

func TestParsePlan_NonExistentFile(t *testing.T) {
	info := parsePlan("/nonexistent/path/task.md")
	if info.state != planStateNone {
		t.Errorf("expected planStateNone for missing file, got %v", info.state)
	}
}

// ---------------------------------------------------------------------------
// Gate exclusion
// ---------------------------------------------------------------------------

func TestParsePlan_GateExclusion(t *testing.T) {
	// A plan with work steps AND all known gate patterns. Gates must not be
	// counted as steps.
	planBody := `**Plan status:** approved

#### Phase 1 — setup

- [x] Real work step one
- [x] Real work step two
- [ ] Phase 1 verification
- [ ] Phase 1 review
- [x] **Phase 1 verification.**
- [x] **Phase 1 review.**
- [ ] Real work step three
- [x] **STATE.md and skillbook update** per standard
- [x] **Task file completion** per guidelines
- [x] **Follow-up note** something
`
	path := writePlanFile(t, planBody)
	info := parsePlan(path)

	if info.state != planStateApproved {
		t.Errorf("state: got %v, want planStateApproved", info.state)
	}
	// 3 real work steps, 2 done (the third is unchecked).
	if info.totalSteps != 3 {
		t.Errorf("totalSteps: got %d, want 3 (gate lines must not be counted)", info.totalSteps)
	}
	if info.doneSteps != 2 {
		t.Errorf("doneSteps: got %d, want 2", info.doneSteps)
	}
}

// TestParsePlan_GateExclusion_LongNoteVariant verifies that real-file gate lines
// with trailing notes (e.g. "**Phase 1 verification.** — crafter-verifier 9/9 …")
// are still excluded from the step count. (#6)
func TestParsePlan_GateExclusion_LongNoteVariant(t *testing.T) {
	planBody := `**Plan status:** approved

#### Phase 1 — setup

- [x] Real work step one
- [x] **Phase 1 verification.** — crafter-verifier 9/9 PASS; all checks green.
- [x] **Phase 1 review.** — crafter-reviewer no findings; scorecard PASS.
- [ ] Real work step two
`
	path := writePlanFile(t, planBody)
	info := parsePlan(path)

	if info.totalSteps != 2 {
		t.Errorf("totalSteps: got %d, want 2 (long-note gate lines must not be counted)", info.totalSteps)
	}
	if info.doneSteps != 1 {
		t.Errorf("doneSteps: got %d, want 1", info.doneSteps)
	}
}

// TestParsePlan_GateExclusion_WorkStepWithReviewWord verifies that a genuine
// work step whose body merely contains the word "review" or "verification" amid
// other words is counted as a step and NOT treated as a gate. (#6)
func TestParsePlan_GateExclusion_WorkStepWithReviewWord(t *testing.T) {
	planBody := `**Plan status:** approved

#### Phase 1 — setup

- [ ] Review the architecture and document findings
- [ ] Perform a verification pass on the config
- [ ] Run the rollout review checklist
`
	path := writePlanFile(t, planBody)
	info := parsePlan(path)

	// All three are genuine work steps — none start with "Phase" so none match gate regex.
	if info.totalSteps != 3 {
		t.Errorf("totalSteps: got %d, want 3 (work steps with 'review'/'verification' words must be counted)", info.totalSteps)
	}
	if info.doneSteps != 0 {
		t.Errorf("doneSteps: got %d, want 0", info.doneSteps)
	}
}

// TestRenderExecuting_MalformedPlan_CurrentPhaseZero verifies the fix #4 guard:
// when currentPhase is 0 (malformed plan — work-step before first Phase heading),
// the renderer degrades silently by omitting the Phase X/Y prefix rather than
// emitting "Phase 0/N". (#4)
func TestRenderExecuting_MalformedPlan_CurrentPhaseZero(t *testing.T) {
	info := planInfo{
		state:        planStateApproved,
		totalPhases:  2,
		currentPhase: 0, // malformed: no phase heading seen before the first unchecked step
		doneSteps:    1,
		totalSteps:   3,
	}
	got := renderSegment(info)

	// Must not contain "Phase 0" or "Phase 0/".
	if strings.Contains(got, "Phase 0") {
		t.Errorf("malformed-plan segment must not contain 'Phase 0', got %q", got)
	}
	// Must start with "crafter".
	if !strings.HasPrefix(got, "crafter") {
		t.Errorf("expected segment to start with 'crafter', got %q", got)
	}
	// Step count + bar must still be present (totalSteps > 0).
	if !strings.Contains(got, "[") {
		t.Errorf("expected progress bar in segment, got %q", got)
	}
}

// TestParsePlan_MalformedPlan_StepBeforePhaseHeading exercises parsePlan itself
// for a plan where a work-step checkbox appears before any Phase heading, which
// leaves currentPhase as 0. The renderSegment output must not contain "Phase 0".
func TestParsePlan_MalformedPlan_StepBeforePhaseHeading(t *testing.T) {
	planBody := `**Plan status:** approved

- [ ] Step before any phase heading

#### Phase 1 — setup
- [x] Step inside phase 1
`
	path := writePlanFile(t, planBody)
	info := parsePlan(path)

	// The pre-heading step is unchecked → currentPhase stays 0.
	if info.currentPhase != 0 {
		t.Errorf("currentPhase: got %d, want 0 (step before first Phase heading)", info.currentPhase)
	}

	got := renderSegment(info)
	if strings.Contains(got, "Phase 0") {
		t.Errorf("segment must not contain 'Phase 0', got %q", got)
	}
	if !strings.HasPrefix(got, "crafter") {
		t.Errorf("expected segment to start with 'crafter', got %q", got)
	}
}

// ---------------------------------------------------------------------------
// Multi-phase detection
// ---------------------------------------------------------------------------

func TestParsePlan_MultiPhase_H4(t *testing.T) {
	// Three H4 phases; first unchecked step is in phase 2.
	planBody := `**Plan status:** approved

#### Phase 1 — first
- [x] Step 1a
- [x] Step 1b

#### Phase 2 — second
- [x] Step 2a
- [ ] Step 2b
- [ ] Step 2c

#### Phase 3 — third
- [ ] Step 3a
`
	path := writePlanFile(t, planBody)
	info := parsePlan(path)

	if info.totalPhases != 3 {
		t.Errorf("totalPhases: got %d, want 3", info.totalPhases)
	}
	if info.currentPhase != 2 {
		t.Errorf("currentPhase: got %d, want 2", info.currentPhase)
	}
}

func TestParsePlan_MultiPhase_H3(t *testing.T) {
	// Three H3 phases (A3/R1 variant): phase headings use ### not ####.
	planBody := `**Plan status:** approved

### Phase 1 — first
- [x] Step 1a

### Phase 2 — second
- [ ] Step 2a

### Phase 3 — third
- [ ] Step 3a
`
	path := writePlanFile(t, planBody)
	info := parsePlan(path)

	if info.totalPhases != 3 {
		t.Errorf("totalPhases: got %d, want 3 (H3 headings must be detected)", info.totalPhases)
	}
	if info.currentPhase != 2 {
		t.Errorf("currentPhase: got %d, want 2", info.currentPhase)
	}
}

func TestParsePlan_AllStepsDone_CurrentPhaseIsLast(t *testing.T) {
	planBody := `**Plan status:** approved

#### Phase 1 — first
- [x] Step 1a

#### Phase 2 — second
- [x] Step 2a
- [x] Step 2b
`
	path := writePlanFile(t, planBody)
	info := parsePlan(path)

	if info.currentPhase != 2 {
		t.Errorf("currentPhase: got %d, want 2 (last phase when all done)", info.currentPhase)
	}
	if info.doneSteps != info.totalSteps {
		t.Errorf("all steps should be done: done=%d total=%d", info.doneSteps, info.totalSteps)
	}
}

// ---------------------------------------------------------------------------
// Divide-by-zero / no work steps
// ---------------------------------------------------------------------------

func TestParsePlan_PhasesNoSteps(t *testing.T) {
	// Plan has phase headings but zero work steps → must not panic; renderSegment
	// must emit the phase info without a bar (per production code: step segment
	// is only rendered when totalSteps > 0).
	planBody := `**Plan status:** approved

#### Phase 1 — setup

#### Phase 2 — main
`
	path := writePlanFile(t, planBody)
	info := parsePlan(path)

	if info.totalSteps != 0 {
		t.Errorf("totalSteps: got %d, want 0", info.totalSteps)
	}
	// Should not panic. renderSegment must produce something sensible.
	got := renderSegment(info)
	if !strings.HasPrefix(got, "crafter") {
		t.Errorf("expected segment to start with 'crafter', got %q", got)
	}
	// No bar or percent should appear since there are no steps.
	if strings.Contains(got, "[") {
		t.Errorf("expected no bar when totalSteps=0, got %q", got)
	}
	if strings.Contains(got, "%") {
		t.Errorf("expected no percent when totalSteps=0, got %q", got)
	}
}

// ---------------------------------------------------------------------------
// resolveActiveTask edge cases (hermetic, using t.TempDir)
// ---------------------------------------------------------------------------

// makeRepo writes a minimal .git/HEAD file under root, recording the given branch.
func makeRepo(t *testing.T, root, branch string) {
	t.Helper()
	headContent := "ref: refs/heads/" + branch + "\n"
	writeFile(t, filepath.Join(root, ".git", "HEAD"), headContent)
}

// makeTaskFile writes a minimal task file with Status and Work branch metadata
// under <root>/.crafter/tasks/<name>.md.
func makeTaskFile(t *testing.T, root, name, status, branch string) string {
	t.Helper()
	content := "## Metadata\n- **Status:** " + status + "\n- **Work branch:** " + branch + "\n"
	path := filepath.Join(root, ".crafter", "tasks", name)
	writeFile(t, path, content)
	return path
}

func TestResolveActiveTask_Match(t *testing.T) {
	root := t.TempDir()
	makeRepo(t, root, "feat/my-feature")
	makeTaskFile(t, root, "20260601-my-task.md", "active", "feat/my-feature")

	m := resolveActiveTask(root)
	if m == nil {
		t.Fatal("expected a match, got nil")
	}
	if !strings.HasSuffix(m.Path, "20260601-my-task.md") {
		t.Errorf("unexpected match path: %s", m.Path)
	}
}

func TestResolveActiveTask_NoCrafterDir(t *testing.T) {
	root := t.TempDir()
	makeRepo(t, root, "feat/my-feature")
	// No .crafter/ directory created.

	m := resolveActiveTask(root)
	if m != nil {
		t.Errorf("expected nil when no .crafter/ dir, got %v", m)
	}
}

func TestResolveActiveTask_BranchMismatch(t *testing.T) {
	root := t.TempDir()
	makeRepo(t, root, "main")
	makeTaskFile(t, root, "20260601-my-task.md", "active", "feat/other-branch")

	m := resolveActiveTask(root)
	if m != nil {
		t.Errorf("expected nil when work branch doesn't match current branch, got %v", m)
	}
}

func TestResolveActiveTask_StatusNotActive(t *testing.T) {
	root := t.TempDir()
	makeRepo(t, root, "feat/my-feature")
	makeTaskFile(t, root, "20260601-my-task.md", "done", "feat/my-feature")

	m := resolveActiveTask(root)
	if m != nil {
		t.Errorf("expected nil for non-active task, got %v", m)
	}
}

func TestResolveActiveTask_MultipleMatchesMostRecentWins(t *testing.T) {
	root := t.TempDir()
	makeRepo(t, root, "feat/my-feature")
	// Two active files matching the same branch; lexicographically larger (later date) wins.
	makeTaskFile(t, root, "20260501-older-task.md", "active", "feat/my-feature")
	makeTaskFile(t, root, "20260601-newer-task.md", "active", "feat/my-feature")

	m := resolveActiveTask(root)
	if m == nil {
		t.Fatal("expected a match, got nil")
	}
	if !strings.HasSuffix(m.Path, "20260601-newer-task.md") {
		t.Errorf("expected newer task to win tie-break, got %s", m.Path)
	}
}

func TestResolveActiveTask_DetachedHead(t *testing.T) {
	root := t.TempDir()
	// Detached HEAD: no "ref: refs/heads/" prefix.
	writeFile(t, filepath.Join(root, ".git", "HEAD"), "abc1234def5678\n")
	makeTaskFile(t, root, "20260601-my-task.md", "active", "feat/my-feature")

	m := resolveActiveTask(root)
	if m != nil {
		t.Errorf("expected nil for detached HEAD, got %v", m)
	}
}

func TestResolveActiveTask_NoGitRepo(t *testing.T) {
	root := t.TempDir()
	// Create .crafter dir but no .git.
	if err := os.MkdirAll(filepath.Join(root, ".crafter", "tasks"), 0o755); err != nil {
		t.Fatalf("creating .crafter/tasks: %v", err)
	}

	m := resolveActiveTask(root)
	if m != nil {
		t.Errorf("expected nil when no git repo, got %v", m)
	}
}

// ---------------------------------------------------------------------------
// Render integration over temp dirs
// ---------------------------------------------------------------------------

func TestRender_EmptyDir(t *testing.T) {
	root := t.TempDir()
	// No .crafter/ directory → should return "".
	got := Render(root)
	if got != "" {
		t.Errorf("expected empty string for non-Crafter workdir, got %q", got)
	}
}

func TestRender_ActiveApprovedTask(t *testing.T) {
	root := t.TempDir()
	makeRepo(t, root, "feat/my-feature")

	// Write a task file with Status+branch metadata and a simple approved plan.
	taskContent := `## Metadata
- **Status:** active
- **Work branch:** feat/my-feature

## Plan
**Plan status:** approved

#### Phase 1 — setup
- [x] Step one
- [ ] Step two
`
	taskPath := filepath.Join(root, ".crafter", "tasks", "20260601-my-task.md")
	writeFile(t, taskPath, taskContent)

	got := Render(root)
	if got == "" {
		t.Error("expected non-empty segment for active approved task, got empty string")
	}
	if !strings.HasPrefix(got, "crafter") {
		t.Errorf("expected segment to start with 'crafter', got %q", got)
	}
	// Verify executing state elements are present.
	if !strings.Contains(got, "Phase") {
		t.Errorf("expected 'Phase' in segment, got %q", got)
	}
	if !strings.Contains(got, "[") {
		t.Errorf("expected progress bar in segment, got %q", got)
	}
}

func TestRender_ActiveDraftTask(t *testing.T) {
	root := t.TempDir()
	makeRepo(t, root, "feat/draft-branch")

	taskContent := `## Metadata
- **Status:** active
- **Work branch:** feat/draft-branch

## Plan
**Plan status:** draft
`
	taskPath := filepath.Join(root, ".crafter", "tasks", "20260601-draft-task.md")
	writeFile(t, taskPath, taskContent)

	got := Render(root)
	if got != "crafter · plan: awaiting approval" {
		t.Errorf("got %q, want %q", got, "crafter · plan: awaiting approval")
	}
}

// ---------------------------------------------------------------------------
// Live-fixture smoke test
// ---------------------------------------------------------------------------

func TestParsePlan_LiveFixture(t *testing.T) {
	// Find the repo root by walking up from the test file's directory.
	// The repo is expected to have .crafter/tasks/ with at least one .md file.
	// If the fixture directory doesn't exist, skip gracefully.
	repoRoot := findRepoRoot(t)
	if repoRoot == "" {
		t.Skip("repository root not found; skipping live-fixture smoke test")
	}

	tasksDir := filepath.Join(repoRoot, ".crafter", "tasks")
	entries, err := os.ReadDir(tasksDir)
	if err != nil {
		t.Skipf(".crafter/tasks/ not found or unreadable (%v); skipping live-fixture smoke test", err)
	}

	var mdFiles []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
			mdFiles = append(mdFiles, filepath.Join(tasksDir, e.Name()))
		}
	}
	if len(mdFiles) == 0 {
		t.Skip("no .md files in .crafter/tasks/; skipping live-fixture smoke test")
	}

	// Use the lexicographically last file (most recent by date prefix).
	latestFile := mdFiles[len(mdFiles)-1]

	info := parsePlan(latestFile)
	seg := renderSegment(info)

	// The segment must either be empty (state-independent) or start with "crafter".
	// We only assert it starts with "crafter" when it is non-empty.
	if seg != "" && !strings.HasPrefix(seg, "crafter") {
		t.Errorf("live-fixture segment does not start with 'crafter': %q", seg)
	}

	// For the known active statusline task file, we expect an executing state.
	if strings.HasSuffix(latestFile, "20260601-plan-progress-statusline.md") {
		if seg == "" {
			t.Errorf("expected non-empty segment for known active task file %s", latestFile)
		}
		// It has an approved plan; should contain "Phase".
		if !strings.Contains(seg, "Phase") {
			t.Errorf("expected 'Phase' in segment for approved plan, got %q", seg)
		}
	}
}

// findRepoRoot walks up from the process working directory until it finds
// a directory containing both ".git" and ".crafter". Returns "" if not found.
func findRepoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	for {
		hasCrafter := false
		hasGit := false
		entries, err := os.ReadDir(dir)
		if err != nil {
			break
		}
		for _, e := range entries {
			if e.Name() == ".crafter" && e.IsDir() {
				hasCrafter = true
			}
			if e.Name() == ".git" {
				hasGit = true
			}
		}
		if hasCrafter && hasGit {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}
