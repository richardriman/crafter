package buffer

import (
	"fmt"
	"os"
	"path/filepath"
)

// MaxEntryBytes is the per-entry size cap (including the trailing LF).
//
// Decision (Implementer Recorded — Phase 2 Step 2): entry size cap — 512 bytes.
// POSIX guarantees write(2) to an O_APPEND file is atomic up to PIPE_BUF.
// The conservative bound is 512 bytes on macOS (per POSIX minimum). Linux is
// 4096 bytes but we target the smallest common denominator so the guarantee
// holds on both platforms without requiring platform-specific caps. Entries
// that exceed 512 bytes are rejected with a clear error; the caller should
// split long content across multiple entries or abbreviate the value.
//
// Note: marker lines are 47–48 bytes (UAT/Gap marker respectively) and are
// well within the cap; only data entries are capped.
const MaxEntryBytes = 512

const uatFilename = "uat-buffer.jsonl"
const gapFilename = "gaps-buffer.jsonl"

// AppendUAT appends a single NDJSON line for entry to <runDir>/uat-buffer.jsonl.
//
// If the file does not exist it is created and the UAT marker line is written
// first. If the file exists but is empty the marker line is written first.
//
// The runDir directory must already exist; this function does not create it.
// Decision (Implementer Recorded — Phase 2 Step 2): missing run-dir behavior —
// error out; do NOT auto-create. Auto-creating the directory tree is a Phase 3
// lifecycle concern and should not happen silently inside the append call.
//
// The file is opened with O_APPEND|O_WRONLY|O_CREAT and a single write(2) is
// issued for the NDJSON line to keep the POSIX atomicity guarantee.
func AppendUAT(runDir string, entry UATEntry) error {
	return appendEntry(
		filepath.Join(runDir, uatFilename),
		UATMarkerLine,
		func() ([]byte, error) { return EncodeUATEntry(entry) },
	)
}

// AppendGap appends a single NDJSON line for entry to <runDir>/gaps-buffer.jsonl.
// See AppendUAT for the full behavioural contract.
func AppendGap(runDir string, entry GapEntry) error {
	return appendEntry(
		filepath.Join(runDir, gapFilename),
		GapMarkerLine,
		func() ([]byte, error) { return EncodeGapEntry(entry) },
	)
}

// appendEntry handles the common open-stat-write sequence for both buffer kinds.
func appendEntry(path, markerLine string, encode func() ([]byte, error)) error {
	// Encode before opening the file so we fail fast on size violations.
	// entry is MaxEntryBytes at most (512 bytes).
	line, err := encode()
	if err != nil {
		return err
	}

	// Open with O_APPEND|O_WRONLY|O_CREAT so a single write(2) is atomic on
	// POSIX local filesystems up to MaxEntryBytes (PIPE_BUF conservative bound).
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o644)
	if err != nil {
		return fmt.Errorf("opening buffer file %s: %w", path, err)
	}
	defer f.Close()

	// First-write path: write marker + first entry as a single write(2) syscall
	// so there is no window in which the file exists with only the marker but
	// no entry. Combined payload is at most ≤48 bytes (marker including LF) +
	// ≤512 bytes (entry including LF) = ≤560 bytes, well below the 4 KB
	// page-cache granularity that Linux ext4/XFS and macOS APFS rely on for
	// atomic regular-file writes. Steady-state entry-only writes stay under
	// the 512-byte POSIX PIPE_BUF bound.
	info, err := f.Stat()
	if err != nil {
		return fmt.Errorf("stat buffer file %s: %w", path, err)
	}
	if info.Size() == 0 {
		combined := append([]byte(markerLine), line...)
		if _, err := f.Write(combined); err != nil {
			return fmt.Errorf("appending to %s: %w", path, err)
		}
		return nil
	}

	// Append the entry as a single write call.
	if _, err := f.Write(line); err != nil {
		return fmt.Errorf("appending to %s: %w", path, err)
	}
	return nil
}
