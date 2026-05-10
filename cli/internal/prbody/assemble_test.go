package prbody

import (
	"path/filepath"
	"testing"
)

// taskFileWith writes a minimal task file containing exactly the given
// decisions body (may be empty) under a "## Decisions" heading, and returns
// the file path.
func taskFileWith(t *testing.T, decisionsBody string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "task.md")
	var content string
	if decisionsBody != "" {
		content = "## Decisions\n\n" + decisionsBody + "\n"
	} else {
		// A task file with no Decisions section.
		content = "# Task\n\nSome content.\n"
	}
	writeFile(t, path, content)
	return path
}

// uatEntry is a minimal single-line UAT NDJSON line.
const uatEntry = `{"title":"Check login","verify":"Open /login and confirm redirect.","why_manual":"Needs browser."}`

// gapEntry is a minimal Gap NDJSON line.
const gapEntry = `{"title":"No rollback path","detail":"Migration 0042 has no down migration.","followup":"Add down migration."}`

// decisionsBody is the content that will be placed under ## Decisions in the
// task file — two H3 decision entries.
const decisionsBody = "### Decision 1\n\nUse Markdown."

func TestAssemble(t *testing.T) {
	tests := []struct {
		name          string
		withUAT       bool
		withGaps      bool
		decisionsBody string
		wantEmpty     bool
		wantContains  []string
		wantAbsent    []string
		wantOrder     [][2]string // pairs where first must precede second
	}{
		{
			name:         "a: all empty → empty output",
			wantEmpty:    true,
			wantAbsent:   []string{"## Manual QA Plan", "## Known Gaps", "## Decisions"},
		},
		{
			name:         "b: only UAT non-empty → only Manual QA Plan heading",
			withUAT:      true,
			wantContains: []string{"## Manual QA Plan", "Check login"},
			wantAbsent:   []string{"## Known Gaps", "## Decisions"},
		},
		{
			name:         "c: only Gaps non-empty → only Known Gaps heading",
			withGaps:     true,
			wantContains: []string{"## Known Gaps", "No rollback path"},
			wantAbsent:   []string{"## Manual QA Plan", "## Decisions"},
		},
		{
			name:          "d: only Decisions non-empty → only Decisions heading",
			decisionsBody: decisionsBody,
			wantContains:  []string{"## Decisions", "Decision 1"},
			wantAbsent:    []string{"## Manual QA Plan", "## Known Gaps"},
		},
		{
			name:         "e: UAT + Gaps non-empty → both headings in correct order",
			withUAT:      true,
			withGaps:     true,
			wantContains: []string{"## Manual QA Plan", "## Known Gaps", "Check login", "No rollback path"},
			wantAbsent:   []string{"## Decisions"},
			wantOrder: [][2]string{
				{"## Manual QA Plan", "## Known Gaps"},
			},
		},
		{
			name:          "f: all three non-empty → all three headings in locked order",
			withUAT:       true,
			withGaps:      true,
			decisionsBody: decisionsBody,
			wantContains:  []string{"## Manual QA Plan", "## Known Gaps", "## Decisions", "Check login", "No rollback path", "Decision 1"},
			wantOrder: [][2]string{
				{"## Manual QA Plan", "## Known Gaps"},
				{"## Known Gaps", "## Decisions"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runDir := t.TempDir()

			if tc.withUAT {
				writeFile(t, filepath.Join(runDir, uatFilename), uatMarker+"\n"+uatEntry+"\n")
			}
			if tc.withGaps {
				writeFile(t, filepath.Join(runDir, gapFilename), gapMarker+"\n"+gapEntry+"\n")
			}

			taskFile := taskFileWith(t, tc.decisionsBody)

			got, err := Assemble(runDir, taskFile)
			if err != nil {
				t.Fatalf("Assemble returned unexpected error: %v", err)
			}

			if tc.wantEmpty {
				if got != "" {
					t.Errorf("expected empty output, got:\n%s", got)
				}
				return
			}

			for _, s := range tc.wantContains {
				if !contains(got, s) {
					t.Errorf("expected output to contain %q, got:\n%s", s, got)
				}
			}

			for _, s := range tc.wantAbsent {
				if contains(got, s) {
					t.Errorf("expected output NOT to contain %q, got:\n%s", s, got)
				}
			}

			for _, pair := range tc.wantOrder {
				a, b := pair[0], pair[1]
				ia := indexOf(got, a)
				ib := indexOf(got, b)
				if ia < 0 {
					t.Errorf("expected %q to appear in output", a)
					continue
				}
				if ib < 0 {
					t.Errorf("expected %q to appear in output", b)
					continue
				}
				if ia >= ib {
					t.Errorf("expected %q to precede %q in output:\n%s", a, b, got)
				}
			}
		})
	}
}

// TestAssemble_SectionsSeparatedByBlankLine verifies that when all three
// sections are present they are separated by exactly one blank line (i.e.,
// the two heading lines are not adjacent with zero blank lines between them).
func TestAssemble_SectionsSeparatedByBlankLine(t *testing.T) {
	runDir := t.TempDir()
	writeFile(t, filepath.Join(runDir, uatFilename), uatMarker+"\n"+uatEntry+"\n")
	writeFile(t, filepath.Join(runDir, gapFilename), gapMarker+"\n"+gapEntry+"\n")
	taskFile := taskFileWith(t, decisionsBody)

	got, err := Assemble(runDir, taskFile)
	if err != nil {
		t.Fatalf("Assemble returned unexpected error: %v", err)
	}

	// Between the end of the UAT content and the start of ## Known Gaps there
	// must be exactly one blank line (two consecutive newlines after content).
	if !contains(got, "\n\n## Known Gaps") {
		t.Errorf("expected blank line before ## Known Gaps:\n%s", got)
	}
	if !contains(got, "\n\n## Decisions") {
		t.Errorf("expected blank line before ## Decisions:\n%s", got)
	}
}

func contains(s, sub string) bool {
	return indexOf(s, sub) >= 0
}

func indexOf(s, sub string) int {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
