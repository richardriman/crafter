package prbody

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
)

// ExtractDecisions reads the file at taskFilePath, locates the first
// "## Decisions" heading that is NOT inside a fenced code block (delimited
// by lines starting with ``` or ~~~), captures all content until the next
// "## " heading at the same level (or EOF), trims leading and trailing blank
// lines, and returns the result.
//
// Returns ("", nil) when:
//   - the file does not exist or cannot be read → returns ("", err) instead
//   - no "## Decisions" heading is found outside a code fence
//   - the section exists but contains only blank lines or HTML comments
//
// Edge cases guaranteed:
//
//	(i)  "## Decisions" followed immediately by another "## " heading →
//	     empty content → returns ""
//	(ii) "### Decision (...)" H3 sub-headings inside the section →
//	     preserved verbatim (H3 lines do NOT end the section)
//	(iii) "## Decisions" at end of file with no following "## " heading →
//	     consumes to EOF
//	(iv) Content between heading and next "## " is blank/whitespace/HTML
//	     comments only → returns ""
func ExtractDecisions(taskFilePath string) (string, error) {
	f, err := os.Open(taskFilePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	var (
		inFence    bool
		inSection  bool
		bodyLines  []string
	)

	const maxScannerBytes = 1 << 20 // 1 MB
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), maxScannerBytes)
	for scanner.Scan() {
		line := scanner.Text()

		// Track fenced code blocks (``` or ~~~ at line start, optional language).
		if isFenceDelimiter(line) {
			inFence = !inFence
			if inSection {
				bodyLines = append(bodyLines, line)
			}
			continue
		}

		if inFence {
			// Inside a code fence: never treat as a heading boundary.
			if inSection {
				bodyLines = append(bodyLines, line)
			}
			continue
		}

		// Outside a code fence: check for H2 headings.
		if strings.HasPrefix(line, "## ") {
			if inSection {
				// Hit the next H2 boundary — stop collecting.
				break
			}
			if line == "## Decisions" {
				// Exact match: heading text is exactly "Decisions" with no suffix.
				inSection = true
			}
			continue
		}

		if inSection {
			bodyLines = append(bodyLines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		if errors.Is(err, bufio.ErrTooLong) {
			return "", fmt.Errorf("%s: line exceeds maximum scanner buffer of %d bytes", taskFilePath, maxScannerBytes)
		}
		return "", err
	}

	return trimDecisionsBody(bodyLines), nil
}

// isFenceDelimiter reports whether line is a fenced code block delimiter.
// A delimiter is a line whose first three non-empty characters are all ` or ~,
// i.e. the line starts with ``` or ~~~ (optionally followed by a language tag).
func isFenceDelimiter(line string) bool {
	if strings.HasPrefix(line, "```") || strings.HasPrefix(line, "~~~") {
		return true
	}
	return false
}

// trimDecisionsBody trims leading and trailing blank lines from the collected
// body lines and joins them. A blank line is one that is empty or whitespace-
// only or is an HTML comment (<!-- ... -->). Returns "" if, after trimming,
// all remaining lines are blank/comment-only.
func trimDecisionsBody(lines []string) string {
	// Trim leading blank lines.
	for len(lines) > 0 && isBlankOrComment(lines[0]) {
		lines = lines[1:]
	}
	// Trim trailing blank lines.
	for len(lines) > 0 && isBlankOrComment(lines[len(lines)-1]) {
		lines = lines[:len(lines)-1]
	}
	if len(lines) == 0 {
		return ""
	}

	// Check whether ALL remaining lines are blank or comment-only.
	allBlank := true
	for _, l := range lines {
		if !isBlankOrComment(l) {
			allBlank = false
			break
		}
	}
	if allBlank {
		return ""
	}

	return strings.Join(lines, "\n")
}

// isBlankOrComment returns true when line is empty, whitespace-only, or an
// HTML comment (starts with <!-- and ends with -->).
func isBlankOrComment(line string) bool {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return true
	}
	if strings.HasPrefix(trimmed, "<!--") && strings.HasSuffix(trimmed, "-->") {
		return true
	}
	return false
}
