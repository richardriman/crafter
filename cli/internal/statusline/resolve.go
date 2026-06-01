package statusline

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const maxScannerBytes = 1 << 20 // 1 MB — task files can reach 50–80 KB

// taskMatch holds the result of a successful active-task resolution.
// It is the internal contract that Render (and Step 1.3) consume.
type taskMatch struct {
	// Path is the absolute path to the matching task file.
	Path string
}

// resolveActiveTask locates the active Crafter task file for the given working
// directory.  It:
//  1. Walks up from workdir to find a .crafter/ context directory (bounded by
//     the filesystem root and the user's home directory).
//  2. Reads the current git branch from the nearest .git/HEAD (walking up from
//     workdir); returns nil if there is no repo or the HEAD is detached.
//  3. Scans <ctxdir>/.crafter/tasks/*.md for a file whose ## Metadata section
//     contains BOTH "- **Status:** active" and "- **Work branch:** <branch>".
//  4. When multiple files match, returns the one with the lexicographically
//     largest filename (i.e. the most-recent date prefix).
//
// Returns nil (never an error) on any failure so the caller can silently
// produce no output — preserving the silent-fail posture of the statusline
// contract.
func resolveActiveTask(workdir string) *taskMatch {
	ctxDir := findCrafterDir(workdir)
	if ctxDir == "" {
		return nil
	}

	branch := readGitBranch(workdir)
	if branch == "" {
		return nil
	}

	tasksDir := filepath.Join(ctxDir, ".crafter", "tasks")
	entries, err := os.ReadDir(tasksDir)
	if err != nil {
		return nil
	}

	var matches []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".md") {
			continue
		}
		path := filepath.Join(tasksDir, name)
		if isActiveTaskForBranch(path, branch) {
			matches = append(matches, path)
		}
	}

	if len(matches) == 0 {
		return nil
	}

	// Tie-break: most recent by filename (date-prefix sort is lexicographic).
	sort.Strings(matches)
	return &taskMatch{Path: matches[len(matches)-1]}
}

// findCrafterDir walks up from dir looking for a .crafter/ subdirectory.
// The walk stops at the filesystem root or the user's home directory,
// whichever comes first (bounded walk — no unbounded ascent).
func findCrafterDir(dir string) string {
	home, _ := os.UserHomeDir() // empty string if unavailable — boundary check skipped gracefully

	for {
		candidate := filepath.Join(dir, ".crafter")
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root.
			return ""
		}

		// Stop when we have ascended past the home directory.
		if home != "" && dir == home {
			return ""
		}

		dir = parent
	}
}

// readGitBranch resolves the current git branch by reading .git/HEAD directly,
// walking up from dir to find the .git directory or file.
// Returns "" on detached HEAD, no repository, or any read error.
func readGitBranch(dir string) string {
	gitDir := findGitDir(dir)
	if gitDir == "" {
		return ""
	}

	headPath := filepath.Join(gitDir, "HEAD")
	data, err := os.ReadFile(headPath)
	if err != nil {
		return ""
	}

	// Expected format: "ref: refs/heads/<branch>\n"
	line := strings.TrimSpace(string(data))
	const prefix = "ref: refs/heads/"
	if !strings.HasPrefix(line, prefix) {
		// Detached HEAD (bare SHA) or unexpected format.
		return ""
	}
	return strings.TrimPrefix(line, prefix)
}

// findGitDir walks up from dir to find the .git directory or file.
// Returns the path to the .git entry (not the parent), or "" if not found.
func findGitDir(dir string) string {
	for {
		candidate := filepath.Join(dir, ".git")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

// isActiveTaskForBranch reports whether the task file at path has both
// "- **Status:** active" and "- **Work branch:** <branch>" in its ## Metadata
// section.
//
// Scanning is limited to the ## Metadata section only (stops at the next ## or
// EOF) to avoid false matches deeper in the file.  The scanner buffer is sized
// to 1 MB to handle large task files.
func isActiveTaskForBranch(path, branch string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	wantBranch := "- **Work branch:** " + branch
	wantStatus := "- **Status:** active"

	var (
		inMeta      bool
		foundStatus bool
		foundBranch bool
	)

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), maxScannerBytes)

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "## ") {
			if inMeta {
				// Leaving the Metadata section — stop early.
				break
			}
			if line == "## Metadata" {
				inMeta = true
			}
			continue
		}

		if !inMeta {
			continue
		}

		if line == wantStatus {
			foundStatus = true
		}
		if line == wantBranch {
			foundBranch = true
		}

		if foundStatus && foundBranch {
			return true
		}
	}

	if err := scanner.Err(); err != nil && !errors.Is(err, bufio.ErrTooLong) {
		// ErrTooLong means the file is unusual; still return false rather than panic.
		_ = err
	}

	return foundStatus && foundBranch
}
