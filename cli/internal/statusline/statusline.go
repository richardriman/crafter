// Package statusline renders the current Crafter plan position as a single
// composable segment for Claude Code's native statusLine bar.
//
// Render is the entry point. Given the working directory of the current
// Claude Code session, it:
//  1. Resolves the active task file (Step 1.2 — git branch → task file lookup).
//  2. Parses the task Markdown and extracts the current phase/step position
//     from the Plan section (Step 1.3 — plan parsing and segment rendering).
//  3. Returns a short, display-ready string such as
//     "crafter · Phase 1/3 · 2/4 [██░░░░░░░░] 50%" that the statusline bar
//     can embed directly.
//
// On any error (no active task, unreadable file, malformed plan) Render
// returns "" so the caller can silently produce no output, preserving the
// silent-fail posture required by the statusline contract.
package statusline

import "fmt"

// Segment strings for rungs 2 and 3.  The rung-1 segment is built dynamically
// by renderSegment / renderExecuting in parse.go, but these two are fixed
// (rung-2) or format-string (rung-3) values that do not require plan parsing.
const (
	// segDone is returned when the current branch has a completed task and no
	// active task (rung 2).
	segDone = "crafter · ✓ done"

	// segActiveElsewhereFmt is a fmt format string for the rung-3 segment.
	// Substitute with the integer count of active tasks on other branches.
	// e.g. fmt.Sprintf(segActiveElsewhereFmt, 2) → "crafter · 2 active elsewhere"
	segActiveElsewhereFmt = "crafter · %d active elsewhere"
)

// Render returns the plan-position segment for the Claude Code status bar.
// workdir is the workspace root resolved from the Claude Code JSON payload
// (workspace.current_dir) or os.Getwd() as a fallback.
//
// Precedence (strict short-circuit):
//  1. Active task on current branch  → full plan-progress segment (parsePlan + renderSegment).
//  2. Completed task on current branch, no active task → "crafter · ✓ done".
//  3. Active tasks on other branches → "crafter · N active elsewhere".
//  4. Otherwise → "" (no output).
//
// Returns "" on any setup failure (no .crafter dir, no git repo, detached HEAD)
// so the caller can silently produce no output, preserving the silent-fail
// posture required by the statusline contract.
func Render(workdir string) string {
	ctxDir := findCrafterDir(workdir)
	if ctxDir == "" {
		return ""
	}

	branch := readGitBranch(workdir)
	if branch == "" {
		return ""
	}

	cls := classifyTasks(ctxDir, branch)

	// Rung 1 — active task on current branch.
	if cls.ActiveCurrent != "" {
		info := parsePlan(cls.ActiveCurrent)
		return renderSegment(info)
	}

	// Rung 2 — completed task on current branch, no active task.
	if cls.CompletedCurrent != "" {
		return segDone
	}

	// Rung 3 — active tasks on other branches (zero count falls through to rung 4).
	if cls.ActiveOtherCount > 0 {
		return fmt.Sprintf(segActiveElsewhereFmt, cls.ActiveOtherCount)
	}

	// Rung 4 — nothing relevant found.
	return ""
}
