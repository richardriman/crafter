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

// Render returns the plan-position segment for the Claude Code status bar.
// workdir is the workspace root resolved from the Claude Code JSON payload
// (workspace.current_dir) or os.Getwd() as a fallback.
//
// Returns "" when there is no active Crafter task for the current directory
// and branch (the caller should emit no output in this case).
func Render(workdir string) string {
	match := resolveActiveTask(workdir)
	if match == nil {
		return ""
	}

	info := parsePlan(match.Path)
	return renderSegment(info)
}
