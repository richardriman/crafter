package claudesettings

import "testing"

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
			name: "absent: statusLine present but non-command shape (no string command)",
			setup: func(s *Settings) {
				s.SetRaw(statusLineKey, []byte(`{"type":"command"}`))
			},
			want: StatusLineAbsent,
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
