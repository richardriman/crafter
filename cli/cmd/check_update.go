package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

// checkUpdateGitHubBaseURL is the base URL for the GitHub API. It is a
// package-level var so tests can point it at an httptest.Server without real
// network access.
var checkUpdateGitHubBaseURL = "https://api.github.com"

// checkUpdateNow returns the current Unix timestamp in seconds. It is a
// package-level var so tests can inject a deterministic clock.
var checkUpdateNow = func() int64 {
	return time.Now().Unix()
}

// checkUpdateCacheFile returns the path to the update-check cache file.
// It is a package-level var so tests can redirect it to a temp dir.
var checkUpdateCacheFile = func() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".claude", "cache", "crafter-update-check.json"), nil
}

// checkUpdateVersionFiles returns (globalVersionFile, projectVersionFile).
// It is a package-level var so tests can inject temp paths.
var checkUpdateVersionFiles = func() (string, string) {
	home, _ := os.UserHomeDir()
	cwd, _ := os.Getwd()
	global := filepath.Join(home, ".claude", "crafter", "VERSION")
	project := filepath.Join(cwd, ".claude", "crafter", "VERSION")
	return global, project
}

// updateCheckCache mirrors the JSON structure of the cache file. Field names
// must remain byte-compatible with the JS hook.
type updateCheckCache struct {
	UpdateAvailable bool   `json:"update_available"`
	Installed       string `json:"installed"`
	Latest          string `json:"latest"`
	Checked         int64  `json:"checked"`
}

const checkUpdateFreshnessWindow = 14400 // 4 hours in seconds

var checkUpdateRefresh bool

var checkUpdateCmd = &cobra.Command{
	Use:          "check-update",
	Short:        "Check for Crafter updates and print a notice when one is available",
	SilenceUsage: true,
	RunE:         runCheckUpdate,
}

func init() {
	checkUpdateCmd.Flags().BoolVar(&checkUpdateRefresh, "refresh", false, "perform the background GitHub refresh (internal use)")
	_ = checkUpdateCmd.Flags().MarkHidden("refresh")
	rootCmd.AddCommand(checkUpdateCmd)
}

func runCheckUpdate(cmd *cobra.Command, args []string) error {
	if checkUpdateRefresh {
		// Background refresh path: fetch GitHub and rewrite the cache.
		// All errors are swallowed; never exit non-zero.
		checkUpdateDoRefresh()
		return nil
	}

	// --- Foreground path ---

	// Resolve installed version. Global takes precedence over project.
	installedVersion, ok := checkUpdateReadInstalledVersion()
	if !ok {
		// Crafter is not installed — exit 0 silently.
		return nil
	}

	// Read cache and print notice if an update is available.
	checkUpdatePrintNoticeIfAvailable(installedVersion, os.Stdout)

	// Spawn detached background refresh.
	checkUpdateSpawnRefresh()

	return nil
}

// checkUpdateReadInstalledVersion resolves the installed VERSION string using
// the same precedence as the JS hook: global first, project second.
// Returns (version, true) when found, ("", false) when neither file exists.
func checkUpdateReadInstalledVersion() (string, bool) {
	globalFile, projectFile := checkUpdateVersionFiles()

	for _, f := range []string{globalFile, projectFile} {
		data, err := os.ReadFile(f)
		if err == nil {
			v := strings.TrimSpace(string(data))
			if v != "" {
				return v, true
			}
		}
	}
	return "", false
}

// checkUpdatePrintNoticeIfAvailable reads the cache and writes the update
// notice to w when the cache says an update is available and the installed
// version matches the cached one. All errors are swallowed silently.
func checkUpdatePrintNoticeIfAvailable(installedVersion string, w io.Writer) {
	cacheFile, err := checkUpdateCacheFile()
	if err != nil {
		return
	}

	raw, err := os.ReadFile(cacheFile)
	if err != nil {
		return
	}

	var cache updateCheckCache
	if err := json.Unmarshal(raw, &cache); err != nil {
		return
	}

	if cache.UpdateAvailable && installedVersion == cache.Installed {
		fmt.Fprintf(w,
			"Note: Crafter update available (installed: %s, latest: %s). Run: curl -fsSL https://raw.githubusercontent.com/richardriman/crafter/main/install.sh | bash\n",
			cache.Installed, cache.Latest,
		)
	}
}

// checkUpdateSpawnRefresh launches a detached self-invocation with --refresh.
// It mirrors the JS spawn(..., {detached:true}); child.unref() pattern.
// If the self-path cannot be resolved, the refresh is skipped silently.
func checkUpdateSpawnRefresh() {
	self, err := os.Executable()
	if err != nil {
		return
	}

	c := exec.Command(self, "check-update", "--refresh")
	c.Stdin = nil
	c.Stdout = nil
	c.Stderr = nil
	c.SysProcAttr = &syscall.SysProcAttr{Setsid: true}

	if err := c.Start(); err != nil {
		return
	}
	// Release the child so it outlives this process (mirrors child.unref()).
	_ = c.Process.Release()
}

// checkUpdateDoRefresh fetches the latest release from GitHub, compares it
// against the installed version, and rewrites the cache file. All errors are
// swallowed; the function never returns an error.
func checkUpdateDoRefresh() {
	cacheFile, err := checkUpdateCacheFile()
	if err != nil {
		return
	}

	globalFile, projectFile := checkUpdateVersionFiles()
	installed := "0.0.0"
	for _, f := range []string{globalFile, projectFile} {
		data, err := os.ReadFile(f)
		if err == nil {
			v := strings.TrimSpace(string(data))
			if v != "" {
				installed = v
				break
			}
		}
	}

	now := checkUpdateNow()

	// Check cache freshness: skip GitHub call if checked within the window and
	// the installed version has not changed since the last write.
	if raw, err := os.ReadFile(cacheFile); err == nil {
		var cache updateCheckCache
		if err := json.Unmarshal(raw, &cache); err == nil {
			if cache.Checked > 0 && (now-cache.Checked) < checkUpdateFreshnessWindow && cache.Installed == installed {
				return
			}
		}
	}

	// Fetch latest release tag from GitHub.
	latest, ok := checkUpdateFetchLatest()
	if !ok {
		return
	}

	// Write updated cache.
	result := updateCheckCache{
		UpdateAvailable: checkUpdateIsNewer(latest, installed),
		Installed:       installed,
		Latest:          latest,
		Checked:         now,
	}

	// Ensure cache directory exists.
	_ = os.MkdirAll(filepath.Dir(cacheFile), 0o755)

	data, err := json.Marshal(result)
	if err != nil {
		return
	}
	_ = os.WriteFile(cacheFile, data, 0o644)
}

// checkUpdateFetchLatest fetches the latest release tag from GitHub and strips
// a leading "v". Returns (tag, true) on success, ("", false) on any failure.
func checkUpdateFetchLatest() (string, bool) {
	url := checkUpdateGitHubBaseURL + "/repos/richardriman/crafter/releases/latest"
	client := &http.Client{Timeout: 5 * time.Second}

	resp, err := client.Get(url)
	if err != nil {
		return "", false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", false
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", false
	}

	var payload struct {
		TagName string `json:"tag_name"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", false
	}
	if payload.TagName == "" {
		return "", false
	}

	tag := strings.TrimPrefix(payload.TagName, "v")
	return tag, true
}

// checkUpdateIsNewer returns true only when latest is strictly newer than
// installed, using the same component-wise comparison as the JS hook: shorter
// versions are padded with zeros on the right.
func checkUpdateIsNewer(latest, installed string) bool {
	a := checkUpdateParseVersion(latest)
	b := checkUpdateParseVersion(installed)
	n := len(a)
	if len(b) > n {
		n = len(b)
	}
	for i := 0; i < n; i++ {
		av := 0
		if i < len(a) {
			av = a[i]
		}
		bv := 0
		if i < len(b) {
			bv = b[i]
		}
		if av > bv {
			return true
		}
		if av < bv {
			return false
		}
	}
	return false
}

// checkUpdateParseVersion splits a dot-separated version string into integer
// components. Non-numeric components parse as 0.
func checkUpdateParseVersion(v string) []int {
	parts := strings.Split(v, ".")
	nums := make([]int, 0, len(parts))
	for _, p := range parts {
		n := 0
		for _, c := range p {
			if c >= '0' && c <= '9' {
				n = n*10 + int(c-'0')
			} else {
				break
			}
		}
		nums = append(nums, n)
	}
	return nums
}
