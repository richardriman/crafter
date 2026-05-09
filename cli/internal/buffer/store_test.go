package buffer

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func makeUATEntry(t *testing.T, title, verify string) UATEntry {
	t.Helper()
	id, err := NewID()
	if err != nil {
		t.Fatalf("NewID failed: %v", err)
	}
	return UATEntry{
		ID:        id,
		Kind:      "uat",
		CreatedAt: NowUTC(),
		CreatedBy: "test-agent",
		TaskID:    "test-task-id",
		Title:     title,
		Source:    "test:1",
		Verify:    verify,
		WhyManual: "testing",
	}
}

func makeGapEntry(t *testing.T, title, detail string) GapEntry {
	t.Helper()
	id, err := NewID()
	if err != nil {
		t.Fatalf("NewID failed: %v", err)
	}
	return GapEntry{
		ID:        id,
		Kind:      "gap",
		CreatedAt: NowUTC(),
		CreatedBy: "test-agent",
		TaskID:    "test-task-id",
		Title:     title,
		Source:    "test:1",
		Detail:    detail,
		Followup:  "fix it",
	}
}

// readLines returns all lines from a file, stripping the trailing newline from each.
func readLines(t *testing.T, path string) []string {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("opening %s: %v", path, err)
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scanning %s: %v", path, err)
	}
	return lines
}

func TestAppendUAT_CreatesFileWithMarkerAndEntry(t *testing.T) {
	dir := t.TempDir()
	entry := makeUATEntry(t, "Login check", "Click sign in.")

	if err := AppendUAT(dir, entry); err != nil {
		t.Fatalf("AppendUAT failed: %v", err)
	}

	path := filepath.Join(dir, "uat-buffer.jsonl")
	lines := readLines(t, path)

	if len(lines) != 2 {
		t.Fatalf("expected 2 lines (marker + entry), got %d", len(lines))
	}

	// First line must be the UAT marker.
	markerTrimmed := strings.TrimSuffix(UATMarkerLine, "\n")
	if lines[0] != markerTrimmed {
		t.Errorf("first line: got %q, want %q", lines[0], markerTrimmed)
	}

	// Second line must parse as a UATEntry.
	var got UATEntry
	if err := json.Unmarshal([]byte(lines[1]), &got); err != nil {
		t.Fatalf("parsing entry line: %v", err)
	}
	if got.Title != entry.Title {
		t.Errorf("title: got %q, want %q", got.Title, entry.Title)
	}
	if got.Kind != "uat" {
		t.Errorf("kind: got %q, want \"uat\"", got.Kind)
	}
}

func TestAppendGap_CreatesFileWithMarkerAndEntry(t *testing.T) {
	dir := t.TempDir()
	entry := makeGapEntry(t, "Missing rollback", "No down migration.")

	if err := AppendGap(dir, entry); err != nil {
		t.Fatalf("AppendGap failed: %v", err)
	}

	path := filepath.Join(dir, "gaps-buffer.jsonl")
	lines := readLines(t, path)

	if len(lines) != 2 {
		t.Fatalf("expected 2 lines (marker + entry), got %d", len(lines))
	}

	markerTrimmed := strings.TrimSuffix(GapMarkerLine, "\n")
	if lines[0] != markerTrimmed {
		t.Errorf("first line: got %q, want %q", lines[0], markerTrimmed)
	}

	var got GapEntry
	if err := json.Unmarshal([]byte(lines[1]), &got); err != nil {
		t.Fatalf("parsing entry line: %v", err)
	}
	if got.Title != entry.Title {
		t.Errorf("title: got %q, want %q", got.Title, entry.Title)
	}
	if got.Kind != "gap" {
		t.Errorf("kind: got %q, want \"gap\"", got.Kind)
	}
}

func TestAppendUAT_TwoSequentialAppends(t *testing.T) {
	dir := t.TempDir()
	e1 := makeUATEntry(t, "First check", "Step 1.")
	e2 := makeUATEntry(t, "Second check", "Step 2.")

	if err := AppendUAT(dir, e1); err != nil {
		t.Fatalf("first AppendUAT failed: %v", err)
	}
	if err := AppendUAT(dir, e2); err != nil {
		t.Fatalf("second AppendUAT failed: %v", err)
	}

	lines := readLines(t, filepath.Join(dir, "uat-buffer.jsonl"))
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines (marker + 2 entries), got %d", len(lines))
	}

	var got1, got2 UATEntry
	if err := json.Unmarshal([]byte(lines[1]), &got1); err != nil {
		t.Fatalf("parsing entry 1: %v", err)
	}
	if err := json.Unmarshal([]byte(lines[2]), &got2); err != nil {
		t.Fatalf("parsing entry 2: %v", err)
	}
	if got1.Title != e1.Title {
		t.Errorf("entry 1 title: got %q, want %q", got1.Title, e1.Title)
	}
	if got2.Title != e2.Title {
		t.Errorf("entry 2 title: got %q, want %q", got2.Title, e2.Title)
	}
}

func TestAppendUAT_CodeFenceContent(t *testing.T) {
	dir := t.TempDir()
	fencedVerify := "Run:\n\n```bash\nnpm test\n```\n\nConfirm all pass."
	entry := makeUATEntry(t, "Fenced code test", fencedVerify)

	if err := AppendUAT(dir, entry); err != nil {
		t.Fatalf("AppendUAT with fenced code failed: %v", err)
	}

	lines := readLines(t, filepath.Join(dir, "uat-buffer.jsonl"))
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}

	var got UATEntry
	if err := json.Unmarshal([]byte(lines[1]), &got); err != nil {
		t.Fatalf("parsing fenced entry: %v", err)
	}
	if got.Verify != fencedVerify {
		t.Errorf("verify round-trip failed:\n  got  %q\n  want %q", got.Verify, fencedVerify)
	}
}

func TestAppendGap_MultiLineContent(t *testing.T) {
	dir := t.TempDir()
	multiLineDetail := "Line one.\n\nLine three after blank.\nLine four."
	entry := makeGapEntry(t, "Multi-line gap", multiLineDetail)

	if err := AppendGap(dir, entry); err != nil {
		t.Fatalf("AppendGap with multi-line content failed: %v", err)
	}

	lines := readLines(t, filepath.Join(dir, "gaps-buffer.jsonl"))
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}

	var got GapEntry
	if err := json.Unmarshal([]byte(lines[1]), &got); err != nil {
		t.Fatalf("parsing multi-line entry: %v", err)
	}
	if got.Detail != multiLineDetail {
		t.Errorf("detail round-trip failed:\n  got  %q\n  want %q", got.Detail, multiLineDetail)
	}
}

func TestAppendUAT_OversizedEntry_ReturnsError(t *testing.T) {
	dir := t.TempDir()
	// Build an entry whose encoded size exceeds MaxEntryBytes.
	entry := UATEntry{
		ID:        "aabbcc112233",
		Kind:      "uat",
		CreatedAt: NowUTC(),
		CreatedBy: "test-agent",
		TaskID:    "test-task-id",
		Title:     "x",
		Source:    "x",
		Verify:    strings.Repeat("v", MaxEntryBytes),
		WhyManual: "x",
	}

	err := AppendUAT(dir, entry)
	if err == nil {
		t.Fatal("expected error for oversized entry, got nil")
	}
	if !strings.Contains(err.Error(), "too large") {
		t.Errorf("expected 'too large' in error, got: %v", err)
	}

	// The buffer file must not have been created (or if created must be empty)
	// because we fail before opening the file.
	path := filepath.Join(dir, "uat-buffer.jsonl")
	if _, statErr := os.Stat(path); statErr == nil {
		// File exists — that would be wrong because encode fails before open.
		t.Error("buffer file should not exist after encode failure")
	}
}

func TestAppendUAT_MissingRunDir_ReturnsError(t *testing.T) {
	nonExistentDir := filepath.Join(t.TempDir(), "does-not-exist")
	entry := makeUATEntry(t, "Title", "Verify.")

	err := AppendUAT(nonExistentDir, entry)
	if err == nil {
		t.Fatal("expected error when run-dir does not exist, got nil")
	}
}

func TestAppendGap_MissingRunDir_ReturnsError(t *testing.T) {
	nonExistentDir := filepath.Join(t.TempDir(), "does-not-exist")
	entry := makeGapEntry(t, "Title", "Detail.")

	err := AppendGap(nonExistentDir, entry)
	if err == nil {
		t.Fatal("expected error when run-dir does not exist, got nil")
	}
}
