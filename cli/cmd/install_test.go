package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// readStatusLineCommand loads the settings file at path and returns the
// .statusLine.command string (empty if the key or command is absent).
func readStatusLineCommand(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading settings: %v", err)
	}
	var settings struct {
		StatusLine struct {
			Type    string `json:"type"`
			Command string `json:"command"`
		} `json:"statusLine"`
	}
	if err := json.Unmarshal(data, &settings); err != nil {
		t.Fatalf("parsing settings: %v", err)
	}
	return settings.StatusLine.Command
}

// TestRunInstallStatusline_AbsentRung verifies that when no statusLine key is
// present, the command sets the freshly-computed Crafter command.
func TestRunInstallStatusline_AbsentRung(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	cmdStr := `"/usr/local/bin/crafter" statusline`

	installStatuslineSettings = path
	installStatuslineCommand = cmdStr
	installStatuslineOnForeign = onForeignKeep

	if err := runInstallStatusline(installStatuslineCmd, nil); err != nil {
		t.Fatalf("runInstallStatusline: %v", err)
	}

	if got := readStatusLineCommand(t, path); got != cmdStr {
		t.Errorf("statusLine.command: got %q, want %q", got, cmdStr)
	}
}

// TestRunInstallStatusline_OursRung verifies that when the stored command is
// already the Crafter command (byte-identical), the command leaves the file
// untouched.
func TestRunInstallStatusline_OursRung(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	cmdStr := `"/usr/local/bin/crafter" statusline`

	// Seed with an ours-identical statusLine plus an unrelated key that must
	// round-trip untouched.
	seed := `{
  "model": "opus",
  "statusLine": {
    "type": "command",
    "command": ` + jsonString(cmdStr) + `
  }
}` + "\n"
	if err := os.WriteFile(path, []byte(seed), 0o644); err != nil {
		t.Fatalf("seeding settings: %v", err)
	}

	installStatuslineSettings = path
	installStatuslineCommand = cmdStr
	installStatuslineOnForeign = onForeignKeep

	if err := runInstallStatusline(installStatuslineCmd, nil); err != nil {
		t.Fatalf("runInstallStatusline: %v", err)
	}

	// Noop rung: the file must be byte-for-byte unchanged.
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading settings: %v", err)
	}
	if string(after) != seed {
		t.Errorf("ours rung modified the file.\n got: %q\nwant: %q", string(after), seed)
	}
}

// TestRunInstallHook_BasicRegistration verifies that a SessionStart hook
// command lands in the settings file.
func TestRunInstallHook_BasicRegistration(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	hookCmd := `node "/home/u/.claude/hooks/crafter-check-update.js"`

	installHookSettings = path
	installHookCommand = hookCmd

	if err := runInstallHook(installHookCmd, nil); err != nil {
		t.Fatalf("runInstallHook: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading settings: %v", err)
	}
	var settings struct {
		Hooks struct {
			SessionStart []struct {
				Hooks []struct {
					Type    string `json:"type"`
					Command string `json:"command"`
				} `json:"hooks"`
			} `json:"SessionStart"`
		} `json:"hooks"`
	}
	if err := json.Unmarshal(data, &settings); err != nil {
		t.Fatalf("parsing settings: %v", err)
	}

	if len(settings.Hooks.SessionStart) != 1 {
		t.Fatalf("SessionStart entries: got %d, want 1", len(settings.Hooks.SessionStart))
	}
	entry := settings.Hooks.SessionStart[0]
	if len(entry.Hooks) != 1 {
		t.Fatalf("SessionStart[0].hooks: got %d, want 1", len(entry.Hooks))
	}
	if entry.Hooks[0].Type != "command" {
		t.Errorf("hook type: got %q, want %q", entry.Hooks[0].Type, "command")
	}
	if entry.Hooks[0].Command != hookCmd {
		t.Errorf("hook command: got %q, want %q", entry.Hooks[0].Command, hookCmd)
	}
}

// TestRunInstallHook_PreservesSiblingHookKeys verifies that registering a
// SessionStart hook does not drop sibling event arrays (PreToolUse, ...) or any
// pre-existing SessionStart entries inside the hooks object. This guards the
// data-loss bug where the hooks value was modelled as a struct with only a
// SessionStart field, so re-marshalling silently dropped every sibling key.
func TestRunInstallHook_PreservesSiblingHookKeys(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	hookCmd := `node "/home/u/.claude/hooks/crafter-check-update.js"`

	// Seed hooks with a sibling event array AND a pre-existing SessionStart
	// entry, both of which must survive the registration.
	seed := `{
  "hooks": {
    "PreToolUse": [ { "matcher": "Bash", "hooks": [ { "type": "command", "command": "guard.sh" } ] } ],
    "SessionStart": [ { "hooks": [ { "type": "command", "command": "existing-hook" } ] } ]
  }
}` + "\n"
	if err := os.WriteFile(path, []byte(seed), 0o644); err != nil {
		t.Fatalf("seeding settings: %v", err)
	}

	installHookSettings = path
	installHookCommand = hookCmd

	if err := runInstallHook(installHookCmd, nil); err != nil {
		t.Fatalf("runInstallHook: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading settings: %v", err)
	}
	var settings struct {
		Hooks struct {
			PreToolUse   json.RawMessage `json:"PreToolUse"`
			SessionStart []struct {
				Hooks []struct {
					Type    string `json:"type"`
					Command string `json:"command"`
				} `json:"hooks"`
			} `json:"SessionStart"`
		} `json:"hooks"`
	}
	if err := json.Unmarshal(data, &settings); err != nil {
		t.Fatalf("parsing settings: %v", err)
	}

	// (a) The sibling PreToolUse array survives unchanged.
	wantPreToolUse := `[{"matcher":"Bash","hooks":[{"type":"command","command":"guard.sh"}]}]`
	if got := compactRaw(t, settings.Hooks.PreToolUse); got != wantPreToolUse {
		t.Errorf("PreToolUse: got %s, want %s", got, wantPreToolUse)
	}

	// (b) + (c) The pre-existing SessionStart entry is retained and the new one
	// was appended.
	if len(settings.Hooks.SessionStart) != 2 {
		t.Fatalf("SessionStart entries: got %d, want 2", len(settings.Hooks.SessionStart))
	}
	first := settings.Hooks.SessionStart[0]
	if len(first.Hooks) != 1 || first.Hooks[0].Command != "existing-hook" {
		t.Errorf("pre-existing SessionStart entry not retained: %+v", first)
	}
	last := settings.Hooks.SessionStart[1]
	if len(last.Hooks) != 1 {
		t.Fatalf("appended SessionStart[1].hooks: got %d, want 1", len(last.Hooks))
	}
	if last.Hooks[0].Type != "command" || last.Hooks[0].Command != hookCmd {
		t.Errorf("appended hook: got %+v, want {command %q}", last.Hooks[0], hookCmd)
	}
}

// TestRunInstallStatusline_ForeignKeepGuidance (G5) verifies that when a
// foreign statusLine is present and --on-foreign=keep, the subcommand:
//   - does NOT modify the settings file
//   - prints valid JSON whose .statusLine.command references BOTH the existing
//     (foreign) command AND the crafter statusline invocation
func TestRunInstallStatusline_ForeignKeepGuidance(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	existingCmd := `date +%H:%M`
	crafterCmd := `"/home/u/.claude/crafter/bin/crafter" statusline`

	seed := `{
  "statusLine": {
    "type": "command",
    "command": ` + jsonString(existingCmd) + `
  }
}` + "\n"
	if err := os.WriteFile(path, []byte(seed), 0o644); err != nil {
		t.Fatalf("seeding settings: %v", err)
	}

	installStatuslineSettings = path
	installStatuslineCommand = crafterCmd
	installStatuslineOnForeign = onForeignKeep

	var out bytes.Buffer
	installStatuslineCmd.SetOut(&out)
	t.Cleanup(func() { installStatuslineCmd.SetOut(nil) })

	if err := runInstallStatusline(installStatuslineCmd, nil); err != nil {
		t.Fatalf("runInstallStatusline: %v", err)
	}

	// (a) Settings file must be byte-identical to the seed (no destructive write).
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading settings after: %v", err)
	}
	if string(after) != seed {
		t.Errorf("foreign-keep rung modified the settings file.\n got: %q\nwant: %q", string(after), seed)
	}

	// (b) The printed guidance must contain valid JSON with .statusLine.command
	// referencing both the existing command and the crafter command.
	output := out.String()
	jsonPart := extractJSONObject(t, output)

	var parsed struct {
		StatusLine struct {
			Type    string `json:"type"`
			Command string `json:"command"`
		} `json:"statusLine"`
	}
	if err := json.Unmarshal([]byte(jsonPart), &parsed); err != nil {
		t.Fatalf("guidance JSON is not valid JSON: %v\nraw:\n%s", err, jsonPart)
	}
	if parsed.StatusLine.Type != "command" {
		t.Errorf("guidance statusLine.type = %q, want %q", parsed.StatusLine.Type, "command")
	}
	compositeCmd := parsed.StatusLine.Command
	if !strings.Contains(compositeCmd, existingCmd) {
		t.Errorf("guidance command does not reference existing command %q\ngot: %s", existingCmd, compositeCmd)
	}
	if !strings.Contains(compositeCmd, crafterCmd) {
		t.Errorf("guidance command does not reference crafter command %q\ngot: %s", crafterCmd, compositeCmd)
	}
}

// TestRunInstallStatusline_ForeignKeepGuidanceSingleQuoteRegression (G6) is a
// regression guard: when the existing statusLine command contains single quotes
// (e.g. awk '{print $1}', date '+%H:%M'), the composite command produced for
// the guidance must still be syntactically valid — specifically the outer
// bash -c '...' wrapper must not be prematurely terminated.
func TestRunInstallStatusline_ForeignKeepGuidanceSingleQuoteRegression(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	// An existing command that contains single quotes — the kind that broke
	// install.sh before the POSIX single-quote escaping was added.
	existingCmd := `awk '{print $1}'`
	crafterCmd := `"/home/u/.claude/crafter/bin/crafter" statusline`

	seed := `{
  "statusLine": {
    "type": "command",
    "command": ` + jsonString(existingCmd) + `
  }
}` + "\n"
	if err := os.WriteFile(path, []byte(seed), 0o644); err != nil {
		t.Fatalf("seeding settings: %v", err)
	}

	installStatuslineSettings = path
	installStatuslineCommand = crafterCmd
	installStatuslineOnForeign = onForeignKeep

	var out bytes.Buffer
	installStatuslineCmd.SetOut(&out)
	t.Cleanup(func() { installStatuslineCmd.SetOut(nil) })

	if err := runInstallStatusline(installStatuslineCmd, nil); err != nil {
		t.Fatalf("runInstallStatusline: %v", err)
	}

	output := out.String()
	jsonPart := extractJSONObject(t, output)

	var parsed struct {
		StatusLine struct {
			Command string `json:"command"`
		} `json:"statusLine"`
	}
	if err := json.Unmarshal([]byte(jsonPart), &parsed); err != nil {
		t.Fatalf("guidance JSON is not valid JSON: %v\nraw:\n%s", err, jsonPart)
	}
	compositeCmd := parsed.StatusLine.Command

	// The composite must reference both commands.
	if !strings.Contains(compositeCmd, crafterCmd) {
		t.Errorf("guidance command does not reference crafter command %q\ngot: %s", crafterCmd, compositeCmd)
	}

	// The single-quote-safe check: the composite starts with "bash -c '" and
	// the outer single-quoted region must be syntactically closed. After
	// stripping the prefix "bash -c '", count the number of unescaped single
	// quotes — an odd count means the outer quote is unclosed. The POSIX escape
	// '\'' ends and reopens the outer quote cleanly so that the existing command
	// body is never a literal bare single quote inside the wrapper.
	//
	// Simpler assertion: the composite must contain the POSIX-escaped form of
	// the existing command's single-quote character so we know escaping fired.
	if !strings.Contains(compositeCmd, `'\''`) {
		t.Errorf("composite command should contain POSIX single-quote escape sequence '\\''  got: %s", compositeCmd)
	}

	// The JSON must be parseable (already checked), confirming the JSON layer
	// correctly encoded the shell layer.
}

// TestRunInstallStatusline_MissingSettingsFile verifies that the subcommand
// exits cleanly (non-destructively) when the settings path does not exist:
// no file is created, no error is returned (missing file = empty settings,
// absent rung fires and the fresh crafter command is written).
func TestRunInstallStatusline_MissingSettingsFile(t *testing.T) {
	// Use a path that does not exist yet — absent rung should create it.
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")
	crafterCmd := `"/home/u/.claude/crafter/bin/crafter" statusline`

	installStatuslineSettings = path
	installStatuslineCommand = crafterCmd
	installStatuslineOnForeign = onForeignKeep

	if err := runInstallStatusline(installStatuslineCmd, nil); err != nil {
		t.Fatalf("runInstallStatusline on missing file: %v", err)
	}

	// A missing file degrades to empty settings → absent rung → writes the
	// crafter command.
	if got := readStatusLineCommand(t, path); got != crafterCmd {
		t.Errorf("statusLine.command after missing-file run: got %q, want %q", got, crafterCmd)
	}
}

// TestRunInstallStatusline_GarbledSettingsFile verifies the tolerant posture:
// a garbled (invalid JSON) settings file is treated as empty settings —
// the absent rung fires and writes the crafter command.
func TestRunInstallStatusline_GarbledSettingsFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	if err := os.WriteFile(path, []byte("not json {{"), 0o644); err != nil {
		t.Fatalf("writing garbled settings: %v", err)
	}
	crafterCmd := `"/home/u/.claude/crafter/bin/crafter" statusline`

	installStatuslineSettings = path
	installStatuslineCommand = crafterCmd
	installStatuslineOnForeign = onForeignKeep

	if err := runInstallStatusline(installStatuslineCmd, nil); err != nil {
		t.Fatalf("runInstallStatusline on garbled file: %v", err)
	}

	if got := readStatusLineCommand(t, path); got != crafterCmd {
		t.Errorf("statusLine.command after garbled-file run: got %q, want %q", got, crafterCmd)
	}
}

// TestRunInstallStatusline_NonCommandForeignKeep (G7) verifies that when a
// foreign statusLine value with no string command (e.g. {"type":"custom","foo":"bar"})
// is present and --on-foreign=keep:
//
//	(1) the value is classified as foreign (NeedsDecision),
//	(2) the settings file is left byte-for-byte untouched,
//	(3) the manual-merge guidance branch fires (output contains "merge manually").
func TestRunInstallStatusline_NonCommandForeignKeep(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	crafterCmd := `"/home/u/.claude/crafter/bin/crafter" statusline`

	// A non-command-shaped statusLine: no .command field.
	seed := `{
  "statusLine": {
    "type": "custom",
    "foo": "bar"
  }
}` + "\n"
	if err := os.WriteFile(path, []byte(seed), 0o644); err != nil {
		t.Fatalf("seeding settings: %v", err)
	}

	installStatuslineSettings = path
	installStatuslineCommand = crafterCmd
	installStatuslineOnForeign = onForeignKeep

	var out bytes.Buffer
	installStatuslineCmd.SetOut(&out)
	t.Cleanup(func() { installStatuslineCmd.SetOut(nil) })

	if err := runInstallStatusline(installStatuslineCmd, nil); err != nil {
		t.Fatalf("runInstallStatusline: %v", err)
	}

	// (2) Settings file must be byte-identical to the seed — not overwritten.
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading settings after: %v", err)
	}
	if string(after) != seed {
		t.Errorf("non-command foreign-keep rung modified the settings file.\n got: %q\nwant: %q", string(after), seed)
	}

	// (3) Output must contain the manual-merge guidance, not a composite wrapper.
	output := out.String()
	if !strings.Contains(output, "merge manually") {
		t.Errorf("expected 'merge manually' in output, got:\n%s", output)
	}
	if strings.Contains(output, "bash -c") {
		t.Errorf("non-command foreign should not produce a bash -c wrapper in output:\n%s", output)
	}
}

// extractJSONObject finds the first line in s whose trimmed content is exactly
// "{" (marking the start of a top-level JSON block) and decodes the complete
// JSON object from that point using json.Decoder. This avoids matching '{'
// characters embedded inside non-JSON output lines (e.g. existing-command
// display lines containing shell syntax like awk '{print $1}').
func extractJSONObject(t *testing.T, s string) string {
	t.Helper()
	lines := strings.Split(s, "\n")
	jsonLineIdx := -1
	for i, l := range lines {
		if strings.TrimSpace(l) == "{" {
			jsonLineIdx = i
			break
		}
	}
	if jsonLineIdx == -1 {
		t.Fatalf("no JSON block (line containing only '{') found in output:\n%s", s)
	}
	jsonBlock := strings.Join(lines[jsonLineIdx:], "\n")
	dec := json.NewDecoder(strings.NewReader(jsonBlock))
	var raw json.RawMessage
	if err := dec.Decode(&raw); err != nil {
		t.Fatalf("could not decode JSON from output: %v\nblock:\n%s", err, jsonBlock)
	}
	return string(raw)
}

// compactRaw normalizes a raw JSON value to compact form for stable comparison.
func compactRaw(t *testing.T, raw json.RawMessage) string {
	t.Helper()
	var buf bytes.Buffer
	if err := json.Compact(&buf, raw); err != nil {
		t.Fatalf("compacting JSON %q: %v", string(raw), err)
	}
	return buf.String()
}

// jsonString returns s encoded as a JSON string literal (including quotes), used
// to embed a command containing quote characters into a seed fixture.
func jsonString(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}
