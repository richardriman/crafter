package buffer

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestNewID_Format(t *testing.T) {
	id, err := NewID()
	if err != nil {
		t.Fatalf("NewID failed: %v", err)
	}
	if len(id) != 12 {
		t.Errorf("expected 12-char hex ID, got %q (len %d)", id, len(id))
	}
	for _, c := range id {
		if !strings.ContainsRune("0123456789abcdef", c) {
			t.Errorf("expected lowercase hex char, got %q in ID %q", c, id)
		}
	}
}

func TestNewID_Uniqueness(t *testing.T) {
	const n = 100
	seen := make(map[string]struct{}, n)
	for i := 0; i < n; i++ {
		id, err := NewID()
		if err != nil {
			t.Fatalf("NewID failed at iteration %d: %v", i, err)
		}
		if _, dup := seen[id]; dup {
			t.Errorf("duplicate ID %q generated", id)
		}
		seen[id] = struct{}{}
	}
}

func TestNowUTC_Parseable(t *testing.T) {
	ts := NowUTC()
	parsed, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		t.Fatalf("NowUTC returned unparseable timestamp %q: %v", ts, err)
	}
	if parsed.Location() != time.UTC {
		t.Errorf("expected UTC location, got %v", parsed.Location())
	}
	if !strings.HasSuffix(ts, "Z") {
		t.Errorf("expected timestamp to end with Z, got %q", ts)
	}
}

func TestEncodeUATEntry_RoundTrip(t *testing.T) {
	entry := UATEntry{
		ID:        "aabbcc112233",
		Kind:      "uat",
		CreatedAt: "2026-05-09T14:30:00Z",
		CreatedBy: "crafter-implementer",
		TaskID:    "20260509-feat-test",
		Title:     "Confirm login works",
		Source:    "auth/login.go:42",
		Verify:    "Click Sign In and confirm redirect to /dashboard.",
		WhyManual: "Requires a live session.",
	}

	line, err := EncodeUATEntry(entry)
	if err != nil {
		t.Fatalf("EncodeUATEntry failed: %v", err)
	}

	// Line must end with exactly one newline.
	if !strings.HasSuffix(string(line), "\n") {
		t.Error("encoded line must end with LF")
	}
	trimmed := strings.TrimSuffix(string(line), "\n")

	var got UATEntry
	if err := json.Unmarshal([]byte(trimmed), &got); err != nil {
		t.Fatalf("round-trip unmarshal failed: %v", err)
	}
	if got != entry {
		t.Errorf("round-trip mismatch:\n  got  %+v\n  want %+v", got, entry)
	}
}

func TestEncodeGapEntry_RoundTrip(t *testing.T) {
	entry := GapEntry{
		ID:        "112233aabbcc",
		Kind:      "gap",
		CreatedAt: "2026-05-09T14:31:00Z",
		CreatedBy: "crafter-reviewer",
		TaskID:    "20260509-feat-test",
		Title:     "Missing rollback migration",
		Source:    "db/migrations/0042.sql",
		Detail:    "No down migration exists.",
		Followup:  "Add a down migration.",
	}

	line, err := EncodeGapEntry(entry)
	if err != nil {
		t.Fatalf("EncodeGapEntry failed: %v", err)
	}

	trimmed := strings.TrimSuffix(string(line), "\n")
	var got GapEntry
	if err := json.Unmarshal([]byte(trimmed), &got); err != nil {
		t.Fatalf("round-trip unmarshal failed: %v", err)
	}
	if got != entry {
		t.Errorf("round-trip mismatch:\n  got  %+v\n  want %+v", got, entry)
	}
}

func TestEncodeUATEntry_CodeFenceContent(t *testing.T) {
	entry := UATEntry{
		ID:        "aabbcc112233",
		Kind:      "uat",
		CreatedAt: "2026-05-09T14:30:00Z",
		CreatedBy: "crafter-implementer",
		TaskID:    "20260509-feat-test",
		Title:     "Fenced code block in verify",
		Source:    "ui/preview.tsx:118",
		Verify:    "Run:\n\n```bash\nDATA=fixtures/empty.json npm run dev\n```\n\nHover and confirm.",
		WhyManual: "Requires a browser.",
	}

	line, err := EncodeUATEntry(entry)
	if err != nil {
		t.Fatalf("EncodeUATEntry failed: %v", err)
	}

	trimmed := strings.TrimSuffix(string(line), "\n")
	var got UATEntry
	if err := json.Unmarshal([]byte(trimmed), &got); err != nil {
		t.Fatalf("round-trip unmarshal of fenced-code-block entry failed: %v", err)
	}
	if got.Verify != entry.Verify {
		t.Errorf("verify field mismatch:\n  got  %q\n  want %q", got.Verify, entry.Verify)
	}
	// Backticks must survive the round-trip unescaped in JSON.
	if !strings.Contains(got.Verify, "```bash") {
		t.Error("expected backtick fence to survive round-trip")
	}
}

func TestEncodeUATEntry_SizeCap(t *testing.T) {
	// Build an entry whose encoded size will exceed MaxEntryBytes.
	entry := UATEntry{
		ID:        "aabbcc112233",
		Kind:      "uat",
		CreatedAt: "2026-05-09T14:30:00Z",
		CreatedBy: "crafter-implementer",
		TaskID:    "20260509-feat-test",
		Title:     "x",
		Source:    "x",
		Verify:    strings.Repeat("v", MaxEntryBytes),
		WhyManual: "x",
	}

	_, err := EncodeUATEntry(entry)
	if err == nil {
		t.Fatal("expected error for oversized entry, got nil")
	}
	if !strings.Contains(err.Error(), "too large") {
		t.Errorf("expected 'too large' in error, got: %v", err)
	}
}

func TestEncodeGapEntry_SizeCap(t *testing.T) {
	entry := GapEntry{
		ID:        "aabbcc112233",
		Kind:      "gap",
		CreatedAt: "2026-05-09T14:30:00Z",
		CreatedBy: "crafter-reviewer",
		TaskID:    "20260509-feat-test",
		Title:     "x",
		Source:    "x",
		Detail:    strings.Repeat("d", MaxEntryBytes),
		Followup:  "x",
	}

	_, err := EncodeGapEntry(entry)
	if err == nil {
		t.Fatal("expected error for oversized entry, got nil")
	}
	if !strings.Contains(err.Error(), "too large") {
		t.Errorf("expected 'too large' in error, got: %v", err)
	}
}

// TestEncodeUATEntry_ExactlyMaxBytesAccepted verifies that MaxEntryBytes is an
// inclusive upper bound: a line of exactly MaxEntryBytes bytes (including the
// trailing LF) must be accepted.
func TestEncodeUATEntry_ExactlyMaxBytesAccepted(t *testing.T) {
	// Derive the overhead by encoding with an empty verify field, then pad to
	// reach exactly MaxEntryBytes bytes total (including the trailing LF).
	base := UATEntry{
		ID:        "aabbcc112233",
		Kind:      "uat",
		CreatedAt: "2026-05-09T14:30:00Z",
		CreatedBy: "crafter-implementer",
		TaskID:    "20260509-feat-test",
		Title:     "x",
		Source:    "x",
		Verify:    "",
		WhyManual: "x",
	}
	baseline, err := EncodeUATEntry(base)
	if err != nil {
		t.Fatalf("baseline encode failed: %v", err)
	}
	// baseline already ends with \n; its length is the overhead with zero verify chars.
	overhead := len(baseline)
	pad := MaxEntryBytes - overhead
	if pad < 0 {
		t.Fatalf("baseline entry (%d bytes) already exceeds MaxEntryBytes (%d); test setup is wrong", overhead, MaxEntryBytes)
	}

	base.Verify = strings.Repeat("v", pad)
	line, err := EncodeUATEntry(base)
	if err != nil {
		t.Fatalf("expected exactly-max-bytes entry to be accepted, got error: %v", err)
	}
	if len(line) != MaxEntryBytes {
		t.Errorf("expected encoded line to be exactly %d bytes, got %d", MaxEntryBytes, len(line))
	}
}

// TestEncodeUATEntry_OneBeyondMaxBytesRejected verifies that an entry whose
// encoded line is MaxEntryBytes+1 bytes is rejected.
func TestEncodeUATEntry_OneBeyondMaxBytesRejected(t *testing.T) {
	base := UATEntry{
		ID:        "aabbcc112233",
		Kind:      "uat",
		CreatedAt: "2026-05-09T14:30:00Z",
		CreatedBy: "crafter-implementer",
		TaskID:    "20260509-feat-test",
		Title:     "x",
		Source:    "x",
		Verify:    "",
		WhyManual: "x",
	}
	baseline, err := EncodeUATEntry(base)
	if err != nil {
		t.Fatalf("baseline encode failed: %v", err)
	}
	overhead := len(baseline)
	// pad+1 pushes the encoded line to MaxEntryBytes+1 bytes.
	pad := MaxEntryBytes - overhead + 1
	if pad < 0 {
		t.Fatalf("baseline entry (%d bytes) already exceeds MaxEntryBytes (%d); test setup is wrong", overhead, MaxEntryBytes)
	}

	base.Verify = strings.Repeat("v", pad)
	_, err = EncodeUATEntry(base)
	if err == nil {
		t.Fatal("expected error for entry one byte over MaxEntryBytes, got nil")
	}
	if !strings.Contains(err.Error(), "too large") {
		t.Errorf("expected 'too large' in error, got: %v", err)
	}
}
