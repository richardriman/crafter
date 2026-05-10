package prbody

import "strings"

// Assemble combines the three appended-sections sources into a single Markdown
// block. For each source, if the rendered content is non-empty the corresponding
// H2 heading is prepended; if it is empty, both heading and content are omitted.
//
// Section ordering (locked per Phase 1 Decision 3):
//
//	## Manual QA Plan  (from RenderUAT)
//	## Known Gaps      (from RenderGaps)
//	## Decisions       (from ExtractDecisions)
//
// Non-empty sections are separated by exactly one blank line. If all three
// sources are empty, Assemble returns ("", nil).
func Assemble(runDir, taskFile string) (string, error) {
	uat, err := RenderUAT(runDir)
	if err != nil {
		return "", err
	}

	gaps, err := RenderGaps(runDir)
	if err != nil {
		return "", err
	}

	decisions, err := ExtractDecisions(taskFile)
	if err != nil {
		return "", err
	}

	var sections []string

	if uat != "" {
		sections = append(sections, "## Manual QA Plan\n\n"+uat)
	}
	if gaps != "" {
		sections = append(sections, "## Known Gaps\n\n"+gaps)
	}
	if decisions != "" {
		sections = append(sections, "## Decisions\n\n"+decisions)
	}

	return strings.Join(sections, "\n\n"), nil
}
