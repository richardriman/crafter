package claudesettings

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestNew_Empty(t *testing.T) {
	s := New()
	if s == nil {
		t.Fatal("expected non-nil settings")
	}
	if s.Has("anything") {
		t.Error("expected no keys in a new settings object")
	}
}

func TestLoad_NonExistentFile_ReturnsEmptySettings(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "does_not_exist.json")

	s, err := Load(path)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if s == nil {
		t.Fatal("expected non-nil settings")
	}
	if s.Has("model") {
		t.Error("expected empty settings for a missing file")
	}
}

func TestLoad_InvalidJSON_ReturnsEmptySettings(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")
	if err := os.WriteFile(path, []byte("{ this is not valid json"), 0o644); err != nil {
		t.Fatalf("setup write failed: %v", err)
	}

	s, err := Load(path)
	if err != nil {
		t.Fatalf("expected no hard error on invalid JSON, got %v", err)
	}
	if s == nil {
		t.Fatal("expected non-nil settings")
	}
	if s.Has("model") {
		t.Error("expected empty settings when content is garbled JSON")
	}
}

func TestSaveAndLoad_PreservesUnknownKeys(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")

	// Seed a file with several arbitrary keys, including nested structures
	// the package has no knowledge of.
	seed := `{
  "model": "claude-opus",
  "permissions": { "allow": ["Read", "Write"] },
  "hooks": { "SessionStart": [ { "hooks": [ { "type": "command", "command": "x" } ] } ] },
  "env": { "FOO": "bar" }
}`
	if err := os.WriteFile(path, []byte(seed), 0o644); err != nil {
		t.Fatalf("setup write failed: %v", err)
	}

	s, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Touch exactly one new key; everything else must round-trip untouched.
	if err := s.Set("statusLine", map[string]string{"type": "command", "command": "crafter statusline"}); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	if err := Save(path, s); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	reloaded, err := Load(path)
	if err != nil {
		t.Fatalf("reload failed: %v", err)
	}

	// Untouched keys must be byte-identical to what was seeded.
	wantUntouched := map[string]string{
		"model":       `"claude-opus"`,
		"permissions": `{"allow":["Read","Write"]}`,
		"hooks":       `{"SessionStart":[{"hooks":[{"type":"command","command":"x"}]}]}`,
		"env":         `{"FOO":"bar"}`,
	}
	for key, want := range wantUntouched {
		raw, ok := reloaded.Get(key)
		if !ok {
			t.Fatalf("expected key %q to be preserved on round-trip", key)
		}
		if got := compactJSON(t, raw); got != want {
			t.Errorf("key %q: expected %s, got %s", key, want, got)
		}
	}

	// The touched key must be present with the new value.
	raw, ok := reloaded.Get("statusLine")
	if !ok {
		t.Fatal("expected statusLine key to be present after Set+Save")
	}
	if got, want := compactJSON(t, raw), `{"command":"crafter statusline","type":"command"}`; got != want {
		t.Errorf("statusLine: expected %s, got %s", want, got)
	}
}

func TestMarshal_ByteShape(t *testing.T) {
	s := New()
	if err := s.Set("model", "claude-opus"); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	data, err := s.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Mirrors JSON.stringify(settings, null, 2) + "\n".
	want := "{\n  \"model\": \"claude-opus\"\n}\n"
	if string(data) != want {
		t.Errorf("unexpected byte shape:\n  got:  %q\n  want: %q", string(data), want)
	}

	if !bytes.HasSuffix(data, []byte("\n")) {
		t.Error("expected output to end with a trailing newline")
	}
}

func TestSave_TmpFileCleanedUp(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")

	s := New()
	if err := Save(path, s); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	tmpPath := path + ".tmp"
	if _, err := os.Stat(tmpPath); !os.IsNotExist(err) {
		t.Error("expected .tmp file to be removed after Save, but it still exists")
	}
}

func TestBackup_CreatesBak(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")
	original := []byte(`{"model":"original"}` + "\n")
	if err := os.WriteFile(path, original, 0o644); err != nil {
		t.Fatalf("setup write failed: %v", err)
	}

	if err := Backup(path); err != nil {
		t.Fatalf("Backup failed: %v", err)
	}

	bak := path + ".bak"
	got, err := os.ReadFile(bak)
	if err != nil {
		t.Fatalf("expected .bak file to exist: %v", err)
	}
	if !bytes.Equal(got, original) {
		t.Errorf(".bak content: expected %q, got %q", string(original), string(got))
	}
}

func TestBackup_DoesNotClobberExistingBak(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")

	// A pre-existing .bak holds an earlier, foreign value that must survive.
	bak := path + ".bak"
	foreign := []byte(`{"model":"foreign-original"}` + "\n")
	if err := os.WriteFile(bak, foreign, 0o644); err != nil {
		t.Fatalf("setup .bak write failed: %v", err)
	}

	// The live settings file now holds a different (newer) value.
	if err := os.WriteFile(path, []byte(`{"model":"newer"}`+"\n"), 0o644); err != nil {
		t.Fatalf("setup settings write failed: %v", err)
	}

	if err := Backup(path); err != nil {
		t.Fatalf("Backup failed: %v", err)
	}

	got, err := os.ReadFile(bak)
	if err != nil {
		t.Fatalf("reading .bak failed: %v", err)
	}
	if !bytes.Equal(got, foreign) {
		t.Errorf("pre-existing .bak was clobbered: expected %q, got %q", string(foreign), string(got))
	}
}

func TestBackup_Idempotent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")
	original := []byte(`{"model":"original"}` + "\n")
	if err := os.WriteFile(path, original, 0o644); err != nil {
		t.Fatalf("setup write failed: %v", err)
	}

	// First backup captures the original.
	if err := Backup(path); err != nil {
		t.Fatalf("first Backup failed: %v", err)
	}

	// Mutate the live file, then back up again — re-running must be safe and
	// must not overwrite the first preserved value.
	if err := os.WriteFile(path, []byte(`{"model":"mutated"}`+"\n"), 0o644); err != nil {
		t.Fatalf("mutating settings failed: %v", err)
	}
	if err := Backup(path); err != nil {
		t.Fatalf("second Backup failed: %v", err)
	}

	got, err := os.ReadFile(path + ".bak")
	if err != nil {
		t.Fatalf("reading .bak failed: %v", err)
	}
	if !bytes.Equal(got, original) {
		t.Errorf("re-run clobbered the first backup: expected %q, got %q", string(original), string(got))
	}
}

func TestBackup_MissingSettingsFile_NoBak(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")

	if err := Backup(path); err != nil {
		t.Fatalf("Backup of missing file should be a no-op, got %v", err)
	}

	if _, err := os.Stat(path + ".bak"); !os.IsNotExist(err) {
		t.Error("expected no .bak when the settings file does not exist")
	}
}

// compactJSON normalizes a raw JSON value to compact form for stable comparison.
func compactJSON(t *testing.T, raw json.RawMessage) string {
	t.Helper()
	var buf bytes.Buffer
	if err := json.Compact(&buf, raw); err != nil {
		t.Fatalf("compacting JSON %q: %v", string(raw), err)
	}
	return buf.String()
}
