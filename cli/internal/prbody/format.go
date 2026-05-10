package prbody

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	uatFilename = "uat-buffer.jsonl"
	gapFilename = "gaps-buffer.jsonl"
)

// RenderUAT reads <runDir>/uat-buffer.jsonl, skips the marker line, parses
// the remaining NDJSON lines as UAT entries, and renders each entry using the
// templates locked in Phase 1 Decision 4. Returns the concatenated Markdown
// content (no ## Manual QA Plan heading). Returns an empty string for a
// missing file, a zero-byte file, or a file containing only the marker line.
func RenderUAT(runDir string) (string, error) {
	entries, err := readUATEntries(filepath.Join(runDir, uatFilename))
	if err != nil {
		return "", err
	}
	if len(entries) == 0 {
		return "", nil
	}

	var parts []string
	for _, e := range entries {
		parts = append(parts, renderUATEntry(e))
	}
	return strings.Join(parts, "\n"), nil
}

// RenderGaps reads <runDir>/gaps-buffer.jsonl, skips the marker line, parses
// the remaining NDJSON lines as Gap entries, and renders each entry using the
// template locked in Phase 1 Decision 4. Returns the concatenated Markdown
// content (no ## Known Gaps heading). Returns an empty string for a missing
// file, a zero-byte file, or a file containing only the marker line.
func RenderGaps(runDir string) (string, error) {
	entries, err := readGapEntries(filepath.Join(runDir, gapFilename))
	if err != nil {
		return "", err
	}
	if len(entries) == 0 {
		return "", nil
	}

	var parts []string
	for _, e := range entries {
		parts = append(parts, renderGapEntry(e))
	}
	return strings.Join(parts, "\n"), nil
}

// renderUATEntry renders a single UAT entry into Markdown per Decision 4.
//
// Single-line verify (no \n): `- [ ] **<title>** — <verify>`
// Multi-line verify (\n present):
//
//	- [ ] **<title>**
//	  (blank separator — 2 spaces)
//	  <verify with each continuation line indented 2 spaces>
//	  (blank separator — 2 spaces)
//	  _Why manual:_ <why_manual>   (omitted if why_manual is empty)
func renderUATEntry(e UATEntry) string {
	if !strings.Contains(e.Verify, "\n") {
		// Single-line form.
		return "- [ ] **" + e.Title + "** — " + e.Verify
	}

	// Multi-line form: indent every line of verify by 2 spaces; blank lines
	// within verify become "  " (2 spaces, no content) per Decision 4 mandate
	// that blank separator lines must carry the 2-space indent to stay inside
	// the GFM list-item context.
	indentedVerify := indentBlock(e.Verify)

	var sb strings.Builder
	sb.WriteString("- [ ] **")
	sb.WriteString(e.Title)
	sb.WriteString("**\n")
	sb.WriteString("  \n")
	sb.WriteString(indentedVerify)

	if strings.TrimSpace(e.WhyManual) != "" {
		sb.WriteString("\n")
		sb.WriteString("  \n")
		sb.WriteString("  _Why manual:_ ")
		sb.WriteString(e.WhyManual)
	}

	return sb.String()
}

// renderGapEntry renders a single Gap entry into Markdown per Decision 4.
//
// Single-line detail (no \n): `- **<title>** — <detail>`
// Multi-line detail (\n present):
//
//	- **<title>** — <first line of detail>
//	  <continuation lines, each indented 2 spaces>
//	(blank line)
//	  _Follow-up:_ <followup>   (omitted if followup is empty)
//
// Blank lines within detail become "  " (2 spaces) to stay inside the GFM
// list-item context (same rule as multi-line UAT verify).
func renderGapEntry(e GapEntry) string {
	var sb strings.Builder
	sb.WriteString("- **")
	sb.WriteString(e.Title)
	sb.WriteString("** — ")

	if strings.Contains(e.Detail, "\n") {
		// Multi-line: first line goes inline after "— "; continuation lines
		// are each prefixed with 2 spaces. Blank lines within detail become
		// "  " (2 spaces) to stay inside the GFM list-item context.
		lines := strings.Split(e.Detail, "\n")
		sb.WriteString(lines[0])
		for _, l := range lines[1:] {
			sb.WriteString("\n")
			if strings.TrimSpace(l) == "" {
				sb.WriteString("  ")
			} else {
				sb.WriteString("  ")
				sb.WriteString(l)
			}
		}
	} else {
		sb.WriteString(e.Detail)
	}

	if strings.TrimSpace(e.Followup) != "" {
		sb.WriteString("\n\n")
		sb.WriteString("  _Follow-up:_ ")
		sb.WriteString(e.Followup)
	}

	return sb.String()
}

// indentBlock splits text on \n and prefixes each line with "  " (2 spaces).
// Blank lines (empty or whitespace-only) are rendered as "  " (2 spaces, no
// trailing content) so they stay inside the GFM list-item context.
// The result does NOT have a trailing newline; the caller appends as needed.
func indentBlock(text string) string {
	lines := strings.Split(text, "\n")
	out := make([]string, len(lines))
	for i, l := range lines {
		if strings.TrimSpace(l) == "" {
			out[i] = "  "
		} else {
			out[i] = "  " + l
		}
	}
	return strings.Join(out, "\n")
}

// readUATEntries reads the UAT buffer file at path and returns parsed entries.
// It skips the first line if it is a marker line (has a non-empty _marker field).
// Missing, zero-byte, or marker-only files all return an empty slice with no error.
func readUATEntries(path string) ([]UATEntry, error) {
	lines, err := readDataLines(path)
	if err != nil {
		return nil, err
	}

	var entries []UATEntry
	for _, line := range lines {
		var e UATEntry
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			// Skip malformed lines rather than crashing; the buffer is
			// append-only and a corrupt line should not block rendering.
			continue
		}
		entries = append(entries, e)
	}
	return entries, nil
}

// readGapEntries reads the Gap buffer file at path and returns parsed entries.
// See readUATEntries for the missing/empty/marker-only semantics.
func readGapEntries(path string) ([]GapEntry, error) {
	lines, err := readDataLines(path)
	if err != nil {
		return nil, err
	}

	var entries []GapEntry
	for _, line := range lines {
		var e GapEntry
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			continue
		}
		entries = append(entries, e)
	}
	return entries, nil
}

// readDataLines opens the file at path, reads all non-empty lines, and skips
// the first line if it is a marker line. Returns an empty slice (no error) if
// the file is missing or zero bytes.
func readDataLines(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	const maxScannerBytes = 1 << 20 // 1 MB
	var lines []string
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), maxScannerBytes)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		lines = append(lines, line)
	}
	if err := scanner.Err(); err != nil {
		if errors.Is(err, bufio.ErrTooLong) {
			return nil, fmt.Errorf("%s: line exceeds maximum scanner buffer of %d bytes", path, maxScannerBytes)
		}
		return nil, err
	}

	if len(lines) == 0 {
		return nil, nil
	}

	// Skip the first line if it is a marker line.
	if isMarkerLine(lines[0]) {
		lines = lines[1:]
	}

	return lines, nil
}

// isMarkerLine returns true if line is a JSON object with a non-empty _marker field.
// Defensive: if JSON parsing fails, the line is treated as a data line.
func isMarkerLine(line string) bool {
	var m markerLine
	if err := json.Unmarshal([]byte(line), &m); err != nil {
		return false
	}
	return m.Marker != ""
}
