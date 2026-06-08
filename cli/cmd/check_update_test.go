package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// --- semver compare ---

func TestCheckUpdateIsNewer(t *testing.T) {
	tests := []struct {
		latest    string
		installed string
		want      bool
	}{
		// Equal versions — not newer.
		{"1.0.0", "1.0.0", false},
		{"0.12.0", "0.12.0", false},

		// Latest is strictly newer.
		{"1.0.1", "1.0.0", true},
		{"1.1.0", "1.0.9", true},
		{"2.0.0", "1.9.9", true},

		// Installed is newer — not newer.
		{"1.0.0", "1.0.1", false},
		{"0.11.9", "0.12.0", false},

		// Differing component counts — pad with zeros.
		{"1.2", "1.2.0", false},  // equal after padding
		{"1.2.0", "1.2", false},  // equal after padding
		{"1.3", "1.2.0", true},   // 1.3.0 > 1.2.0
		{"1.2.0", "1.3", false},  // 1.2.0 < 1.3.0

		// v-prefix already stripped by caller, but test robustness of parser.
		{"0.0.1", "0.0.0", true},
		{"0.0.0", "0.0.1", false},
	}

	for _, tc := range tests {
		got := checkUpdateIsNewer(tc.latest, tc.installed)
		if got != tc.want {
			t.Errorf("isNewer(%q, %q) = %v; want %v", tc.latest, tc.installed, got, tc.want)
		}
	}
}

// --- cache read + notice printing ---

// makeTestEnv sets up a temporary HOME-like directory and overrides the
// package-level funcs to point at it. Returns (tempDir, restore).
func makeTestEnv(t *testing.T) (string, func()) {
	t.Helper()
	dir := t.TempDir()

	origCacheFile := checkUpdateCacheFile
	origVersionFiles := checkUpdateVersionFiles
	origNow := checkUpdateNow

	checkUpdateCacheFile = func() (string, error) {
		return filepath.Join(dir, ".claude", "cache", "crafter-update-check.json"), nil
	}
	checkUpdateVersionFiles = func() (string, string) {
		global := filepath.Join(dir, ".claude", "crafter", "VERSION")
		project := filepath.Join(dir, "project", ".claude", "crafter", "VERSION")
		return global, project
	}
	checkUpdateNow = func() int64 { return 1000000 } // fixed epoch

	restore := func() {
		checkUpdateCacheFile = origCacheFile
		checkUpdateVersionFiles = origVersionFiles
		checkUpdateNow = origNow
	}
	return dir, restore
}

// writeFile creates parent dirs and writes content to path.
func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

// TestCheckUpdate_NoVersion verifies: when neither VERSION file exists, the
// foreground path returns nil (exit 0) and prints nothing.
func TestCheckUpdate_NoVersion(t *testing.T) {
	_, restore := makeTestEnv(t)
	defer restore()

	var buf bytes.Buffer
	// No VERSION files written — installed version cannot be resolved.
	checkUpdatePrintNoticeIfAvailable("", &buf)
	// Separately, runCheckUpdate must also return nil.
	if err := runCheckUpdate(checkUpdateCmd, nil); err != nil {
		t.Errorf("runCheckUpdate with no VERSION: got non-nil error: %v", err)
	}
	if buf.Len() != 0 {
		t.Errorf("expected no output, got: %q", buf.String())
	}
}

// TestCheckUpdate_NoVersion_ReadInstalledVersion verifies the version resolver
// returns (_, false) when no VERSION files are present.
func TestCheckUpdate_NoVersion_ReadInstalledVersion(t *testing.T) {
	_, restore := makeTestEnv(t)
	defer restore()

	_, ok := checkUpdateReadInstalledVersion()
	if ok {
		t.Error("expected ok=false when no VERSION files exist, got true")
	}
}

// TestCheckUpdate_GlobalVersionTakesPrecedence verifies that the global
// VERSION file is preferred over the project-local one.
func TestCheckUpdate_GlobalVersionTakesPrecedence(t *testing.T) {
	dir, restore := makeTestEnv(t)
	defer restore()

	globalFile, projectFile := checkUpdateVersionFiles()
	writeFile(t, globalFile, "1.0.0\n")
	writeFile(t, projectFile, "2.0.0\n")

	_ = dir
	v, ok := checkUpdateReadInstalledVersion()
	if !ok {
		t.Fatal("expected ok=true")
	}
	if v != "1.0.0" {
		t.Errorf("expected global version 1.0.0, got %q", v)
	}
}

// TestCheckUpdate_ProjectVersionFallback verifies that when only the project
// VERSION file exists it is used.
func TestCheckUpdate_ProjectVersionFallback(t *testing.T) {
	_, restore := makeTestEnv(t)
	defer restore()

	_, projectFile := checkUpdateVersionFiles()
	writeFile(t, projectFile, "0.9.5\n")

	v, ok := checkUpdateReadInstalledVersion()
	if !ok {
		t.Fatal("expected ok=true")
	}
	if v != "0.9.5" {
		t.Errorf("expected project version 0.9.5, got %q", v)
	}
}

// TestCheckUpdate_CacheUpdateAvailable_MatchingInstalled verifies that when
// the cache says update_available=true and the installed field matches the
// current installed version, the notice is printed to stdout (byte-exact).
func TestCheckUpdate_CacheUpdateAvailable_MatchingInstalled(t *testing.T) {
	dir, restore := makeTestEnv(t)
	defer restore()

	globalFile, _ := checkUpdateVersionFiles()
	writeFile(t, globalFile, "0.12.0\n")

	cacheFile, _ := checkUpdateCacheFile()
	cache := updateCheckCache{
		UpdateAvailable: true,
		Installed:       "0.12.0",
		Latest:          "0.13.0",
		Checked:         999990,
	}
	data, _ := json.Marshal(cache)
	writeFile(t, cacheFile, string(data))

	_ = dir
	var buf bytes.Buffer
	checkUpdatePrintNoticeIfAvailable("0.12.0", &buf)

	want := "Note: Crafter update available (installed: 0.12.0, latest: 0.13.0). Run: curl -fsSL https://raw.githubusercontent.com/richardriman/crafter/main/install.sh | bash\n"
	if buf.String() != want {
		t.Errorf("notice mismatch:\ngot:  %q\nwant: %q", buf.String(), want)
	}
}

// TestCheckUpdate_CacheUpdateAvailable_InstalledMismatch verifies that when
// cache.installed does not match the current installed version, no notice is
// printed (the user already upgraded).
func TestCheckUpdate_CacheUpdateAvailable_InstalledMismatch(t *testing.T) {
	dir, restore := makeTestEnv(t)
	defer restore()

	cacheFile, _ := checkUpdateCacheFile()
	cache := updateCheckCache{
		UpdateAvailable: true,
		Installed:       "0.11.0", // stale — user is now on 0.12.0
		Latest:          "0.13.0",
		Checked:         999990,
	}
	data, _ := json.Marshal(cache)
	writeFile(t, cacheFile, string(data))

	_ = dir
	var buf bytes.Buffer
	checkUpdatePrintNoticeIfAvailable("0.12.0", &buf)

	if buf.Len() != 0 {
		t.Errorf("expected no output on installed mismatch, got: %q", buf.String())
	}
}

// TestCheckUpdate_CacheUpdateFalse verifies that when update_available=false
// no notice is printed even if installed matches.
func TestCheckUpdate_CacheUpdateFalse(t *testing.T) {
	dir, restore := makeTestEnv(t)
	defer restore()

	cacheFile, _ := checkUpdateCacheFile()
	cache := updateCheckCache{
		UpdateAvailable: false,
		Installed:       "0.12.0",
		Latest:          "0.12.0",
		Checked:         999990,
	}
	data, _ := json.Marshal(cache)
	writeFile(t, cacheFile, string(data))

	_ = dir
	var buf bytes.Buffer
	checkUpdatePrintNoticeIfAvailable("0.12.0", &buf)

	if buf.Len() != 0 {
		t.Errorf("expected no output when update_available=false, got: %q", buf.String())
	}
}

// TestCheckUpdate_CacheMissing verifies that a missing cache file produces no
// notice and does not panic or error.
func TestCheckUpdate_CacheMissing(t *testing.T) {
	_, restore := makeTestEnv(t)
	defer restore()
	// No cache file written.

	var buf bytes.Buffer
	checkUpdatePrintNoticeIfAvailable("0.12.0", &buf)

	if buf.Len() != 0 {
		t.Errorf("expected no output with missing cache, got: %q", buf.String())
	}
}

// TestCheckUpdate_CacheGarbledJSON verifies that a garbled cache file produces
// no panic, no notice, and no error (silent swallow).
func TestCheckUpdate_CacheGarbledJSON(t *testing.T) {
	_, restore := makeTestEnv(t)
	defer restore()

	cacheFile, _ := checkUpdateCacheFile()
	writeFile(t, cacheFile, "this is not json {{{{")

	var buf bytes.Buffer
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("checkUpdatePrintNoticeIfAvailable panicked on garbled JSON: %v", r)
		}
	}()
	checkUpdatePrintNoticeIfAvailable("0.12.0", &buf)

	if buf.Len() != 0 {
		t.Errorf("expected no output on garbled JSON, got: %q", buf.String())
	}
}

// TestCheckUpdate_RunCheckUpdate_NoVersionExitsZero verifies that the full
// runCheckUpdate foreground path exits 0 (nil error) when no VERSION exists.
func TestCheckUpdate_RunCheckUpdate_NoVersionExitsZero(t *testing.T) {
	_, restore := makeTestEnv(t)
	defer restore()

	checkUpdateRefresh = false
	if err := runCheckUpdate(checkUpdateCmd, nil); err != nil {
		t.Errorf("expected nil error, got: %v", err)
	}
}

// --- background refresh ---

// TestCheckUpdateDoRefresh_FreshnessSkip verifies that a fresh cache (checked
// within the 14400s window, same installed version) causes the refresh to skip
// the GitHub fetch and NOT overwrite the cache.
func TestCheckUpdateDoRefresh_FreshnessSkip(t *testing.T) {
	dir, restore := makeTestEnv(t)
	defer restore()

	globalFile, _ := checkUpdateVersionFiles()
	writeFile(t, globalFile, "0.12.0\n")

	cacheFile, _ := checkUpdateCacheFile()
	// checked = now - 100 (well within 14400s window)
	origCache := updateCheckCache{
		UpdateAvailable: false,
		Installed:       "0.12.0",
		Latest:          "0.12.0",
		Checked:         1000000 - 100,
	}
	data, _ := json.Marshal(origCache)
	writeFile(t, cacheFile, string(data))

	// Point GitHub base URL at a server that must NOT be called.
	origBase := checkUpdateGitHubBaseURL
	checkUpdateGitHubBaseURL = "http://127.0.0.1:0" // invalid — would fail if called
	defer func() { checkUpdateGitHubBaseURL = origBase }()

	_ = dir
	checkUpdateDoRefresh()

	// Cache must remain unchanged.
	raw, err := os.ReadFile(cacheFile)
	if err != nil {
		t.Fatalf("reading cache: %v", err)
	}
	var got updateCheckCache
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("parsing cache: %v", err)
	}
	if got.Checked != origCache.Checked {
		t.Errorf("cache was rewritten during freshness window: checked changed from %d to %d", origCache.Checked, got.Checked)
	}
}

// TestCheckUpdateDoRefresh_StaleCache verifies that a stale cache (checked
// beyond the 14400s window) triggers a GitHub fetch and rewrites the cache.
func TestCheckUpdateDoRefresh_StaleCache(t *testing.T) {
	dir, restore := makeTestEnv(t)
	defer restore()

	globalFile, _ := checkUpdateVersionFiles()
	writeFile(t, globalFile, "0.12.0\n")

	cacheFile, _ := checkUpdateCacheFile()
	// checked = now - 20000 (beyond 14400s)
	oldCache := updateCheckCache{
		UpdateAvailable: false,
		Installed:       "0.12.0",
		Latest:          "0.12.0",
		Checked:         1000000 - 20000,
	}
	data, _ := json.Marshal(oldCache)
	writeFile(t, cacheFile, string(data))

	// Serve a fake GitHub releases/latest response.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/releases/latest") {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"tag_name":"v0.13.0"}`))
	}))
	defer srv.Close()

	origBase := checkUpdateGitHubBaseURL
	checkUpdateGitHubBaseURL = srv.URL
	defer func() { checkUpdateGitHubBaseURL = origBase }()

	_ = dir
	checkUpdateDoRefresh()

	raw, err := os.ReadFile(cacheFile)
	if err != nil {
		t.Fatalf("reading cache after refresh: %v", err)
	}
	var got updateCheckCache
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("parsing cache after refresh: %v", err)
	}
	if got.Checked != 1000000 {
		t.Errorf("cache.checked: got %d, want 1000000", got.Checked)
	}
	if got.Installed != "0.12.0" {
		t.Errorf("cache.installed: got %q, want 0.12.0", got.Installed)
	}
	if got.Latest != "0.13.0" {
		t.Errorf("cache.latest: got %q, want 0.13.0", got.Latest)
	}
	if !got.UpdateAvailable {
		t.Errorf("cache.update_available: got false, want true (0.13.0 > 0.12.0)")
	}
}

// TestCheckUpdateDoRefresh_VersionChangedInvalidatesCache verifies that when
// installed version changed since the last cache write, the refresh runs even
// if the timestamp is within the freshness window.
func TestCheckUpdateDoRefresh_VersionChangedInvalidatesCache(t *testing.T) {
	dir, restore := makeTestEnv(t)
	defer restore()

	globalFile, _ := checkUpdateVersionFiles()
	writeFile(t, globalFile, "0.13.0\n") // upgraded since last cache

	cacheFile, _ := checkUpdateCacheFile()
	// Cache is fresh (100s ago) but for old version 0.12.0.
	oldCache := updateCheckCache{
		UpdateAvailable: false,
		Installed:       "0.12.0", // stale installed field
		Latest:          "0.13.0",
		Checked:         1000000 - 100,
	}
	data, _ := json.Marshal(oldCache)
	writeFile(t, cacheFile, string(data))

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"tag_name":"v0.13.0"}`))
	}))
	defer srv.Close()

	origBase := checkUpdateGitHubBaseURL
	checkUpdateGitHubBaseURL = srv.URL
	defer func() { checkUpdateGitHubBaseURL = origBase }()

	_ = dir
	checkUpdateDoRefresh()

	raw, err := os.ReadFile(cacheFile)
	if err != nil {
		t.Fatalf("reading cache: %v", err)
	}
	var got updateCheckCache
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("parsing cache: %v", err)
	}
	// Cache should be refreshed for the new installed version.
	if got.Installed != "0.13.0" {
		t.Errorf("cache.installed after version change: got %q, want 0.13.0", got.Installed)
	}
	if got.UpdateAvailable {
		t.Errorf("cache.update_available: got true, want false (0.13.0 == 0.13.0)")
	}
}

// TestCheckUpdateDoRefresh_GitHubFailure verifies that a GitHub fetch failure
// does not overwrite the cache or panic.
func TestCheckUpdateDoRefresh_GitHubFailure(t *testing.T) {
	dir, restore := makeTestEnv(t)
	defer restore()

	globalFile, _ := checkUpdateVersionFiles()
	writeFile(t, globalFile, "0.12.0\n")

	cacheFile, _ := checkUpdateCacheFile()
	// Stale cache so freshness check won't skip the fetch.
	oldCache := updateCheckCache{
		UpdateAvailable: false,
		Installed:       "0.12.0",
		Latest:          "0.12.0",
		Checked:         1000000 - 20000,
	}
	data, _ := json.Marshal(oldCache)
	writeFile(t, cacheFile, string(data))

	// Serve a 500 to simulate failure.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}))
	defer srv.Close()

	origBase := checkUpdateGitHubBaseURL
	checkUpdateGitHubBaseURL = srv.URL
	defer func() { checkUpdateGitHubBaseURL = origBase }()

	_ = dir
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("checkUpdateDoRefresh panicked on GitHub failure: %v", r)
		}
	}()
	checkUpdateDoRefresh()

	// Cache must remain unchanged (old checked timestamp).
	raw, err := os.ReadFile(cacheFile)
	if err != nil {
		t.Fatalf("reading cache: %v", err)
	}
	var got updateCheckCache
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("parsing cache: %v", err)
	}
	if got.Checked != oldCache.Checked {
		t.Errorf("cache was rewritten on GitHub failure: checked changed from %d to %d", oldCache.Checked, got.Checked)
	}
}

// TestCheckUpdateDoRefresh_VPrefixStripped verifies that a GitHub tag_name
// with a leading "v" is stripped before writing to the cache.
func TestCheckUpdateDoRefresh_VPrefixStripped(t *testing.T) {
	dir, restore := makeTestEnv(t)
	defer restore()

	globalFile, _ := checkUpdateVersionFiles()
	writeFile(t, globalFile, "0.12.0\n")

	// No cache file — starts missing, so freshness check won't skip.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"tag_name":"v1.0.0"}`))
	}))
	defer srv.Close()

	origBase := checkUpdateGitHubBaseURL
	checkUpdateGitHubBaseURL = srv.URL
	defer func() { checkUpdateGitHubBaseURL = origBase }()

	_ = dir
	checkUpdateDoRefresh()

	cacheFile, _ := checkUpdateCacheFile()
	raw, err := os.ReadFile(cacheFile)
	if err != nil {
		t.Fatalf("reading cache: %v", err)
	}
	var got updateCheckCache
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("parsing cache: %v", err)
	}
	if got.Latest != "1.0.0" {
		t.Errorf("v-prefix not stripped: got %q, want 1.0.0", got.Latest)
	}
}

// TestCheckUpdate_RefreshFlag_RunsDoRefresh verifies that when --refresh is
// set, runCheckUpdate calls the refresh path (not the foreground path).
// We do this by pointing GitHub at a test server and verifying the cache is
// rewritten (the foreground path never touches the cache).
func TestCheckUpdate_RefreshFlag_RunsDoRefresh(t *testing.T) {
	dir, restore := makeTestEnv(t)
	defer restore()

	globalFile, _ := checkUpdateVersionFiles()
	writeFile(t, globalFile, "0.12.0\n")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"tag_name":"v0.13.0"}`))
	}))
	defer srv.Close()

	origBase := checkUpdateGitHubBaseURL
	checkUpdateGitHubBaseURL = srv.URL
	defer func() { checkUpdateGitHubBaseURL = origBase }()

	_ = dir

	origRefresh := checkUpdateRefresh
	checkUpdateRefresh = true
	defer func() { checkUpdateRefresh = origRefresh }()

	if err := runCheckUpdate(checkUpdateCmd, nil); err != nil {
		t.Errorf("runCheckUpdate with --refresh: got non-nil error: %v", err)
	}

	cacheFile, _ := checkUpdateCacheFile()
	raw, err := os.ReadFile(cacheFile)
	if err != nil {
		t.Fatalf("cache not written by --refresh path: %v", err)
	}
	var got updateCheckCache
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("parsing cache: %v", err)
	}
	if got.Latest != "0.13.0" {
		t.Errorf("cache.latest: got %q, want 0.13.0", got.Latest)
	}
}
