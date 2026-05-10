package buffer

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"time"
)

// UATMarkerLine is the fixed first line written to a new uat-buffer.jsonl file.
// Future tooling (PR composer, debug tools) can detect the file kind by reading
// this line without inspecting the filename.
const UATMarkerLine = `{"_marker":"uat-buffer","_format":"ndjson-v1"}` + "\n"

// GapMarkerLine is the fixed first line written to a new gaps-buffer.jsonl file.
const GapMarkerLine = `{"_marker":"gaps-buffer","_format":"ndjson-v1"}` + "\n"

// NewID generates a 12-character lowercase hex ID using crypto/rand.
// Decision (Implementer Recorded — Phase 2 Step 2): ID generation scheme —
// 12-char hex from crypto/rand. Chosen for simplicity: no new dependencies
// (crypto/rand is in the standard library), sufficient uniqueness for a
// per-run append-only log (2^48 ≈ 281 trillion distinct values), and matches
// the existing skillbook.NewID pattern (which uses 8 bytes / 16 chars; we use
// 6 bytes / 12 chars to keep entries shorter).
func NewID() (string, error) {
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generating id: %w", err)
	}
	return fmt.Sprintf("%x", b), nil
}

// NowUTC returns the current time formatted as ISO 8601 UTC (suffix Z).
func NowUTC() string {
	return time.Now().UTC().Format(time.RFC3339)
}

// EncodeUATEntry encodes entry as a single-line JSON object followed by a LF.
// Returns an error if the encoded line (including the trailing newline) exceeds
// MaxEntryBytes, since POSIX atomicity of O_APPEND write(2) is guaranteed only
// up to PIPE_BUF (512 bytes on macOS, 4096 on Linux). We use the conservative
// macOS bound.
func EncodeUATEntry(e UATEntry) ([]byte, error) {
	data, err := json.Marshal(e)
	if err != nil {
		return nil, fmt.Errorf("encoding UAT entry: %w", err)
	}
	line := append(data, '\n')
	if len(line) > MaxEntryBytes {
		return nil, fmt.Errorf(
			"UAT entry too large: %d bytes (max %d); shorten --verify or --why-manual",
			len(line), MaxEntryBytes,
		)
	}
	return line, nil
}

// EncodeGapEntry encodes entry as a single-line JSON object followed by a LF.
// Returns an error if the encoded line exceeds MaxEntryBytes (see EncodeUATEntry).
func EncodeGapEntry(e GapEntry) ([]byte, error) {
	data, err := json.Marshal(e)
	if err != nil {
		return nil, fmt.Errorf("encoding Gap entry: %w", err)
	}
	line := append(data, '\n')
	if len(line) > MaxEntryBytes {
		return nil, fmt.Errorf(
			"Gap entry too large: %d bytes (max %d); shorten --detail or --followup",
			len(line), MaxEntryBytes,
		)
	}
	return line, nil
}
