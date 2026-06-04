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
			want: "plan: awaiting approval",
		},
		{
			name: "none (pending) → planning",
			info: planInfo{state: planStateNone},
			want: "planning",
		},
		{
			name: "approved no phases no steps → empty string",
			info: planInfo{state: planStateApproved},
			want: "",
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
			want: "Phase 2/3 · 7/12 [█████░░░░░] 58%",
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
			want: "Phase 2/2 · 4/4 [██████████] 100%",
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
			want: "Phase 1/2 · 0/5 [░░░░░░░░░░] 0%",
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
			wantSegment: "Phase 1/1 · 7/12 [█████░░░░░] 58%",
		},
		{
			name:        "1/3 → 33% 3 filled",
			done:        1,
			total:       3,
			wantPct:     33,
			wantFilled:  3,
			wantSegment: "Phase 1/1 · 1/3 [███░░░░░░░] 33%",
		},
		{
			name:        "all done → 100% 10 filled",
			done:        3,
			total:       3,
			wantPct:     100,
			wantFilled:  10,
			wantSegment: "Phase 1/1 · 3/3 [██████████] 100%",
		},
		{
			name:        "0 done → 0% 0 filled",
			done:        0,
			total:       6,
			wantPct:     0,
			wantFilled:  0,
			wantSegment: "Phase 1/1 · 0/6 [░░░░░░░░░░] 0%",
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
	// Step count + bar must still be present (totalSteps > 0).
	if !strings.Contains(got, "[") {
		t.Errorf("expected progress bar in segment, got %q", got)
	}
	// hasPhase gate: with no phase segment preceding it, the step/bar segment
	// must NOT be prefixed by the " · " separator. This pins the renderExecuting
	// hasPhase flag against regression (a leading separator would appear if the
	// separator were emitted unconditionally).
	if strings.HasPrefix(got, " · ") {
		t.Errorf("phaseless segment must not start with a leading separator, got %q", got)
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
	// Should not panic. renderSegment must produce something sensible (no bar/percent).
	got := renderSegment(info)
	// No bar or percent should appear since there are no steps.
	if strings.Contains(got, "[") {
		t.Errorf("expected no bar when totalSteps=0, got %q", got)
	}
	if strings.Contains(got, "%") {
		t.Errorf("expected no percent when totalSteps=0, got %q", got)
	}
}

// ---------------------------------------------------------------------------
// classifyTasks edge cases (hermetic, using t.TempDir)
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

// classifyFromRoot is a convenience helper that resolves the crafter context
// directory and git branch from root (identical to what Render does), then
// calls classifyTasks.  Returns a zero taskClassification if either lookup
// fails, mirroring the silent-fail behaviour of the production path.
func classifyFromRoot(root string) taskClassification {
	ctxDir := findCrafterDir(root)
	if ctxDir == "" {
		return taskClassification{}
	}
	branch := readGitBranch(root)
	if branch == "" {
		return taskClassification{}
	}
	return classifyTasks(ctxDir, branch)
}

// TestClassifyTasks_Match verifies that an active task whose work branch
// matches the current branch is selected as ActiveCurrent.
func TestClassifyTasks_Match(t *testing.T) {
	root := t.TempDir()
	makeRepo(t, root, "feat/my-feature")
	makeTaskFile(t, root, "20260601-my-task.md", "active", "feat/my-feature")

	cls := classifyFromRoot(root)
	if !strings.HasSuffix(cls.ActiveCurrent, "20260601-my-task.md") {
		t.Errorf("unexpected ActiveCurrent path: %s", cls.ActiveCurrent)
	}
}

// TestClassifyTasks_NoCrafterDir verifies that a workdir with no .crafter/
// directory produces a zero classification (ActiveCurrent == "").
func TestClassifyTasks_NoCrafterDir(t *testing.T) {
	root := t.TempDir()
	makeRepo(t, root, "feat/my-feature")
	// No .crafter/ directory created.

	cls := classifyFromRoot(root)
	if cls.ActiveCurrent != "" {
		t.Errorf("expected empty ActiveCurrent when no .crafter/ dir, got %q", cls.ActiveCurrent)
	}
}

// TestClassifyTasks_BranchMismatch verifies that a task whose work branch does
// not match the current branch is NOT selected as ActiveCurrent.
func TestClassifyTasks_BranchMismatch(t *testing.T) {
	root := t.TempDir()
	makeRepo(t, root, "main")
	makeTaskFile(t, root, "20260601-my-task.md", "active", "feat/other-branch")

	cls := classifyFromRoot(root)
	if cls.ActiveCurrent != "" {
		t.Errorf("expected empty ActiveCurrent when work branch doesn't match, got %q", cls.ActiveCurrent)
	}
}

// TestClassifyTasks_StatusNotActive verifies that a task with a non-active
// status is not selected as ActiveCurrent.
func TestClassifyTasks_StatusNotActive(t *testing.T) {
	root := t.TempDir()
	makeRepo(t, root, "feat/my-feature")
	makeTaskFile(t, root, "20260601-my-task.md", "done", "feat/my-feature")

	cls := classifyFromRoot(root)
	if cls.ActiveCurrent != "" {
		t.Errorf("expected empty ActiveCurrent for non-active task, got %q", cls.ActiveCurrent)
	}
}

// TestClassifyTasks_MultipleMatchesMostRecentWins verifies that when multiple
// active tasks match the current branch, the lexicographically-largest filename
// (most-recent date prefix) is selected as ActiveCurrent.
func TestClassifyTasks_MultipleMatchesMostRecentWins(t *testing.T) {
	root := t.TempDir()
	makeRepo(t, root, "feat/my-feature")
	// Two active files matching the same branch; lexicographically larger (later date) wins.
	makeTaskFile(t, root, "20260501-older-task.md", "active", "feat/my-feature")
	makeTaskFile(t, root, "20260601-newer-task.md", "active", "feat/my-feature")

	cls := classifyFromRoot(root)
	if !strings.HasSuffix(cls.ActiveCurrent, "20260601-newer-task.md") {
		t.Errorf("expected newer task to win tie-break, got %s", cls.ActiveCurrent)
	}
}

// TestClassifyTasks_DetachedHead verifies that a detached HEAD (bare SHA in
// .git/HEAD) produces a zero classification (ActiveCurrent == "").
func TestClassifyTasks_DetachedHead(t *testing.T) {
	root := t.TempDir()
	// Detached HEAD: no "ref: refs/heads/" prefix.
	writeFile(t, filepath.Join(root, ".git", "HEAD"), "abc1234def5678\n")
	makeTaskFile(t, root, "20260601-my-task.md", "active", "feat/my-feature")

	cls := classifyFromRoot(root)
	if cls.ActiveCurrent != "" {
		t.Errorf("expected empty ActiveCurrent for detached HEAD, got %q", cls.ActiveCurrent)
	}
}

// TestClassifyTasks_NoGitRepo verifies that a workdir with a .crafter/
// directory but no .git repo produces a zero classification (ActiveCurrent == "").
func TestClassifyTasks_NoGitRepo(t *testing.T) {
	root := t.TempDir()
	// Create .crafter dir but no .git.
	if err := os.MkdirAll(filepath.Join(root, ".crafter", "tasks"), 0o755); err != nil {
		t.Fatalf("creating .crafter/tasks: %v", err)
	}

	cls := classifyFromRoot(root)
	if cls.ActiveCurrent != "" {
		t.Errorf("expected empty ActiveCurrent when no git repo, got %q", cls.ActiveCurrent)
	}
}

// TestClassifyTasks_CompletedTieBreak directly asserts that when multiple
// completed tasks match the current branch, CompletedCurrent is set to the
// lexicographically-largest (most-recent-by-filename) path.
func TestClassifyTasks_CompletedTieBreak(t *testing.T) {
	root := t.TempDir()
	makeRepo(t, root, "feat/done-branch")
	makeTaskFile(t, root, "20260401-older-done.md", "completed", "feat/done-branch")
	makeTaskFile(t, root, "20260601-newer-done.md", "completed", "feat/done-branch")

	ctxDir := findCrafterDir(root)
	branch := readGitBranch(root)
	cls := classifyTasks(ctxDir, branch)

	if !strings.HasSuffix(cls.CompletedCurrent, "20260601-newer-done.md") {
		t.Errorf("completed tie-break: expected newer file as CompletedCurrent, got %q", cls.CompletedCurrent)
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
	want := "plan: awaiting approval │ ⎇ feat/draft-branch"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
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

	// Prefix-free invariant: whatever state the live fixture is in, a non-empty
	// rendered segment must never begin with a stray separator nor leak the
	// degraded "Phase 0" indicator. (Empty segments are valid for unknown states.)
	if seg != "" {
		if strings.HasPrefix(seg, " · ") {
			t.Errorf("live-fixture segment must not start with a leading separator, got %q", seg)
		}
		if strings.Contains(seg, "Phase 0") {
			t.Errorf("live-fixture segment must not contain 'Phase 0', got %q", seg)
		}
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

// ---------------------------------------------------------------------------
// Render cascade — rung 2, 3, 4 and edge cases
// ---------------------------------------------------------------------------

// TestRender_Rung2_CompletedCurrentBranch verifies that a completed task on
// the current branch (no active task) produces the "✓ done" segment.
func TestRender_Rung2_CompletedCurrentBranch(t *testing.T) {
	root := t.TempDir()
	makeRepo(t, root, "feat/done-branch")
	makeTaskFile(t, root, "20260601-finished.md", "completed", "feat/done-branch")

	got := Render(root)
	want := segDone + " │ ⎇ feat/done-branch"
	if got != want {
		t.Errorf("rung 2: got %q, want %q", got, want)
	}
}

// TestRender_Rung2_TieBreakMostRecentWins verifies that when multiple completed
// tasks on the current branch exist, the lexicographically-largest filename wins.
func TestRender_Rung2_TieBreakMostRecentWins(t *testing.T) {
	root := t.TempDir()
	makeRepo(t, root, "feat/done-branch")
	makeTaskFile(t, root, "20260401-older-done.md", "completed", "feat/done-branch")
	makeTaskFile(t, root, "20260601-newer-done.md", "completed", "feat/done-branch")

	got := Render(root)
	// Both completed tasks match; the lexicographically-larger (newer) one wins.
	// The result is still segDone regardless of which file was selected, but the
	// rung-2 path must be taken (not rung 3 or 4).
	want := segDone + " │ ⎇ feat/done-branch"
	if got != want {
		t.Errorf("rung 2 tie-break: got %q, want %q", got, want)
	}
}

// TestRender_Rung1_BeatsRung2 verifies that when both an active and a completed
// task exist on the current branch, rung 1 (active) wins over rung 2 (completed).
func TestRender_Rung1_BeatsRung2(t *testing.T) {
	root := t.TempDir()
	makeRepo(t, root, "feat/mixed-branch")

	// Write a completed task.
	makeTaskFile(t, root, "20260501-done-task.md", "completed", "feat/mixed-branch")

	// Write a full active task with an approved plan so rung 1 produces a non-empty
	// segment distinct from segDone.
	activeContent := `## Metadata
- **Status:** active
- **Work branch:** feat/mixed-branch

## Plan
**Plan status:** approved

#### Phase 1 — work
- [x] Step one
- [ ] Step two
`
	activeTaskPath := filepath.Join(root, ".crafter", "tasks", "20260601-active-task.md")
	writeFile(t, activeTaskPath, activeContent)

	got := Render(root)
	// Rung 1 must win: result must NOT be segDone and must contain plan-progress markers.
	if got == segDone {
		t.Errorf("rung 1 must beat rung 2: got segDone %q but expected active plan segment", got)
	}
	if got == "" {
		t.Error("rung 1 must beat rung 2: got empty string but expected active plan segment")
	}
	if !strings.Contains(got, "Phase") {
		t.Errorf("rung 1 must beat rung 2: expected 'Phase' in segment, got %q", got)
	}
}

// TestRender_Rung3_Singular verifies that one active task on a different branch
// produces "1 active elsewhere".
func TestRender_Rung3_Singular(t *testing.T) {
	root := t.TempDir()
	makeRepo(t, root, "main")
	makeTaskFile(t, root, "20260601-other-task.md", "active", "feat/other-branch")

	got := Render(root)
	want := "1 active elsewhere │ ⎇ main"
	if got != want {
		t.Errorf("rung 3 singular: got %q, want %q", got, want)
	}
}

// TestRender_Rung3_Plural verifies that multiple active tasks on other branches
// produce "N active elsewhere" with the correct count.
func TestRender_Rung3_Plural(t *testing.T) {
	root := t.TempDir()
	makeRepo(t, root, "main")
	makeTaskFile(t, root, "20260601-task-a.md", "active", "feat/branch-a")
	makeTaskFile(t, root, "20260602-task-b.md", "active", "feat/branch-b")
	makeTaskFile(t, root, "20260603-task-c.md", "active", "feat/branch-c")

	got := Render(root)
	want := "3 active elsewhere │ ⎇ main"
	if got != want {
		t.Errorf("rung 3 plural: got %q, want %q", got, want)
	}
}

// TestRender_Rung4_ZeroActiveOther verifies that when there are no active tasks
// anywhere, Render returns "" (not "0 active elsewhere").
func TestRender_Rung4_Zero(t *testing.T) {
	root := t.TempDir()
	makeRepo(t, root, "main")
	// A completed task on a different branch — no active tasks anywhere.
	makeTaskFile(t, root, "20260601-old-task.md", "completed", "feat/other-branch")

	got := Render(root)
	// Plan section is empty (rung 4), so the panel is just the vcs branch token.
	// It must NOT be "0 active elsewhere".
	want := "⎇ main"
	if got != want {
		t.Errorf("rung 4: expected %q, got %q (must not be '0 active elsewhere')", want, got)
	}
}

// TestRender_Guard_NoCrafterDir verifies that a workdir with no .crafter/ directory
// returns "".
func TestRender_Guard_NoCrafterDir(t *testing.T) {
	root := t.TempDir()
	makeRepo(t, root, "main")
	// No .crafter/ directory created.

	got := Render(root)
	// No .crafter → plan section suppressed; the panel still renders the vcs
	// branch token (the guard only suppresses the plan section, not the panel).
	want := "⎇ main"
	if got != want {
		t.Errorf("no .crafter dir: expected %q, got %q", want, got)
	}
}

// TestRender_Guard_DetachedHead verifies that a detached HEAD returns "".
func TestRender_Guard_DetachedHead(t *testing.T) {
	root := t.TempDir()
	// Detached HEAD: write a bare SHA instead of "ref: refs/heads/…".
	writeFile(t, filepath.Join(root, ".git", "HEAD"), "abc1234def5678\n")
	makeTaskFile(t, root, "20260601-task.md", "active", "feat/some-branch")

	got := Render(root)
	if got != "" {
		t.Errorf("detached HEAD: expected %q, got %q", "", got)
	}
}

// TestRender_Guard_TasksDirMissingOrEmpty verifies that a .crafter/ with no tasks
// returns "".
func TestRender_Guard_TasksDirMissingOrEmpty(t *testing.T) {
	root := t.TempDir()
	makeRepo(t, root, "main")
	// Create .crafter/ but no tasks/ subdirectory.
	if err := os.MkdirAll(filepath.Join(root, ".crafter"), 0o755); err != nil {
		t.Fatalf("creating .crafter: %v", err)
	}

	got := Render(root)
	// Tasks dir missing → plan section suppressed; the panel still renders the
	// vcs branch token.
	want := "⎇ main"
	if got != want {
		t.Errorf("tasks dir missing: expected %q, got %q", want, got)
	}
}

// TestRender_KnownLimitation_NonStandardBranchField verifies the settled scope
// decision: a task file that uses only the non-standard "- **Branch:** <value>"
// field (without "- **Work branch:** ") is NOT counted by rung 3.  This pins the
// strict-match behaviour documented in the task spec as a known limitation.
func TestRender_KnownLimitation_NonStandardBranchField(t *testing.T) {
	root := t.TempDir()
	makeRepo(t, root, "main")

	// Write the task file directly (not via makeTaskFile, which emits the standard
	// "- **Work branch:** " field) to reproduce the non-standard alias pattern.
	nonStandardContent := "## Metadata\n- **Status:** active\n- **Branch:** feat/other-branch\n"
	writeFile(t, filepath.Join(root, ".crafter", "tasks", "20260421-nonstandard-branch.md"), nonStandardContent)

	// There are no tasks with the standard "- **Work branch:** " field, so the
	// active-other count must be 0 and Render must return "".
	got := Render(root)
	// Rung 3 must not count the non-standard field, so the plan section is empty;
	// the panel is just the vcs branch token (not "1 active elsewhere").
	want := "⎇ main"
	if got != want {
		t.Errorf("non-standard branch field: expected %q (count 0), got %q — non-standard '- **Branch:** ' must NOT be counted by rung 3", want, got)
	}
}

// TestRender_Rung2_BeatsRung3 verifies rung precedence: when both a completed
// task on the current branch AND active tasks on other branches exist, rung 2
// (completed-current) wins over rung 3 (active-elsewhere).
func TestRender_Rung2_BeatsRung3(t *testing.T) {
	root := t.TempDir()
	makeRepo(t, root, "feat/done-branch")
	// Completed task on current branch (rung 2 candidate).
	makeTaskFile(t, root, "20260601-finished.md", "completed", "feat/done-branch")
	// Active task on a different branch (rung 3 candidate).
	makeTaskFile(t, root, "20260602-other-active.md", "active", "feat/other-branch")

	got := Render(root)
	want := segDone + " │ ⎇ feat/done-branch"
	if got != want {
		t.Errorf("rung 2 beats rung 3: got %q, want %q", got, want)
	}
}

// ---------------------------------------------------------------------------
// modelSection — render and degradation
// ---------------------------------------------------------------------------

// TestModelSection covers the model-section string and every degradation path:
// effort present/absent, the k/M capacity abbreviation (1M / 200k / 128k /
// sub-1000), capacity 0 (token omitted), and empty display_name (whole section
// omitted).
func TestModelSection(t *testing.T) {
	tests := []struct {
		name string
		p    Payload
		want string
	}{
		{
			name: "full: display + 1M + effort",
			p:    Payload{ModelDisplayName: "Opus 4.8", ContextWindowSize: 1_000_000, EffortLevel: "high"},
			want: "Opus 4.8 1M (high)",
		},
		{
			name: "effort absent → no parens",
			p:    Payload{ModelDisplayName: "Opus 4.8", ContextWindowSize: 1_000_000},
			want: "Opus 4.8 1M",
		},
		{
			name: "200k capacity",
			p:    Payload{ModelDisplayName: "Opus 4.8", ContextWindowSize: 200_000, EffortLevel: "medium"},
			want: "Opus 4.8 200k (medium)",
		},
		{
			name: "128k capacity",
			p:    Payload{ModelDisplayName: "Sonnet 4.8", ContextWindowSize: 128_000},
			want: "Sonnet 4.8 128k",
		},
		{
			name: "sub-1000 capacity → raw integer",
			p:    Payload{ModelDisplayName: "Tiny", ContextWindowSize: 512},
			want: "Tiny 512",
		},
		{
			name: "capacity 0 → capacity token omitted, effort kept",
			p:    Payload{ModelDisplayName: "Opus 4.8", ContextWindowSize: 0, EffortLevel: "high"},
			want: "Opus 4.8 (high)",
		},
		{
			name: "capacity 0 and no effort → display only",
			p:    Payload{ModelDisplayName: "Opus 4.8"},
			want: "Opus 4.8",
		},
		{
			name: "empty display_name → whole section omitted",
			p:    Payload{ContextWindowSize: 1_000_000, EffortLevel: "high"},
			want: "",
		},
		// --- context-parenthetical stripping ---
		{
			// Headline fix: Claude Code embeds "(1M context)" in the display name;
			// crafter must strip it before appending its own "1M" token.
			name: "display_name with (1M context) parenthetical + capacity → stripped",
			p:    Payload{ModelDisplayName: "Opus 4.8 (1M context)", ContextWindowSize: 1_000_000, EffortLevel: "xhigh"},
			want: "Opus 4.8 1M (xhigh)",
		},
		{
			// Display name without a context parenthetical → no change in behaviour.
			name: "display_name without parenthetical + capacity → unchanged",
			p:    Payload{ModelDisplayName: "Opus 4.8", ContextWindowSize: 1_000_000, EffortLevel: "xhigh"},
			want: "Opus 4.8 1M (xhigh)",
		},
		{
			// Guard: when ContextWindowSize == 0, the parenthetical must NOT be
			// stripped so the display name is preserved verbatim.
			name: "display_name with (1M context) but capacity 0 → NOT stripped",
			p:    Payload{ModelDisplayName: "Opus 4.8 (1M context)", ContextWindowSize: 0, EffortLevel: "xhigh"},
			want: "Opus 4.8 (1M context) (xhigh)",
		},
		{
			// A trailing parenthetical that does NOT mention "context" must be
			// preserved — only context-bearing parentheticals are stripped.
			name: "non-context trailing parenthetical preserved when capacity present",
			p:    Payload{ModelDisplayName: "Sonnet 4.6 (beta)", ContextWindowSize: 200_000, EffortLevel: "low"},
			want: "Sonnet 4.6 (beta) 200k (low)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := modelSection(tt.p); got != tt.want {
				t.Errorf("modelSection() = %q, want %q", got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// costSection — render and degradation
// ---------------------------------------------------------------------------

// floatPtr returns a pointer to the given float64, for building cost payloads.
func floatPtr(v float64) *float64 { return &v }

// TestCostSection covers the cost-section string: a positive value renders as
// "$X.XX" (2 decimals); zero and absent (nil) both omit the section.
func TestCostSection(t *testing.T) {
	tests := []struct {
		name string
		p    Payload
		want string
	}{
		{
			name: "positive → $X.XX",
			p:    Payload{TotalCostUSD: floatPtr(0.42)},
			want: "$0.42",
		},
		{
			name: "positive rounds/pads to 2 decimals",
			p:    Payload{TotalCostUSD: floatPtr(1.5)},
			want: "$1.50",
		},
		{
			name: "zero (*0) → omitted",
			p:    Payload{TotalCostUSD: floatPtr(0)},
			want: "",
		},
		{
			name: "absent (nil) → omitted",
			p:    Payload{TotalCostUSD: nil},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := costSection(tt.p); got != tt.want {
				t.Errorf("costSection() = %q, want %q", got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// vcsSection — grouped project + branch + diff, ANSI colors, configurable icon
// ---------------------------------------------------------------------------

// TestVcsSection_Full verifies the full grouped string with the default icon and
// the raw escape sequences: dim project name, normal branch token, and the
// green/red diff suffix.
func TestVcsSection_Full(t *testing.T) {
	root := t.TempDir()
	makeRepo(t, root, "feat/plan-progress")

	p := Payload{
		Workdir:           root,
		ProjectDir:        "/some/path/crafter",
		TotalLinesAdded:   120,
		TotalLinesRemoved: 30,
	}

	want := "\033[2mcrafter\033[0m ⎇ feat/plan-progress " +
		"\033[32m+120\033[0m/\033[31m-30\033[0m"

	if got := vcsSection(p); got != want {
		t.Errorf("vcsSection() = %q, want %q", got, want)
	}
}

// TestVcsSection_CustomIconOverride verifies the branch icon is read from
// CRAFTER_STATUSLINE_BRANCH_ICON at render time, and that an empty override falls
// back to the default glyph.
func TestVcsSection_CustomIconOverride(t *testing.T) {
	root := t.TempDir()
	makeRepo(t, root, "main")

	p := Payload{Workdir: root, ProjectDir: "/x/crafter"}

	t.Run("custom icon", func(t *testing.T) {
		t.Setenv("CRAFTER_STATUSLINE_BRANCH_ICON", "★")
		want := "\033[2mcrafter\033[0m ★ main"
		if got := vcsSection(p); got != want {
			t.Errorf("vcsSection() = %q, want %q", got, want)
		}
	})

	t.Run("empty override falls back to default", func(t *testing.T) {
		t.Setenv("CRAFTER_STATUSLINE_BRANCH_ICON", "")
		want := "\033[2mcrafter\033[0m ⎇ main"
		if got := vcsSection(p); got != want {
			t.Errorf("vcsSection() = %q, want %q", got, want)
		}
	})
}

// TestVcsSection_Degradation covers the four intra-group degradation paths, each
// asserting there is no stray, leading, doubled, or trailing space.
func TestVcsSection_Degradation(t *testing.T) {
	t.Run("diff omitted when both line counts zero", func(t *testing.T) {
		root := t.TempDir()
		makeRepo(t, root, "main")
		p := Payload{Workdir: root, ProjectDir: "/x/crafter", TotalLinesAdded: 0, TotalLinesRemoved: 0}
		want := "\033[2mcrafter\033[0m ⎇ main"
		if got := vcsSection(p); got != want {
			t.Errorf("vcsSection() = %q, want %q", got, want)
		}
	})

	t.Run("project omitted when project_dir empty — no leading space", func(t *testing.T) {
		root := t.TempDir()
		makeRepo(t, root, "main")
		p := Payload{Workdir: root, ProjectDir: "", TotalLinesAdded: 5, TotalLinesRemoved: 2}
		want := "⎇ main \033[32m+5\033[0m/\033[31m-2\033[0m"
		got := vcsSection(p)
		if got != want {
			t.Errorf("vcsSection() = %q, want %q", got, want)
		}
		if strings.HasPrefix(got, " ") {
			t.Errorf("vcsSection() has a leading space: %q", got)
		}
	})

	t.Run("branch omitted when no git — only project renders", func(t *testing.T) {
		root := t.TempDir() // no .git
		p := Payload{Workdir: root, ProjectDir: "/x/crafter", TotalLinesAdded: 5, TotalLinesRemoved: 2}
		// Branch absent → branch token (and the diff that attaches to it) dropped.
		want := "\033[2mcrafter\033[0m"
		got := vcsSection(p)
		if got != want {
			t.Errorf("vcsSection() = %q, want %q", got, want)
		}
		if strings.HasSuffix(got, " ") {
			t.Errorf("vcsSection() has a trailing space: %q", got)
		}
	})

	t.Run("whole group omitted when project and branch both absent", func(t *testing.T) {
		root := t.TempDir() // no .git
		p := Payload{Workdir: root, ProjectDir: "", TotalLinesAdded: 5, TotalLinesRemoved: 2}
		if got := vcsSection(p); got != "" {
			t.Errorf("vcsSection() = %q, want empty string", got)
		}
	})
}

// ---------------------------------------------------------------------------
// ctxSection — context-window bar, reuses the plan bar; null → omitted
// ---------------------------------------------------------------------------

// TestCtxSection covers the ctx-section string at several percentages (0/42/100),
// a fractional value that rounds, and the null omission. The bar glyphs and
// "[bar] N%" format mirror the plan bar exactly.
func TestCtxSection(t *testing.T) {
	tests := []struct {
		name string
		p    Payload
		want string
	}{
		{
			name: "0% → empty bar",
			p:    Payload{UsedPercentage: floatPtr(0)},
			want: "[░░░░░░░░░░] 0%",
		},
		{
			name: "42% → 4 filled",
			p:    Payload{UsedPercentage: floatPtr(42)},
			want: "[████░░░░░░] 42%",
		},
		{
			name: "100% → full bar",
			p:    Payload{UsedPercentage: floatPtr(100)},
			want: "[██████████] 100%",
		},
		{
			name: "fractional rounds to nearest integer (42.5 → 43)",
			p:    Payload{UsedPercentage: floatPtr(42.5)},
			want: "[████░░░░░░] 43%",
		},
		{
			name: "null (nil) → omitted",
			p:    Payload{UsedPercentage: nil},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ctxSection(tt.p); got != tt.want {
				t.Errorf("ctxSection() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestRenderBar_PlanBarUnchanged pins that the shared renderBar helper still
// produces the exact bracketed bars the plan bar embedded before the extraction,
// proving the plan bar's output is byte-unchanged. The existing plan-section
// tests (TestRenderSegment_States, TestPercentAndBar) also assert the full
// "Phase X/Y · D/T [bar] N%" strings unchanged, which is the primary evidence.
func TestRenderBar_PlanBarUnchanged(t *testing.T) {
	tests := []struct {
		pct  int
		want string
	}{
		{0, "[░░░░░░░░░░]"},
		{33, "[███░░░░░░░]"},
		{58, "[█████░░░░░]"},
		{100, "[██████████]"},
	}
	for _, tt := range tests {
		if got := renderBar(tt.pct); got != tt.want {
			t.Errorf("renderBar(%d) = %q, want %q", tt.pct, got, tt.want)
		}
	}
}

// ---------------------------------------------------------------------------
// RenderPanel — full panel assembly degradation matrix
// ---------------------------------------------------------------------------

// makePlanRepo sets up a workspace root with a .git/HEAD on the given branch and
// a single active+approved task whose plan renders the fixed segment
// "Phase 1/1 · 1/2 [█████░░░░░] 50%". It is used by the panel-assembly matrix
// rows that need a non-empty plan section.
func makePlanRepo(t *testing.T, root, branch string) {
	t.Helper()
	makeRepo(t, root, branch)
	taskContent := `## Metadata
- **Status:** active
- **Work branch:** ` + branch + `

## Plan
**Plan status:** approved

#### Phase 1 — setup
- [x] Step one
- [ ] Step two
`
	writeFile(t, filepath.Join(root, ".crafter", "tasks", "20260601-panel-task.md"), taskContent)
}

// planSeg is the fixed plan segment produced by makePlanRepo (1/2 steps = 50%).
const planSeg = "Phase 1/1 · 1/2 [█████░░░░░] 50%"

// TestRenderPanel_DegradationMatrix pins the panel assembler's join/filter
// behaviour AND the vcs group's intra-degradation, asserting the EXACT full
// joined panel string for every combination that matters. Sections join with
// " │ " (space U+2502 space) in the order plan │ model │ vcs │ ctx │ cost.
//
// Each row builds its payload, runs the real RenderPanel entry point, asserts
// the exact expected string, and checks the no-double-separator property
// (no "│ │", no leading/trailing " │ ").
func TestRenderPanel_DegradationMatrix(t *testing.T) {
	// Each row gets its own temp root so git/crafter state is hermetic. A row
	// supplies a setup func that prepares the root (repo/plan/no-git) and a
	// build func that produces the payload (the setup func receives the root so
	// the build func can set Workdir = root).
	type row struct {
		name  string
		setup func(t *testing.T, root string)
		build func(root string) Payload
		want  string
	}

	// Common payload fragments reused across rows.
	const (
		modelStr = "Opus 4.8 1M (high)"
		ctxStr   = "[████░░░░░░] 42%"
		costStr  = "$0.42"
	)
	projTok := ansiDim + "crafter" + ansiReset
	diffTok := ansiGreen + "+120" + ansiReset + "/" + ansiRed + "-30" + ansiReset

	rows := []row{
		{
			// 1. All five sections present.
			name:  "all_five_present",
			setup: func(t *testing.T, root string) { makePlanRepo(t, root, "feat/panel") },
			build: func(root string) Payload {
				return Payload{
					Workdir:           root,
					ModelDisplayName:  "Opus 4.8",
					ContextWindowSize: 1_000_000,
					EffortLevel:       "high",
					UsedPercentage:    floatPtr(42),
					TotalCostUSD:      floatPtr(0.42),
					ProjectDir:        "/some/path/crafter",
					TotalLinesAdded:   120,
					TotalLinesRemoved: 30,
				}
			},
			want: planSeg + " │ " + modelStr + " │ " + projTok + " ⎇ feat/panel " + diffTok + " │ " + ctxStr + " │ " + costStr,
		},
		{
			// 2. Plan absent (no active task) — rest renders.
			name:  "plan_absent",
			setup: func(t *testing.T, root string) { makeRepo(t, root, "feat/panel") },
			build: func(root string) Payload {
				return Payload{
					Workdir:           root,
					ModelDisplayName:  "Opus 4.8",
					ContextWindowSize: 1_000_000,
					EffortLevel:       "high",
					UsedPercentage:    floatPtr(42),
					TotalCostUSD:      floatPtr(0.42),
					ProjectDir:        "/some/path/crafter",
					TotalLinesAdded:   120,
					TotalLinesRemoved: 30,
				}
			},
			want: modelStr + " │ " + projTok + " ⎇ feat/panel " + diffTok + " │ " + ctxStr + " │ " + costStr,
		},
		{
			// 3. Model absent (empty display_name) — model section dropped.
			name:  "model_absent",
			setup: func(t *testing.T, root string) { makePlanRepo(t, root, "feat/panel") },
			build: func(root string) Payload {
				return Payload{
					Workdir:           root,
					ContextWindowSize: 1_000_000,
					EffortLevel:       "high",
					UsedPercentage:    floatPtr(42),
					TotalCostUSD:      floatPtr(0.42),
					ProjectDir:        "/some/path/crafter",
					TotalLinesAdded:   120,
					TotalLinesRemoved: 30,
				}
			},
			want: planSeg + " │ " + projTok + " ⎇ feat/panel " + diffTok + " │ " + ctxStr + " │ " + costStr,
		},
		{
			// 4. Ctx absent (null used_percentage) — ctx section dropped.
			name:  "ctx_absent",
			setup: func(t *testing.T, root string) { makePlanRepo(t, root, "feat/panel") },
			build: func(root string) Payload {
				return Payload{
					Workdir:           root,
					ModelDisplayName:  "Opus 4.8",
					ContextWindowSize: 1_000_000,
					EffortLevel:       "high",
					UsedPercentage:    nil,
					TotalCostUSD:      floatPtr(0.42),
					ProjectDir:        "/some/path/crafter",
					TotalLinesAdded:   120,
					TotalLinesRemoved: 30,
				}
			},
			want: planSeg + " │ " + modelStr + " │ " + projTok + " ⎇ feat/panel " + diffTok + " │ " + costStr,
		},
		{
			// 5. Cost absent (zero/absent total_cost_usd) — cost section dropped.
			name:  "cost_absent",
			setup: func(t *testing.T, root string) { makePlanRepo(t, root, "feat/panel") },
			build: func(root string) Payload {
				return Payload{
					Workdir:           root,
					ModelDisplayName:  "Opus 4.8",
					ContextWindowSize: 1_000_000,
					EffortLevel:       "high",
					UsedPercentage:    floatPtr(42),
					TotalCostUSD:      nil,
					ProjectDir:        "/some/path/crafter",
					TotalLinesAdded:   120,
					TotalLinesRemoved: 30,
				}
			},
			want: planSeg + " │ " + modelStr + " │ " + projTok + " ⎇ feat/panel " + diffTok + " │ " + ctxStr,
		},
		{
			// 6. VCS group: project + branch + diff present.
			name:  "vcs_project_branch_diff",
			setup: func(t *testing.T, root string) { makeRepo(t, root, "feat/panel") },
			build: func(root string) Payload {
				return Payload{
					Workdir:           root,
					ProjectDir:        "/some/path/crafter",
					TotalLinesAdded:   120,
					TotalLinesRemoved: 30,
				}
			},
			want: projTok + " ⎇ feat/panel " + diffTok,
		},
		{
			// 7. VCS group: project absent → no leading space before ⎇.
			name:  "vcs_project_absent",
			setup: func(t *testing.T, root string) { makeRepo(t, root, "feat/panel") },
			build: func(root string) Payload {
				return Payload{
					Workdir:           root,
					ProjectDir:        "",
					TotalLinesAdded:   120,
					TotalLinesRemoved: 30,
				}
			},
			want: "⎇ feat/panel " + diffTok,
		},
		{
			// 8. VCS group: branch absent (no git). The branch token — and the diff
			// that attaches to it — is omitted; only the project token renders. The
			// diff does NOT render without a branch (verified against the assembler).
			name:  "vcs_branch_absent",
			setup: func(t *testing.T, root string) { /* no .git */ },
			build: func(root string) Payload {
				return Payload{
					Workdir:           root,
					ProjectDir:        "/some/path/crafter",
					TotalLinesAdded:   120,
					TotalLinesRemoved: 30,
				}
			},
			want: projTok,
		},
		{
			// 9. VCS group: diff zero (both counts 0) → no +N/-N suffix.
			name:  "vcs_diff_zero",
			setup: func(t *testing.T, root string) { makeRepo(t, root, "feat/panel") },
			build: func(root string) Payload {
				return Payload{
					Workdir:           root,
					ProjectDir:        "/some/path/crafter",
					TotalLinesAdded:   0,
					TotalLinesRemoved: 0,
				}
			},
			want: projTok + " ⎇ feat/panel",
		},
		{
			// 10. Whole VCS group absent (no project, no branch) → group dropped.
			name:  "vcs_group_absent",
			setup: func(t *testing.T, root string) { /* no .git */ },
			build: func(root string) Payload {
				return Payload{
					Workdir:           root,
					ProjectDir:        "",
					TotalLinesAdded:   5,
					TotalLinesRemoved: 2,
				}
			},
			want: "",
		},
		{
			// 11. All-absent: no task, empty payload, no git → panel == "".
			name:  "all_absent_empty_panel",
			setup: func(t *testing.T, root string) { /* no .git, no .crafter */ },
			build: func(root string) Payload {
				return Payload{Workdir: root}
			},
			want: "",
		},
	}

	for _, tc := range rows {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			tc.setup(t, root)
			got := RenderPanel(tc.build(root))

			if got != tc.want {
				t.Errorf("RenderPanel mismatch:\n  got  %q\n  want %q", got, tc.want)
			}

			// No-double-separator property: no "│ │", and no leading/trailing
			// " │ " around the assembled panel.
			if strings.Contains(got, "│ │") {
				t.Errorf("panel contains a doubled separator %q in %q", "│ │", got)
			}
			if strings.HasPrefix(got, " │ ") {
				t.Errorf("panel has a leading separator: %q", got)
			}
			if strings.HasSuffix(got, " │ ") {
				t.Errorf("panel has a trailing separator: %q", got)
			}
		})
	}
}

// ---------------------------------------------------------------------------

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
