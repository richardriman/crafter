package prbody

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeFile is a test helper that writes content to path, creating intermediate
// directories. Calls t.Fatal on any error.
func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("creating directories for %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("writing %s: %v", path, err)
	}
}

const uatMarker = `{"_marker":"uat-buffer","_format":"ndjson-v1"}`
const gapMarker = `{"_marker":"gaps-buffer","_format":"ndjson-v1"}`

// ---------------------------------------------------------------------------
// RenderUAT tests
// ---------------------------------------------------------------------------

func TestRenderUAT_MissingFile(t *testing.T) {
	// (iv) missing file → empty output
	dir := t.TempDir()
	got, err := RenderUAT(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "" {
		t.Errorf("expected empty string for missing file, got %q", got)
	}
}

func TestRenderUAT_ZeroByteFile(t *testing.T) {
	// (v) zero-byte file → empty output
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, uatFilename), "")
	got, err := RenderUAT(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "" {
		t.Errorf("expected empty string for zero-byte file, got %q", got)
	}
}

func TestRenderUAT_MarkerOnly(t *testing.T) {
	// (i) marker-only file → empty output
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, uatFilename), uatMarker+"\n")
	got, err := RenderUAT(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "" {
		t.Errorf("expected empty string for marker-only file, got %q", got)
	}
}

func TestRenderUAT_SingleLineVerify(t *testing.T) {
	// (vi) UAT short verify (single-line) → inline form
	dir := t.TempDir()
	entry := `{"title":"Confirm OAuth callback redirect","verify":"Click Sign In; confirm browser lands on /dashboard not /login.","why_manual":"Requires a live OAuth provider."}`
	writeFile(t, filepath.Join(dir, uatFilename), uatMarker+"\n"+entry+"\n")

	got, err := RenderUAT(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "- [ ] **Confirm OAuth callback redirect** — Click Sign In; confirm browser lands on /dashboard not /login."
	if got != want {
		t.Errorf("single-line UAT render mismatch:\n  got  %q\n  want %q", got, want)
	}
}

func TestRenderUAT_MultiLineVerify(t *testing.T) {
	// (vii) UAT multi-line verify → nested form with 2-space indent
	dir := t.TempDir()
	// verify contains a newline between two steps
	entry := `{"title":"Check session cookie flags","verify":"Deploy to staging and log in.\n\nOpen DevTools and confirm HttpOnly is set.","why_manual":"Requires a live TLS terminator."}`
	writeFile(t, filepath.Join(dir, uatFilename), uatMarker+"\n"+entry+"\n")

	got, err := RenderUAT(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Must start with the checkbox and bold title.
	if !strings.HasPrefix(got, "- [ ] **Check session cookie flags**\n") {
		t.Errorf("expected multi-line UAT to start with checkbox+title, got:\n%s", got)
	}
	// Must contain 2-space-indented continuation lines.
	if !strings.Contains(got, "\n  Deploy to staging and log in.") {
		t.Errorf("expected 2-space-indented verify content, got:\n%s", got)
	}
	// Blank separator line within verify must be "  " (2 spaces, not empty).
	if !strings.Contains(got, "\n  \n  Open DevTools") {
		t.Errorf("expected blank separator as 2-space line, got:\n%s", got)
	}
	// Must include _Why manual:_ because why_manual is non-empty.
	if !strings.Contains(got, "_Why manual:_ Requires a live TLS terminator.") {
		t.Errorf("expected _Why manual:_ line in multi-line entry, got:\n%s", got)
	}
}

func TestRenderUAT_MultiLineVerifyWithCodeFence(t *testing.T) {
	// (viii) UAT multi-line verify with embedded fenced code block
	dir := t.TempDir()
	// Backticks in JSON strings are unescaped; the verify value contains a real fenced block.
	entry := "{\"title\":\"Verify hover preview\",\"verify\":\"Run the app:\\n\\n```bash\\nDATA=fixtures/empty.json npm run dev\\n```\\n\\nHover and confirm.\",\"why_manual\":\"Requires a browser.\"}"
	writeFile(t, filepath.Join(dir, uatFilename), uatMarker+"\n"+entry+"\n")

	got, err := RenderUAT(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The fenced code block must survive in the output (real backticks, not escaped).
	if !strings.Contains(got, "```bash") {
		t.Errorf("expected fenced code block to survive rendering, got:\n%s", got)
	}
	// Continuation lines must be 2-space-indented.
	if !strings.Contains(got, "  ```bash") {
		t.Errorf("expected code fence line to be 2-space-indented, got:\n%s", got)
	}
	if !strings.Contains(got, "  DATA=fixtures/empty.json npm run dev") {
		t.Errorf("expected code block content to be 2-space-indented, got:\n%s", got)
	}
}

func TestRenderUAT_MultiLineVerifyEmptyWhyManual(t *testing.T) {
	// (ix) UAT multi-line verify with empty why_manual → _Why manual:_ line omitted
	dir := t.TempDir()
	entry := `{"title":"Check something","verify":"Step one.\n\nStep two.","why_manual":""}`
	writeFile(t, filepath.Join(dir, uatFilename), uatMarker+"\n"+entry+"\n")

	got, err := RenderUAT(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.Contains(got, "_Why manual:_") {
		t.Errorf("expected _Why manual:_ to be omitted when why_manual is empty, got:\n%s", got)
	}
	// Verify content must still appear.
	if !strings.Contains(got, "  Step one.") {
		t.Errorf("expected verify content to appear in output, got:\n%s", got)
	}
}

func TestRenderUAT_MultiLineVerifyMissingWhyManual(t *testing.T) {
	// (ix) UAT multi-line verify with missing why_manual field → _Why manual:_ line omitted
	dir := t.TempDir()
	// why_manual key is absent from the JSON object.
	entry := `{"title":"Check something else","verify":"Step A.\n\nStep B."}`
	writeFile(t, filepath.Join(dir, uatFilename), uatMarker+"\n"+entry+"\n")

	got, err := RenderUAT(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.Contains(got, "_Why manual:_") {
		t.Errorf("expected _Why manual:_ to be omitted when why_manual is missing, got:\n%s", got)
	}
}

func TestRenderUAT_MarkerPlusOneEntry(t *testing.T) {
	// (ii) marker + 1 entry → exactly one rendered block
	dir := t.TempDir()
	entry := `{"title":"Single entry","verify":"Do this.","why_manual":"Manual only."}`
	writeFile(t, filepath.Join(dir, uatFilename), uatMarker+"\n"+entry+"\n")

	got, err := RenderUAT(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "- [ ] **Single entry** — Do this."
	if got != want {
		t.Errorf("single-entry render mismatch:\n  got  %q\n  want %q", got, want)
	}
}

func TestRenderUAT_MarkerPlusMultipleEntries(t *testing.T) {
	// (iii) marker + multiple entries → joined blocks (newline-separated)
	dir := t.TempDir()
	e1 := `{"title":"First","verify":"Step 1.","why_manual":"x"}`
	e2 := `{"title":"Second","verify":"Step 2.","why_manual":"x"}`
	writeFile(t, filepath.Join(dir, uatFilename), uatMarker+"\n"+e1+"\n"+e2+"\n")

	got, err := RenderUAT(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(got, "- [ ] **First** — Step 1.") {
		t.Errorf("expected first entry in output, got:\n%s", got)
	}
	if !strings.Contains(got, "- [ ] **Second** — Step 2.") {
		t.Errorf("expected second entry in output, got:\n%s", got)
	}
	// Both entries must appear in order.
	firstIdx := strings.Index(got, "First")
	secondIdx := strings.Index(got, "Second")
	if firstIdx >= secondIdx {
		t.Errorf("expected First to precede Second in output, got:\n%s", got)
	}
}

// ---------------------------------------------------------------------------
// RenderGaps tests
// ---------------------------------------------------------------------------

func TestRenderGaps_MissingFile(t *testing.T) {
	// (iv) missing file → empty output
	dir := t.TempDir()
	got, err := RenderGaps(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "" {
		t.Errorf("expected empty string for missing file, got %q", got)
	}
}

func TestRenderGaps_ZeroByteFile(t *testing.T) {
	// (v) zero-byte file → empty output
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, gapFilename), "")
	got, err := RenderGaps(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "" {
		t.Errorf("expected empty string for zero-byte file, got %q", got)
	}
}

func TestRenderGaps_MarkerOnly(t *testing.T) {
	// (i) marker-only file → empty output
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, gapFilename), gapMarker+"\n")
	got, err := RenderGaps(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "" {
		t.Errorf("expected empty string for marker-only file, got %q", got)
	}
}

func TestRenderGaps_WithDetailAndFollowup(t *testing.T) {
	// (x) Gap with both detail and followup
	dir := t.TempDir()
	entry := `{"title":"Database migrations lack a rollback path","detail":"Migration 0042 adds an audit_log table. There is no down migration.","followup":"Add a down migration for 0042."}`
	writeFile(t, filepath.Join(dir, gapFilename), gapMarker+"\n"+entry+"\n")

	got, err := RenderGaps(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.HasPrefix(got, "- **Database migrations lack a rollback path** — Migration 0042") {
		t.Errorf("expected gap title+detail prefix, got:\n%s", got)
	}
	if !strings.Contains(got, "_Follow-up:_ Add a down migration for 0042.") {
		t.Errorf("expected _Follow-up:_ line, got:\n%s", got)
	}
	// The follow-up must be on an indented line after a blank separator.
	if !strings.Contains(got, "\n\n  _Follow-up:_") {
		t.Errorf("expected blank line before follow-up with 2-space indent, got:\n%s", got)
	}
}

func TestRenderGaps_EmptyFollowup(t *testing.T) {
	// (xi) Gap with empty followup → _Follow-up:_ line omitted
	dir := t.TempDir()
	entry := `{"title":"Rate limit not applied","detail":"Internal endpoints are unthrottled.","followup":""}`
	writeFile(t, filepath.Join(dir, gapFilename), gapMarker+"\n"+entry+"\n")

	got, err := RenderGaps(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.Contains(got, "_Follow-up:_") {
		t.Errorf("expected _Follow-up:_ to be omitted when followup is empty, got:\n%s", got)
	}
	if !strings.Contains(got, "Rate limit not applied") {
		t.Errorf("expected gap title in output, got:\n%s", got)
	}
}

func TestRenderGaps_MissingFollowupField(t *testing.T) {
	// (xi) Gap with missing followup field → _Follow-up:_ line omitted
	dir := t.TempDir()
	entry := `{"title":"No followup","detail":"Some detail."}`
	writeFile(t, filepath.Join(dir, gapFilename), gapMarker+"\n"+entry+"\n")

	got, err := RenderGaps(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.Contains(got, "_Follow-up:_") {
		t.Errorf("expected _Follow-up:_ to be omitted when followup field is missing, got:\n%s", got)
	}
}

func TestRenderGaps_MarkerPlusMultipleEntries(t *testing.T) {
	// (iii) marker + multiple entries → joined blocks
	dir := t.TempDir()
	e1 := `{"title":"First gap","detail":"Detail A.","followup":"Fix A."}`
	e2 := `{"title":"Second gap","detail":"Detail B.","followup":"Fix B."}`
	writeFile(t, filepath.Join(dir, gapFilename), gapMarker+"\n"+e1+"\n"+e2+"\n")

	got, err := RenderGaps(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(got, "First gap") {
		t.Errorf("expected first gap entry, got:\n%s", got)
	}
	if !strings.Contains(got, "Second gap") {
		t.Errorf("expected second gap entry, got:\n%s", got)
	}
	firstIdx := strings.Index(got, "First gap")
	secondIdx := strings.Index(got, "Second gap")
	if firstIdx >= secondIdx {
		t.Errorf("expected First gap to precede Second gap, got:\n%s", got)
	}
}

func TestRenderGaps_MultiLineDetail(t *testing.T) {
	// Multi-line detail with a single newline: first line inline after em-dash,
	// second line on its own line with 2-space indent.
	dir := t.TempDir()
	// detail = "Para 1\nPara 2"
	entry := `{"title":"Incomplete audit log","detail":"Para 1\nPara 2","followup":""}`
	writeFile(t, filepath.Join(dir, gapFilename), gapMarker+"\n"+entry+"\n")

	got, err := RenderGaps(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := "- **Incomplete audit log** — Para 1\n  Para 2"
	if got != want {
		t.Errorf("multi-line detail mismatch:\n  got  %q\n  want %q", got, want)
	}
}

func TestRenderGaps_MultiParagraphDetail(t *testing.T) {
	// Multi-paragraph detail (blank line between paragraphs): first line inline
	// after em-dash, blank separator becomes "  " (2 spaces), second paragraph
	// indented 2 spaces.
	dir := t.TempDir()
	// detail = "Para 1.\n\nPara 2."
	entry := `{"title":"Auth gap","detail":"Para 1.\n\nPara 2.","followup":""}`
	writeFile(t, filepath.Join(dir, gapFilename), gapMarker+"\n"+entry+"\n")

	got, err := RenderGaps(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := "- **Auth gap** — Para 1.\n  \n  Para 2."
	if got != want {
		t.Errorf("multi-paragraph detail mismatch:\n  got  %q\n  want %q", got, want)
	}
}

// ---------------------------------------------------------------------------
// indentBlock unit tests
// ---------------------------------------------------------------------------

func TestIndentBlock_SingleLine(t *testing.T) {
	got := indentBlock("hello world")
	want := "  hello world"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestIndentBlock_MultiLine(t *testing.T) {
	got := indentBlock("line one\nline two")
	want := "  line one\n  line two"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestIndentBlock_BlankLineBecomesDoubleSpace(t *testing.T) {
	// Blank separator lines must be "  " (2 spaces), not empty, per Decision 4.
	got := indentBlock("para one\n\npara two")
	// The middle line (originally empty) must become "  " not "".
	lines := strings.Split(got, "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d: %q", len(lines), got)
	}
	if lines[1] != "  " {
		t.Errorf("blank separator line: got %q, want %q", lines[1], "  ")
	}
}

// ---------------------------------------------------------------------------
// isMarkerLine unit tests
// ---------------------------------------------------------------------------

func TestIsMarkerLine_UATMarker(t *testing.T) {
	if !isMarkerLine(`{"_marker":"uat-buffer","_format":"ndjson-v1"}`) {
		t.Error("expected UAT marker line to be detected as marker")
	}
}

func TestIsMarkerLine_GapMarker(t *testing.T) {
	if !isMarkerLine(`{"_marker":"gaps-buffer","_format":"ndjson-v1"}`) {
		t.Error("expected Gap marker line to be detected as marker")
	}
}

func TestIsMarkerLine_DataEntry(t *testing.T) {
	entry := `{"id":"abc123","kind":"uat","title":"Test"}`
	if isMarkerLine(entry) {
		t.Error("expected data entry NOT to be detected as marker")
	}
}

func TestIsMarkerLine_MalformedJSON(t *testing.T) {
	if isMarkerLine("not json at all") {
		t.Error("expected malformed JSON NOT to be detected as marker")
	}
}
