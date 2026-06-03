package claudesettings

import (
	"encoding/json"
	"strings"
	"testing"
)

// compositeWrapper is the tee-wrapper command this feature used to emit as
// guidance (built by install.sh's install_statusline node block). It EMBEDS the
// crafter statusline call among other commands and must classify as FOREIGN,
// not ours.
const compositeWrapper = `bash -c 'in=$(cat); printf "%s %s" "$(printf "%s" "$in" | date +%H:%M)" "$(printf "%s" "$in" | "/home/u/.claude/crafter/bin/crafter" statusline)"'`

func TestIsOurStatusLine(t *testing.T) {
	tests := []struct {
		name    string
		command string
		want    bool
	}{
		{
			name:    "true-positive: plain quoted crafter command (install.sh shape)",
			command: `"/home/u/.claude/crafter/bin/crafter" statusline`,
			want:    true,
		},
		{
			name:    "true-positive: unquoted crafter command",
			command: `/home/u/.claude/crafter/bin/crafter statusline`,
			want:    true,
		},
		{
			name:    "true-positive: single-quoted binary path",
			command: `'/home/u/.claude/crafter/bin/crafter' statusline`,
			want:    true,
		},
		{
			name:    "true-positive: local install path",
			command: `"/proj/.claude/crafter/bin/crafter" statusline`,
			want:    true,
		},
		{
			name:    "true-negative: composite tee-wrapper embedding crafter call",
			command: compositeWrapper,
			want:    false,
		},
		{
			name:    "true-negative: extra argument after statusline",
			command: `"/home/u/.claude/crafter/bin/crafter" statusline --foo`,
			want:    false,
		},
		{
			name:    "true-negative: wrong subcommand token",
			command: `"/home/u/.claude/crafter/bin/crafter" status`,
			want:    false,
		},
		{
			name:    "true-negative: wrong binary basename",
			command: `"/usr/bin/other" statusline`,
			want:    false,
		},
		{
			name:    "true-negative: bare statusline token only",
			command: `statusline`,
			want:    false,
		},
		{
			name:    "true-negative: unrelated user command",
			command: `~/.local/bin/some-statusline-tool`,
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsOurStatusLine(tt.command); got != tt.want {
				t.Errorf("IsOurStatusLine(%q) = %v, want %v", tt.command, got, tt.want)
			}
		})
	}
}

func TestClassifyStatusLine(t *testing.T) {
	const ours = `"/home/u/.claude/crafter/bin/crafter" statusline`

	tests := []struct {
		name  string
		setup func(*Settings)
		want  StatusLineClass
	}{
		{
			name:  "absent: no statusLine key",
			setup: func(s *Settings) {},
			want:  StatusLineAbsent,
		},
		{
			name: "foreign: statusLine present but non-command shape (no string command)",
			setup: func(s *Settings) {
				s.SetRaw(statusLineKey, []byte(`{"type":"command"}`))
			},
			want: StatusLineForeign,
		},
		{
			name: "foreign: statusLine present with custom type and no command field",
			setup: func(s *Settings) {
				s.SetRaw(statusLineKey, []byte(`{"type":"custom","foo":"bar"}`))
			},
			want: StatusLineForeign,
		},
		{
			name: "ours: crafter command",
			setup: func(s *Settings) {
				if err := s.Set(statusLineKey, newStatusLineValue(ours)); err != nil {
					t.Fatalf("setup Set failed: %v", err)
				}
			},
			want: StatusLineOurs,
		},
		{
			name: "foreign: composite wrapper",
			setup: func(s *Settings) {
				if err := s.Set(statusLineKey, newStatusLineValue(compositeWrapper)); err != nil {
					t.Fatalf("setup Set failed: %v", err)
				}
			},
			want: StatusLineForeign,
		},
		{
			name: "foreign: unrelated user command",
			setup: func(s *Settings) {
				if err := s.Set(statusLineKey, newStatusLineValue("starship")); err != nil {
					t.Fatalf("setup Set failed: %v", err)
				}
			},
			want: StatusLineForeign,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := New()
			tt.setup(s)
			if got := ClassifyStatusLine(s); got != tt.want {
				t.Errorf("ClassifyStatusLine() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReconcileStatusLine(t *testing.T) {
	const fresh = `"/home/u/.claude/crafter/bin/crafter" statusline`
	// "ours" but at an old binary path (binary moved) — differs from fresh.
	const stale = `"/home/u/.crafter/bin/crafter" statusline`

	tests := []struct {
		name        string
		setup       func(*Settings)
		crafterCmd  string
		wantAction  StatusLineAction
		wantWrite   bool   // whether a value-to-write is expected
		wantCommand string // expected .command when wantWrite
	}{
		{
			name:        "absent -> set",
			setup:       func(s *Settings) {},
			crafterCmd:  fresh,
			wantAction:  StatusLineSet,
			wantWrite:   true,
			wantCommand: fresh,
		},
		{
			name: "ours-identical -> noop (no write)",
			setup: func(s *Settings) {
				if err := s.Set(statusLineKey, newStatusLineValue(fresh)); err != nil {
					t.Fatalf("setup Set failed: %v", err)
				}
			},
			crafterCmd: fresh,
			wantAction: StatusLineNoop,
			wantWrite:  false,
		},
		{
			name: "ours-differs (moved binary path) -> updated",
			setup: func(s *Settings) {
				if err := s.Set(statusLineKey, newStatusLineValue(stale)); err != nil {
					t.Fatalf("setup Set failed: %v", err)
				}
			},
			crafterCmd:  fresh,
			wantAction:  StatusLineUpdated,
			wantWrite:   true,
			wantCommand: fresh,
		},
		{
			name: "foreign composite wrapper -> needs decision (no write)",
			setup: func(s *Settings) {
				if err := s.Set(statusLineKey, newStatusLineValue(compositeWrapper)); err != nil {
					t.Fatalf("setup Set failed: %v", err)
				}
			},
			crafterCmd: fresh,
			wantAction: StatusLineNeedsDecision,
			wantWrite:  false,
		},
		{
			name: "foreign user command -> needs decision (no write)",
			setup: func(s *Settings) {
				if err := s.Set(statusLineKey, newStatusLineValue("starship")); err != nil {
					t.Fatalf("setup Set failed: %v", err)
				}
			},
			crafterCmd: fresh,
			wantAction: StatusLineNeedsDecision,
			wantWrite:  false,
		},
		{
			name: "foreign non-command shape -> needs decision (no write)",
			setup: func(s *Settings) {
				s.SetRaw(statusLineKey, []byte(`{"type":"custom","foo":"bar"}`))
			},
			crafterCmd: fresh,
			wantAction: StatusLineNeedsDecision,
			wantWrite:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := New()
			tt.setup(s)

			gotAction, gotValue := ReconcileStatusLine(s, tt.crafterCmd)
			if gotAction != tt.wantAction {
				t.Errorf("action = %v, want %v", gotAction, tt.wantAction)
			}

			if tt.wantWrite {
				if gotValue.Type != "command" {
					t.Errorf("value.Type = %q, want %q", gotValue.Type, "command")
				}
				if gotValue.Command != tt.wantCommand {
					t.Errorf("value.Command = %q, want %q", gotValue.Command, tt.wantCommand)
				}
			} else if (gotValue != StatusLineValue{}) {
				t.Errorf("expected zero value-to-write for action %v, got %+v", gotAction, gotValue)
			}
		})
	}
}

// TestShSingleQuoteEscape verifies the POSIX single-quote escaping helper.
func TestShSingleQuoteEscape(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{input: `no quotes`, want: `no quotes`},
		{input: `awk '{print $1}'`, want: `awk '\''{print $1}'\''`},
		{input: `date +%H:%M`, want: `date +%H:%M`},
		{input: `a'b'c`, want: `a'\''b'\''c`},
		// A single bare quote becomes '\''
		{input: `'`, want: `'\''`},
	}

	for _, tt := range tests {
		got := shSingleQuoteEscape(tt.input)
		if got != tt.want {
			t.Errorf("shSingleQuoteEscape(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

// TestCompositeStatusLineCmd verifies that compositeStatusLineCmd builds a
// bash -c wrapper that references both the existing and crafter commands.
func TestCompositeStatusLineCmd(t *testing.T) {
	existing := `date +%H:%M`
	crafter := `"/home/u/.claude/crafter/bin/crafter" statusline`
	got := compositeStatusLineCmd(existing, crafter)

	if !strings.HasPrefix(got, "bash -c '") {
		t.Errorf("composite does not start with bash -c ': %s", got)
	}
	if !strings.Contains(got, existing) {
		t.Errorf("composite does not contain existing command %q: %s", existing, got)
	}
	if !strings.Contains(got, crafter) {
		t.Errorf("composite does not contain crafter command %q: %s", crafter, got)
	}
}

// TestCompositeStatusLineCmd_SingleQuoteInExisting verifies POSIX escaping
// fires when the existing command contains single quotes.
func TestCompositeStatusLineCmd_SingleQuoteInExisting(t *testing.T) {
	existing := `awk '{print $1}'`
	crafter := `"/home/u/.claude/crafter/bin/crafter" statusline`
	got := compositeStatusLineCmd(existing, crafter)

	// The outer wrapper is bash -c '...': a bare single quote from the existing
	// command would prematurely close it. After escaping each ' becomes '\''
	// so the raw awk '{print $1}' is NOT present as-is inside the composite.
	if strings.Contains(got, `awk '{print $1}'`) {
		t.Errorf("composite contains un-escaped single quotes from existing command: %s", got)
	}
	// The POSIX escape sequence must appear.
	if !strings.Contains(got, `'\''`) {
		t.Errorf("composite does not contain POSIX single-quote escape sequence: %s", got)
	}
	// The crafter command must still be present.
	if !strings.Contains(got, crafter) {
		t.Errorf("composite does not contain crafter command %q: %s", crafter, got)
	}
}

// extractGuidanceJSON finds the first element in lines that begins with '{'
// (after joining lines that are multi-line strings) and decodes the complete
// JSON object from that element using json.Decoder.
//
// ForeignKeepGuidanceLines returns the JSON blob as a single slice element
// (the result of json.MarshalIndent, which is a multi-line string). So we scan
// elements for one whose TrimSpace starts with '{', then decode from it.
func extractGuidanceJSON(t *testing.T, lines []string) string {
	t.Helper()
	for _, l := range lines {
		if strings.HasPrefix(strings.TrimSpace(l), "{") {
			dec := json.NewDecoder(strings.NewReader(l))
			var raw json.RawMessage
			if err := dec.Decode(&raw); err != nil {
				t.Fatalf("could not decode JSON from guidance line: %v\nline:\n%s", err, l)
			}
			return string(raw)
		}
	}
	t.Fatalf("no JSON element (starting with '{') found in guidance lines:\n%s", strings.Join(lines, "\n"))
	return ""
}

// TestForeignKeepGuidanceLines_CommandShape verifies that
// ForeignKeepGuidanceLines returns lines whose embedded JSON is valid and
// references both the existing and crafter commands.
func TestForeignKeepGuidanceLines_CommandShape(t *testing.T) {
	existingCmd := `date +%H:%M`
	crafterCmd := `"/home/u/.claude/crafter/bin/crafter" statusline`
	rawExisting, _ := json.Marshal(newStatusLineValue(existingCmd))

	lines, err := ForeignKeepGuidanceLines("/home/u/.claude/settings.json", rawExisting, crafterCmd)
	if err != nil {
		t.Fatalf("ForeignKeepGuidanceLines: %v", err)
	}
	if len(lines) == 0 {
		t.Fatal("expected non-empty guidance lines")
	}

	jsonStr := extractGuidanceJSON(t, lines)
	var parsed struct {
		StatusLine struct {
			Type    string `json:"type"`
			Command string `json:"command"`
		} `json:"statusLine"`
	}
	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		t.Fatalf("guidance JSON is not valid: %v\nraw:\n%s", err, jsonStr)
	}
	if parsed.StatusLine.Type != "command" {
		t.Errorf("statusLine.type = %q, want %q", parsed.StatusLine.Type, "command")
	}
	compositeCmd := parsed.StatusLine.Command
	if !strings.Contains(compositeCmd, existingCmd) {
		t.Errorf("composite does not reference existing command %q: %s", existingCmd, compositeCmd)
	}
	if !strings.Contains(compositeCmd, crafterCmd) {
		t.Errorf("composite does not reference crafter command %q: %s", crafterCmd, compositeCmd)
	}
}

// TestForeignKeepGuidanceLines_SingleQuoteSafe (G6 analog) verifies that when
// the existing command contains single quotes, the guidance JSON is still valid
// and the composite command uses POSIX escaping.
func TestForeignKeepGuidanceLines_SingleQuoteSafe(t *testing.T) {
	existingCmd := `awk '{print $1}'`
	crafterCmd := `"/home/u/.claude/crafter/bin/crafter" statusline`
	rawExisting, _ := json.Marshal(newStatusLineValue(existingCmd))

	lines, err := ForeignKeepGuidanceLines("/home/u/.claude/settings.json", rawExisting, crafterCmd)
	if err != nil {
		t.Fatalf("ForeignKeepGuidanceLines: %v", err)
	}

	jsonStr := extractGuidanceJSON(t, lines)
	var parsed struct {
		StatusLine struct {
			Command string `json:"command"`
		} `json:"statusLine"`
	}
	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		t.Fatalf("guidance JSON invalid for single-quote-containing existing command: %v\nraw:\n%s", err, jsonStr)
	}
	compositeCmd := parsed.StatusLine.Command
	// POSIX escaping must have fired.
	if !strings.Contains(compositeCmd, `'\''`) {
		t.Errorf("composite does not contain POSIX single-quote escape: %s", compositeCmd)
	}
	// The crafter command must still appear.
	if !strings.Contains(compositeCmd, crafterCmd) {
		t.Errorf("composite does not reference crafter command %q: %s", crafterCmd, compositeCmd)
	}
}

// TestForeignKeepGuidanceLines_NonCommandShape verifies the fallback for an
// existing statusLine that has no string command (malformed or non-command type).
func TestForeignKeepGuidanceLines_NonCommandShape(t *testing.T) {
	// A statusLine value with no .command field.
	rawExisting := json.RawMessage(`{"type":"custom","foo":"bar"}`)
	crafterCmd := `"/home/u/.claude/crafter/bin/crafter" statusline`

	lines, err := ForeignKeepGuidanceLines("/home/u/.claude/settings.json", rawExisting, crafterCmd)
	if err != nil {
		t.Fatalf("ForeignKeepGuidanceLines (non-command): %v", err)
	}

	combined := strings.Join(lines, "\n")
	if !strings.Contains(combined, "merge manually") {
		t.Errorf("expected 'merge manually' fallback message, got:\n%s", combined)
	}
	// Should NOT contain a composite bash -c wrapper.
	if strings.Contains(combined, "bash -c") {
		t.Errorf("non-command shape should not produce a bash -c wrapper:\n%s", combined)
	}
}

// TestReconcileStatusLine_NoMutation guards the purity contract: Reconcile must
// not mutate the Settings it inspects, regardless of the action returned.
func TestReconcileStatusLine_NoMutation(t *testing.T) {
	s := New()
	if err := s.Set(statusLineKey, newStatusLineValue("starship")); err != nil {
		t.Fatalf("setup Set failed: %v", err)
	}
	before, _ := s.Get(statusLineKey)
	beforeStr := string(before)

	ReconcileStatusLine(s, `"/home/u/.claude/crafter/bin/crafter" statusline`)

	after, ok := s.Get(statusLineKey)
	if !ok {
		t.Fatal("statusLine key disappeared after Reconcile")
	}
	if string(after) != beforeStr {
		t.Errorf("Reconcile mutated settings: before %q, after %q", beforeStr, string(after))
	}
}
