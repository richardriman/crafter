// Package statusline renders the Crafter plan position as a section of the
// Claude Code native statusLine bar.
//
// Render is the panel-assembly entry point. Given the working directory of the
// current Claude Code session, it assembles the full status panel by collecting
// non-empty sections and joining them with " │ ". Currently the only section is
// the plan section (produced by the four-rung cascade); additional sections
// (model, vcs, ctx, cost) will be added in Phase 2.
//
// On any error (no active task, unreadable file, malformed plan) the plan
// section degrades to "" and the assembled panel returns "", preserving the
// silent-fail posture required by the statusline contract.
package statusline

import (
	"fmt"
	"strings"
)

// Segment strings for rungs 2 and 3.  The rung-1 segment is built dynamically
// by renderSegment / renderExecuting in parse.go, but these two are fixed
// (rung-2) or format-string (rung-3) values that do not require plan parsing.
const (
	// segDone is returned when the current branch has a completed task and no
	// active task (rung 2).
	segDone = "✓ done"

	// segActiveElsewhereFmt is a fmt format string for the rung-3 segment.
	// Substitute with the integer count of active tasks on other branches.
	// e.g. fmt.Sprintf(segActiveElsewhereFmt, 2) → "2 active elsewhere"
	segActiveElsewhereFmt = "%d active elsewhere"
)

// planSection returns the plan-position section string for the given workdir.
// It applies the four-rung cascade with strict priority:
//  1. Active task on current branch  → full plan-progress segment.
//  2. Completed task on current branch, no active task → "✓ done".
//  3. Active tasks on other branches → "N active elsewhere".
//  4. Otherwise → "" (empty section).
//
// Returns "" when .crafter dir or git branch are unavailable — those guards
// suppress the plan section only, not the whole panel.
func planSection(workdir string) string {
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

// Render assembles and returns the full status panel for the Claude Code status
// bar. workdir is the workspace root resolved from the Claude Code JSON payload
// (workspace.current_dir) or os.Getwd() as a fallback.
//
// The panel is the non-empty sections joined by " │ ". Currently the only
// section is the plan section; additional sections will be added in Phase 2.
// Returns "" when all sections are empty.
func Render(workdir string) string {
	var sections []string

	if s := planSection(workdir); s != "" {
		sections = append(sections, s)
	}

	return strings.Join(sections, " │ ")
}
