package prbody

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeTaskFile writes content to a temporary file inside t.TempDir() and
// returns the full path. Calls t.Fatal on any error.
func writeTaskFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "task.md")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("writing task fixture: %v", err)
	}
	return path
}

// ---------------------------------------------------------------------------
// Edge case (i): "## Decisions" followed immediately by another "## " heading
// → content is empty → returns ""
// ---------------------------------------------------------------------------

func TestExtractDecisions_EmptySection(t *testing.T) {
	content := `## Metadata
- foo

## Decisions

## Outcome
done
`
	path := writeTaskFile(t, content)
	got, err := ExtractDecisions(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "" {
		t.Errorf("expected empty string for section with only blank lines, got %q", got)
	}
}

// ---------------------------------------------------------------------------
// Edge case (ii): "### Decision (...)" H3 sub-headings inside the section →
// preserved verbatim; H3 does NOT end the section.
// ---------------------------------------------------------------------------

func TestExtractDecisions_H3SubHeadingsPreserved(t *testing.T) {
	content := `## Metadata
- foo

## Decisions

### Decision 1 — Foo (Accepted)

**Chosen:** Option A.

### Decision 2 — Bar (Accepted)

**Chosen:** Option B.

## Outcome
done
`
	path := writeTaskFile(t, content)
	got, err := ExtractDecisions(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Both H3 headings must survive verbatim.
	if !strings.Contains(got, "### Decision 1 — Foo (Accepted)") {
		t.Errorf("expected H3 heading 'Decision 1' to be preserved, got:\n%s", got)
	}
	if !strings.Contains(got, "### Decision 2 — Bar (Accepted)") {
		t.Errorf("expected H3 heading 'Decision 2' to be preserved, got:\n%s", got)
	}
	// The H2 boundary heading must NOT appear in the output.
	if strings.Contains(got, "## Outcome") {
		t.Errorf("expected '## Outcome' to NOT appear in extracted body, got:\n%s", got)
	}
}

// ---------------------------------------------------------------------------
// Edge case (iii): "## Decisions" at end of file with no following "## " →
// consumes to EOF.
// ---------------------------------------------------------------------------

func TestExtractDecisions_AtEOF(t *testing.T) {
	content := `## Metadata
- foo

## Decisions

### Decision 1 — Something

**Chosen:** X.
`
	path := writeTaskFile(t, content)
	got, err := ExtractDecisions(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(got, "### Decision 1 — Something") {
		t.Errorf("expected decision content at EOF, got:\n%s", got)
	}
	if !strings.Contains(got, "**Chosen:** X.") {
		t.Errorf("expected chosen text at EOF, got:\n%s", got)
	}
}

// ---------------------------------------------------------------------------
// Edge case (iv): Content between heading and next "## " is blank/whitespace/
// HTML comments only → returns "".
// ---------------------------------------------------------------------------

func TestExtractDecisions_BlankAndCommentOnly(t *testing.T) {
	content := `## Metadata
- foo

## Decisions

<!-- no decisions recorded yet -->

## Outcome
done
`
	path := writeTaskFile(t, content)
	got, err := ExtractDecisions(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "" {
		t.Errorf("expected empty string for HTML-comment-only section, got %q", got)
	}
}

// ---------------------------------------------------------------------------
// Happy path: section with real content and H3 sub-headings is extracted
// verbatim with leading/trailing blank lines trimmed.
// ---------------------------------------------------------------------------

func TestExtractDecisions_HappyPath(t *testing.T) {
	content := `## Metadata
- date: 2026-05-09

## Plan
Do stuff.

## Decisions

### Decision 1 — Use Go (Accepted)

**Chosen:** Go binary.

**Rationale:** Reliable.

## Outcome
Shipped.
`
	path := writeTaskFile(t, content)
	got, err := ExtractDecisions(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Must start with the first non-blank line after "## Decisions".
	if !strings.HasPrefix(got, "### Decision 1 — Use Go (Accepted)") {
		t.Errorf("expected output to start with first H3, got:\n%q", got)
	}
	// Must not have leading blank lines.
	if strings.HasPrefix(got, "\n") {
		t.Errorf("expected no leading newline, got:\n%q", got)
	}
	// Must not have trailing blank lines.
	if strings.HasSuffix(got, "\n") {
		t.Errorf("expected no trailing newline, got:\n%q", got)
	}
	// Must contain the rationale text.
	if !strings.Contains(got, "**Rationale:** Reliable.") {
		t.Errorf("expected rationale text in output, got:\n%s", got)
	}
	// The following H2 section must not appear.
	if strings.Contains(got, "## Outcome") {
		t.Errorf("expected '## Outcome' to not appear in extracted body, got:\n%s", got)
	}
}

// ---------------------------------------------------------------------------
// Code-fence-aware: "## Decisions" inside a fenced code block must NOT be
// matched as the section heading; only the true heading outside is matched.
// ---------------------------------------------------------------------------

func TestExtractDecisions_InsideCodeFenceNotMatched(t *testing.T) {
	content := `## Request

Here is the desired outcome:

` + "```" + `
## Manual QA Plan
<uat content>

## Known Gaps
<gaps content>

## Decisions
<decisions content>
` + "```" + `

## Decisions

### Decision 1 — Real one (Accepted)

**Chosen:** Option A.

## Outcome
done
`
	path := writeTaskFile(t, content)
	got, err := ExtractDecisions(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Must contain the real decision, not the fenced fake one.
	if !strings.Contains(got, "### Decision 1 — Real one (Accepted)") {
		t.Errorf("expected real decision content, got:\n%s", got)
	}
	// The fenced placeholder text must not appear as extracted content.
	if strings.Contains(got, "<decisions content>") {
		t.Errorf("expected fenced content to NOT appear in extracted output, got:\n%s", got)
	}
}

// ---------------------------------------------------------------------------
// Code-fence-aware with tilde fences (~~~).
// ---------------------------------------------------------------------------

func TestExtractDecisions_TildeFenceNotMatched(t *testing.T) {
	content := `## Request

` + "~~~" + `
## Decisions
inside tilde fence
` + "~~~" + `

## Decisions

Real content here.

## Outcome
done
`
	path := writeTaskFile(t, content)
	got, err := ExtractDecisions(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(got, "Real content here.") {
		t.Errorf("expected real content, got:\n%s", got)
	}
	if strings.Contains(got, "inside tilde fence") {
		t.Errorf("expected tilde-fenced content to NOT appear in extracted output, got:\n%s", got)
	}
}

// ---------------------------------------------------------------------------
// Missing file → error returned.
// ---------------------------------------------------------------------------

func TestExtractDecisions_MissingFile(t *testing.T) {
	_, err := ExtractDecisions("/nonexistent/path/task.md")
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}

// ---------------------------------------------------------------------------
// "## Decisions" heading absent → returns "".
// ---------------------------------------------------------------------------

func TestExtractDecisions_NoDecisionsHeading(t *testing.T) {
	content := `## Metadata
- foo

## Plan
Do stuff.

## Outcome
done
`
	path := writeTaskFile(t, content)
	got, err := ExtractDecisions(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "" {
		t.Errorf("expected empty string when no ## Decisions heading, got %q", got)
	}
}

// ---------------------------------------------------------------------------
// Live task file smoke test: extract from the real GH#16 task file and verify
// H3 sub-headings are preserved (non-hermetic, skipped if file absent).
// ---------------------------------------------------------------------------

func TestExtractDecisions_LiveGH16TaskFile(t *testing.T) {
	taskFile := "../../../.crafter/tasks/20260509-feat-gh-16-buffer-skill.md"

	// Resolve relative to the test binary working directory.
	abs, err := filepath.Abs(taskFile)
	if err != nil {
		t.Skipf("cannot resolve path %s: %v", taskFile, err)
	}
	if _, err := os.Stat(abs); os.IsNotExist(err) {
		t.Skipf("live task file not found at %s — skipping non-hermetic test", abs)
	}

	got, err := ExtractDecisions(abs)
	if err != nil {
		t.Fatalf("unexpected error on live file: %v", err)
	}

	// The GH#16 Decisions section has multiple "### Decision (...)" sub-headings.
	if !strings.Contains(got, "### Decision") {
		t.Errorf("expected H3 Decision sub-headings in GH#16 output, got:\n%s", got)
	}
	// The section heading itself must not appear in the body.
	if strings.HasPrefix(got, "## Decisions") {
		t.Errorf("expected extracted body to NOT start with the heading line, got:\n%s", got)
	}
	// Must not be empty.
	if got == "" {
		t.Error("expected non-empty extraction from live GH#16 task file")
	}
}

// ---------------------------------------------------------------------------
// Live task file smoke test: extract from the real GH#17 task file and verify
// the fenced "## Decisions" in the Request section is NOT matched.
// ---------------------------------------------------------------------------

func TestExtractDecisions_LiveGH17TaskFile(t *testing.T) {
	taskFile := "../../../.crafter/tasks/20260510-feat-gh-17-pr-composer.md"

	abs, err := filepath.Abs(taskFile)
	if err != nil {
		t.Skipf("cannot resolve path %s: %v", taskFile, err)
	}
	if _, err := os.Stat(abs); os.IsNotExist(err) {
		t.Skipf("live task file not found at %s — skipping non-hermetic test", abs)
	}

	got, err := ExtractDecisions(abs)
	if err != nil {
		t.Fatalf("unexpected error on live GH#17 file: %v", err)
	}

	// The real ## Decisions section has multiple "### Decision N —" lines.
	if !strings.Contains(got, "### Decision 1") {
		t.Errorf("expected ### Decision 1 from the real section, got:\n%s", got)
	}
	// Content from inside the fenced block in the Request section must not appear.
	if strings.Contains(got, "<contents of the task file") {
		t.Errorf("expected fenced placeholder to NOT appear in extracted body, got:\n%s", got)
	}
	if got == "" {
		t.Error("expected non-empty extraction from live GH#17 task file")
	}
}
