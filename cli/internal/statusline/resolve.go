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

// taskClassification holds the three signals produced by a single scan of the
// tasks directory.  It is the output of classifyTasks.
type taskClassification struct {
	// ActiveCurrent is the absolute path of the lexicographically-largest
	// active task whose work branch matches the current branch, or "" if none.
	// This is the rung-1 winner.
	ActiveCurrent string

	// CompletedCurrent is the absolute path of the lexicographically-largest
	// completed task (status == "completed") whose work branch matches the
	// current branch, or "" if none.  This is the rung-2 winner.
	CompletedCurrent string

	// ActiveOtherCount is the number of active tasks whose work branch differs
	// from the current branch.  This is the rung-3 signal.
	ActiveOtherCount int
}

// classifyTasks performs a single os.ReadDir of the tasks directory and opens
// each .md file once via extractTaskMeta to classify it into one of the three
// rung buckets.  It returns a zero taskClassification on any setup failure
// (no dir, unreadable dir) so the caller silently produces no output.
func classifyTasks(ctxDir, branch string) taskClassification {
	tasksDir := filepath.Join(ctxDir, ".crafter", "tasks")
	entries, err := os.ReadDir(tasksDir)
	if err != nil {
		return taskClassification{}
	}

	var (
		activeCurrent    []string
		completedCurrent []string
		activeOtherCount int
	)

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".md") {
			continue
		}
		path := filepath.Join(tasksDir, name)
		meta := extractTaskMeta(path)

		switch {
		case meta.status == "active" && meta.workBranch == branch:
			activeCurrent = append(activeCurrent, path)
		case meta.status == "completed" && meta.workBranch == branch:
			completedCurrent = append(completedCurrent, path)
		case meta.status == "active" && meta.workBranch != branch && meta.workBranch != "":
			activeOtherCount++
		}
	}

	var result taskClassification
	result.ActiveOtherCount = activeOtherCount

	if len(activeCurrent) > 0 {
		sort.Strings(activeCurrent)
		result.ActiveCurrent = activeCurrent[len(activeCurrent)-1]
	}
	if len(completedCurrent) > 0 {
		sort.Strings(completedCurrent)
		result.CompletedCurrent = completedCurrent[len(completedCurrent)-1]
	}

	return result
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

// taskMeta holds the status and work-branch strings extracted from a task
// file's ## Metadata section.  Both fields are empty if the file cannot be
// read or has no ## Metadata section.
type taskMeta struct {
	status     string // e.g. "active", "completed", "abandoned"
	workBranch string // value of "- **Work branch:** <branch>"
}

// extractTaskMeta opens path once, scans its ## Metadata section, and returns
// the raw status and work-branch strings found there.  Scanning stops at the
// first ## heading after ## Metadata, or at EOF, whichever comes first —
// exactly the same section-boundary semantics as the previous
// isActiveTaskForBranch implementation.
//
// On any open or read error the function returns a zero taskMeta (both fields
// empty) so the caller silently treats the file as "no match", preserving the
// silent-fail posture of the statusline contract.  The scanner buffer is sized
// to 1 MB to handle large task files.
func extractTaskMeta(path string) taskMeta {
	f, err := os.Open(path)
	if err != nil {
		return taskMeta{}
	}
	defer f.Close()

	const statusPrefix = "- **Status:** "
	const branchPrefix = "- **Work branch:** "

	var (
		inMeta bool
		meta   taskMeta
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

		if meta.status == "" && strings.HasPrefix(line, statusPrefix) {
			meta.status = strings.TrimPrefix(line, statusPrefix)
		}
		if meta.workBranch == "" && strings.HasPrefix(line, branchPrefix) {
			meta.workBranch = strings.TrimPrefix(line, branchPrefix)
		}

		if meta.status != "" && meta.workBranch != "" {
			// Both fields found; no need to read further.
			break
		}
	}

	if err := scanner.Err(); err != nil && !errors.Is(err, bufio.ErrTooLong) {
		// ErrTooLong means the file is unusual; still return partial/zero meta
		// rather than panic.
		_ = err
	}

	return meta
}
