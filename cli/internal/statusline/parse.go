package statusline

import (
	"bufio"
	"errors"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// planState represents the three observable states of a task's ## Plan section.
type planState int

const (
	planStateNone     planState = iota // no plan yet (_(pending)_ or no ## Plan heading)
	planStateDraft                     // **Plan status:** draft
	planStateApproved                  // **Plan status:** approved
)

// planInfo holds the parsed data from the ## Plan section.
type planInfo struct {
	state        planState
	totalPhases  int
	currentPhase int
	doneSteps    int
	totalSteps   int
}

// Glyphs used in the progress bar.
const (
	glyphFilled = "█"
	glyphEmpty  = "░"
	barSegments = 10
)

// rePhaseHeading matches any H3 or H4 heading whose text contains "Phase <N>",
// capturing the phase number.  e.g.:
//
//	"### Phase 1 — outcome"
//	"#### Phase 2 — outcome"
var rePhaseHeading = regexp.MustCompile(`^#{3,4}\s+Phase\s+(\d+)`)

// isGateLine reports whether a checkbox line is a workflow gate/ceremony line
// that must be excluded from step counts (A3).
//
// Gate patterns observed in real task files:
//
//   - [ ] Phase 1 verification
//   - [ ] Phase 1 review
//   - [ ] Phase verification          (template uses this shorter form)
//   - [ ] Phase review
//   - [x] **Phase 1 verification.**   (checked variant with bold markup)
//   - [x] **Phase 1 review.**
//   - [x] **Phase 1 verification.** — long note...  (real files append a note)
//
// Post-change checkboxes (also excluded):
//
//   - [x] **STATE.md and skillbook update** per ...
//   - [x] **Task file completion** per ...
//   - [x] **Follow-up note** ...
//
// The approach: strip the checkbox prefix and any bold markers, then match
// the canonical gate wording.
var (
	// reGatePhase matches "Phase [N] verification" or "Phase [N] review" at the
	// start of the body. The alternation (verification|review) is explicitly
	// grouped to avoid |‑precedence bugs. The \b word boundary prevents partial
	// matches like "Phase 1 reviewer" or "Phase 1 verification_notes". A word
	// boundary is used rather than a full-line anchor ($) so that real gate lines
	// with trailing notes (e.g. "Phase 1 verification.** — crafter-verifier …")
	// are still detected correctly.
	reGatePhase      = regexp.MustCompile(`(?i)^Phase(\s+\d+)?\s+(verification|review)\b`)
	reGatePostChange = regexp.MustCompile(`(?i)^(STATE\.md|Task file completion|Follow-up note)`)
)

// checkboxBody extracts the text after the `- [ ] ` or `- [x] ` prefix,
// stripping surrounding bold markers (`**`).
// Returns ("", false) if the line is not a checkbox line.
func checkboxBody(line string) (body string, checked bool, ok bool) {
	// Must start with "- [" followed by exactly one character + "] "
	if !strings.HasPrefix(line, "- [") || len(line) < 6 {
		return "", false, false
	}
	mark := line[3]
	if line[4] != ']' || line[5] != ' ' {
		return "", false, false
	}
	body = strings.TrimSpace(line[6:])
	checked = mark == 'x' || mark == 'X'
	// Strip bold wrapper: **text** or **text.**
	body = strings.TrimPrefix(body, "**")
	body = strings.TrimSuffix(body, "**")
	body = strings.TrimSuffix(body, ".")
	body = strings.TrimSpace(body)
	return body, checked, true
}

// isGate reports whether the checkbox body text matches a gate/ceremony pattern.
// It additionally strips leading/trailing `*` characters and trailing `.`/`:`
// before matching, so bold-wrapped variants like "**Phase 1 verification.**"
// are normalised to "Phase 1 verification" before the regex runs.
func isGate(body string) bool {
	b := strings.TrimRight(strings.TrimLeft(body, "*"), "*")
	b = strings.TrimRight(b, ".:")
	b = strings.TrimSpace(b)
	return reGatePhase.MatchString(b) || reGatePostChange.MatchString(b)
}

// parsePlan reads the task file at path and returns the parsed planInfo.
// On any error or unexpected content it degrades gracefully: returns the
// zero planInfo (planStateNone) rather than panicking.
func parsePlan(path string) planInfo {
	f, err := os.Open(path)
	if err != nil {
		return planInfo{}
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), maxScannerBytes)

	var (
		inPlan bool // are we inside the ## Plan section?

		state planState = planStateNone

		totalPhases  int
		currentPhase int // phase number of the first unchecked work-step; 0 = not yet found
		lastPhase    int // most recently seen phase number

		doneSteps  int
		totalSteps int

		foundFirstUnchecked bool
	)

	for scanner.Scan() {
		line := scanner.Text()

		// Track the ## Plan section boundary.
		if strings.HasPrefix(line, "## ") {
			if inPlan {
				// Leaving the Plan section — stop.
				break
			}
			if line == "## Plan" {
				inPlan = true
			}
			continue
		}

		if !inPlan {
			continue
		}

		// --- Inside ## Plan ---

		// Detect plan state from the status line.
		if strings.HasPrefix(line, "**Plan status:**") {
			status := strings.TrimSpace(strings.TrimPrefix(line, "**Plan status:**"))
			switch status {
			case "approved":
				state = planStateApproved
			case "draft":
				state = planStateDraft
			}
			continue
		}

		// Detect _(pending)_ marker. The state stays planStateNone regardless;
		// we just skip the line so it is not mis-parsed as a step.
		if strings.Contains(line, "_(pending)_") {
			continue
		}

		// Detect phase headings (H3 or H4 containing "Phase N").
		if m := rePhaseHeading.FindStringSubmatch(line); m != nil {
			n, err := strconv.Atoi(m[1])
			if err == nil {
				totalPhases++
				lastPhase = n
			}
			continue
		}

		// Detect checkbox lines.
		body, checked, ok := checkboxBody(line)
		if !ok {
			continue
		}

		// Skip gate/ceremony lines.
		if isGate(body) {
			continue
		}

		// Work step: count it.
		totalSteps++
		if checked {
			doneSteps++
		} else if !foundFirstUnchecked {
			// The first unchecked work-step determines the current phase.
			currentPhase = lastPhase
			foundFirstUnchecked = true
		}
	}

	// Tolerate scanner errors (e.g. ErrTooLong) — degrade to whatever we parsed.
	if err := scanner.Err(); err != nil && !errors.Is(err, bufio.ErrTooLong) {
		return planInfo{state: state}
	}

	// If all steps are done, current phase = last phase.
	if totalSteps > 0 && !foundFirstUnchecked {
		currentPhase = lastPhase
	}

	return planInfo{
		state:        state,
		totalPhases:  totalPhases,
		currentPhase: currentPhase,
		doneSteps:    doneSteps,
		totalSteps:   totalSteps,
	}
}

// renderSegment converts a planInfo into the display string.
// Returns "" when info represents the empty state (no active task).
func renderSegment(info planInfo) string {
	switch info.state {
	case planStateDraft:
		return "plan: awaiting approval"

	case planStateNone:
		return "planning"

	case planStateApproved:
		return renderExecuting(info)

	default:
		return ""
	}
}

// renderExecuting renders the full executing segment for an approved plan.
func renderExecuting(info planInfo) string {
	var sb strings.Builder

	// Phase segment — only when both currentPhase and totalPhases are positive.
	// A malformed plan (e.g. a work-step checkbox before the first Phase heading)
	// can leave currentPhase as 0; in that case we degrade silently by omitting
	// the phase indicator rather than emitting a confusing "Phase 0/N" string.
	hasPhase := info.totalPhases > 0 && info.currentPhase > 0
	if hasPhase {
		sb.WriteString("Phase ")
		sb.WriteString(strconv.Itoa(info.currentPhase))
		sb.WriteByte('/')
		sb.WriteString(strconv.Itoa(info.totalPhases))
	}

	// Step segment + bar + percent — only when there are steps to count.
	if info.totalSteps > 0 {
		pct := int(math.Round(float64(info.doneSteps) / float64(info.totalSteps) * 100))
		filled := pct / 10
		if filled > barSegments {
			filled = barSegments
		}

		if hasPhase {
			sb.WriteString(" · ")
		}
		sb.WriteString(strconv.Itoa(info.doneSteps))
		sb.WriteByte('/')
		sb.WriteString(strconv.Itoa(info.totalSteps))
		sb.WriteString(" [")
		for i := 0; i < barSegments; i++ {
			if i < filled {
				sb.WriteString(glyphFilled)
			} else {
				sb.WriteString(glyphEmpty)
			}
		}
		sb.WriteString("] ")
		sb.WriteString(strconv.Itoa(pct))
		sb.WriteByte('%')
	}

	return sb.String()
}
