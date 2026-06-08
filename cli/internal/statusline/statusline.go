// Package statusline renders the Crafter plan position as a section of the
// Claude Code native statusLine bar.
//
// RenderPanel is the panel-assembly entry point. Given the decoded Claude Code
// statusLine payload, it assembles the full status panel by collecting non-empty
// sections and joining them with " │ ". Render is a thin wrapper that takes only
// the working directory. The sections are the plan section (produced by the
// four-rung cascade), the model section, the vcs section, the ctx section and
// the cost section.
//
// On any error (no active task, unreadable file, malformed plan) the plan
// section degrades to "" and the assembled panel returns "", preserving the
// silent-fail posture required by the statusline contract.
package statusline

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
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

// abbrevCapacity abbreviates a context-window size using a general k/M rule:
//
//	size >= 1_000_000 → "<size/1_000_000>M" (e.g. 1000000 → "1M")
//	size >= 1_000     → "<size/1_000>k"     (e.g. 200000 → "200k", 128000 → "128k")
//	otherwise         → the raw integer.
//
// Division is integer-clean for the pinned values (1M, 200k). Callers omit the
// capacity token entirely when size is 0 (this returns "0" in that case).
func abbrevCapacity(size int) string {
	switch {
	case size >= 1_000_000:
		return strconv.Itoa(size/1_000_000) + "M"
	case size >= 1_000:
		return strconv.Itoa(size/1_000) + "k"
	default:
		return strconv.Itoa(size)
	}
}

// reContextParenthetical matches a trailing " (… context)" group in a model
// display name, e.g. " (1M context)" in "Opus 4.8 (1M context)". It is used
// by modelSection to strip the redundant capacity parenthetical when crafter
// is itself appending its own abbreviated capacity token.
//
// The pattern requires:
//   - a literal space before the opening paren (so "Foo(context)" is not matched)
//   - any non-empty content before the word "context" (case-insensitive)
//   - optional whitespace around the content
//   - end-of-string anchor
var reContextParenthetical = regexp.MustCompile(`(?i)\s+\([^)]*context[^)]*\)\s*$`)

// modelSection renders the model section, e.g. "Opus 4.8 1M (high)".
//
// It concatenates ModelDisplayName, the abbreviated ContextWindowSize, and the
// EffortLevel in parentheses. Degradation:
//   - empty ModelDisplayName → "" (omit the whole section).
//   - ContextWindowSize == 0 → omit the capacity token (no "0"/"0k").
//   - empty EffortLevel → omit the " (...)" suffix.
//
// When ContextWindowSize > 0, any trailing " (… context)" parenthetical in
// ModelDisplayName is stripped before appending the abbreviated capacity, so
// a display name like "Opus 4.8 (1M context)" with ContextWindowSize=1000000
// renders as "Opus 4.8 1M (high)" rather than "Opus 4.8 (1M context) 1M (high)".
// The strip is skipped when ContextWindowSize == 0 so no information is lost.
//
// So the possible forms are "Opus 4.8 1M (high)", "Opus 4.8 1M",
// "Opus 4.8 (high)", and "Opus 4.8".
func modelSection(p Payload) string {
	if p.ModelDisplayName == "" {
		return ""
	}

	s := p.ModelDisplayName
	if p.ContextWindowSize > 0 {
		s = strings.TrimRight(reContextParenthetical.ReplaceAllString(s, ""), " ")
		s += " " + abbrevCapacity(p.ContextWindowSize)
	}
	if p.EffortLevel != "" {
		s += " (" + p.EffortLevel + ")"
	}
	return s
}

// costSection renders the cost section, e.g. "$0.42", from TotalCostUSD.
//
// It is omitted (returns "") when TotalCostUSD is nil or zero; only a positive
// value renders, formatted as "$%.2f".
func costSection(p Payload) string {
	if p.TotalCostUSD == nil || *p.TotalCostUSD == 0 {
		return ""
	}
	return fmt.Sprintf("$%.2f", *p.TotalCostUSD)
}

// Raw ANSI escape sequences emitted by the vcs section. Claude Code's statusLine
// renders them: dim grey for the project name, green for the added-lines count,
// red for the removed-lines count. ansiReset terminates each styled run.
const (
	ansiDim   = "\033[2m"
	ansiGreen = "\033[32m"
	ansiRed   = "\033[31m"
	ansiReset = "\033[0m"
)

// envBranchIcon is the environment variable that overrides the branch icon.
// Empty or unset falls back to defaultBranchIcon.
const envBranchIcon = "CRAFTER_STATUSLINE_BRANCH_ICON"

// defaultBranchIcon is the branch glyph used when envBranchIcon is unset/empty.
// U+2387 (ALTERNATIVE KEY SYMBOL) — chosen over a Nerd-font glyph so it renders
// without a special font installed.
const defaultBranchIcon = "⎇"

// branchIcon reads the configured branch icon from the environment at render
// time (so tests can override it via t.Setenv), falling back to the default.
func branchIcon() string {
	if v := os.Getenv(envBranchIcon); v != "" {
		return v
	}
	return defaultBranchIcon
}

// vcsSection renders the grouped vcs section: "<project> ⎇ <branch> +N/-N".
//
// It is built from up to three independent tokens, each degrading on its own:
//   - project token: basename of ProjectDir, dim grey; omitted when ProjectDir
//     is empty.
//   - branch token: the configured icon + the git branch (normal style);
//     omitted when readGitBranch returns "" (no repo / detached HEAD).
//   - diff token: "+N" (green) "/" "-N" (red) from TotalLinesAdded/Removed;
//     appended to the branch token only when there are changes (added>0 OR
//     removed>0). The diff attaches to the branch token, so when the branch is
//     absent the diff suffix is dropped too — it has nothing to attach to.
//
// Present tokens are joined with single spaces so a missing token never leaves a
// stray, leading, or doubled space. When no token is present the whole section
// is "" and the panel assembler drops it.
func vcsSection(p Payload) string {
	var tokens []string

	if p.ProjectDir != "" {
		tokens = append(tokens, ansiDim+filepath.Base(p.ProjectDir)+ansiReset)
	}

	branch := readGitBranch(p.Workdir)
	if branch != "" {
		token := branchIcon() + " " + branch
		if p.TotalLinesAdded > 0 || p.TotalLinesRemoved > 0 {
			diff := ansiGreen + "+" + strconv.Itoa(p.TotalLinesAdded) + ansiReset +
				"/" + ansiRed + "-" + strconv.Itoa(p.TotalLinesRemoved) + ansiReset
			token += " " + diff
		}
		tokens = append(tokens, token)
	}

	return strings.Join(tokens, " ")
}

// ctxSection renders the context-window section, e.g. "[████░░░░░░] 42%", from
// UsedPercentage.
//
// It reuses the plan bar via renderBar so the ctx bar and the plan bar are
// visually identical. UsedPercentage (a float, e.g. 42.5) is rounded to the
// nearest integer with math.Round — matching how the plan bar derives its
// integer percentage — and that integer drives both the bar fill (pct/10) and
// the displayed number. The section is omitted (returns "") when
// UsedPercentage is nil (context_window.used_percentage null/absent).
func ctxSection(p Payload) string {
	if p.UsedPercentage == nil {
		return ""
	}
	pct := int(math.Round(*p.UsedPercentage))
	return renderBar(pct) + " " + strconv.Itoa(pct) + "%"
}

// Payload carries the data the panel assembler needs to render every section.
// Workdir is the workspace root resolved from the Claude Code JSON payload
// (workspace.current_dir) or os.Getwd() as a fallback; the remaining fields are
// decoded from the rest of the Claude Code statusLine payload by the cmd layer.
//
// Pointer fields make "absent/null" distinguishable from a real zero value:
//   - UsedPercentage is nil when context_window.used_percentage is null/absent.
//   - TotalCostUSD is nil when cost.total_cost_usd is absent; a present zero is
//     *0, distinct from a positive value.
type Payload struct {
	Workdir string

	ModelDisplayName  string
	EffortLevel       string
	UsedPercentage    *float64
	ContextWindowSize int
	TotalCostUSD      *float64
	TotalLinesAdded   int
	TotalLinesRemoved int
	ProjectDir        string
}

// Render assembles and returns the full status panel from just a workdir. It is
// a thin wrapper over RenderPanel for callers that only have the workspace root
// (and for the existing plan-section tests). Additional payload-derived sections
// are rendered via RenderPanel.
func Render(workdir string) string {
	return RenderPanel(Payload{Workdir: workdir})
}

// RenderPanel assembles and returns the full status panel for the Claude Code
// status bar from the decoded payload.
//
// The panel is the non-empty sections joined by " │ ", in the order
// plan │ model │ vcs │ ctx │ cost. Returns "" when all sections are empty.
func RenderPanel(p Payload) string {
	var sections []string

	// Panel order: plan │ model │ vcs │ ctx │ cost.
	if s := planSection(p.Workdir); s != "" {
		sections = append(sections, s)
	}
	if s := modelSection(p); s != "" {
		sections = append(sections, s)
	}
	if s := vcsSection(p); s != "" {
		sections = append(sections, s)
	}
	if s := ctxSection(p); s != "" {
		sections = append(sections, s)
	}
	if s := costSection(p); s != "" {
		sections = append(sections, s)
	}

	return strings.Join(sections, " │ ")
}
