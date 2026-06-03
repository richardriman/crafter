package claudesettings

import (
	"encoding/json"
	"path"
	"strings"
)

// statusLineKey is the settings.json key under which Claude Code stores the
// status-line command. install.sh writes it as
// `{ "type": "command", "command": "<cmd>" }`.
const statusLineKey = "statusLine"

// crafterStatuslineToken is the subcommand that follows the crafter binary in
// a crafter status-line invocation (install.sh builds `"<bin>" statusline`).
const crafterStatuslineToken = "statusline"

// crafterBinBasename is the basename of the crafter executable; the first token
// of an "ours" command is the binary path, which must end in this name.
const crafterBinBasename = "crafter"

// StatusLineClass classifies an existing statusLine value in settings relative
// to the crafter status-line command.
type StatusLineClass int

const (
	// StatusLineAbsent means there is no statusLine key (or it carries no
	// usable string command).
	StatusLineAbsent StatusLineClass = iota
	// StatusLineOurs means the stored .command is recognizably the crafter
	// statusline invocation (see IsOurStatusLine for the exact rule).
	StatusLineOurs
	// StatusLineForeign means a statusLine command is present but is not ours
	// (a user/other-tool command, including a composite wrapper that merely
	// embeds the crafter call among other commands).
	StatusLineForeign
)

// StatusLineAction is the action a caller should take after classifying and
// reconciling the statusLine value against the freshly-computed crafter command.
type StatusLineAction int

const (
	// StatusLineSet: no crafter statusLine was present; the caller should set
	// the freshly-computed command.
	StatusLineSet StatusLineAction = iota
	// StatusLineNoop: the stored command is ours AND byte-identical to the
	// freshly-computed command; the caller must NOT write.
	StatusLineNoop
	// StatusLineUpdated: the stored command is ours but differs (e.g. the
	// binary path moved); the caller should overwrite with the fresh command.
	StatusLineUpdated
	// StatusLineNeedsDecision: a foreign command is present. This step does
	// NOT act on it; a later phase decides keep-vs-overwrite (with backup and
	// guidance). The caller must not write here.
	StatusLineNeedsDecision
)

// statusLineValue is the structural shape of a settings.json statusLine value:
// `{ "type": "command", "command": "<cmd>" }`. Only .command is load-bearing
// for classification.
type statusLineValue struct {
	Command string `json:"command"`
}

// statusLineCommand extracts the .command string from the raw statusLine value.
// It returns ("", false) when the value is absent, not an object, or has no
// string command (mirroring install.sh's `typeof existingStatusLine.command ===
// "string"` guard, which treats a non-command-shaped value as "no command").
func statusLineCommand(raw json.RawMessage, present bool) (string, bool) {
	if !present {
		return "", false
	}
	var v statusLineValue
	if err := json.Unmarshal(raw, &v); err != nil {
		return "", false
	}
	if v.Command == "" {
		return "", false
	}
	return v.Command, true
}

// IsOurStatusLine reports whether command is the crafter status-line invocation.
//
// Matcher rule (authoritative boundary for this feature): the command must be
// EXACTLY two shell-style tokens — a binary path whose basename is "crafter",
// followed by the bare "statusline" subcommand, with nothing after it. The
// binary path may be quoted (install.sh emits `"<bin>" statusline`); surrounding
// single/double quotes on each token are stripped before comparison.
//
// This structural "exactly two tokens" rule is deliberately NOT a substring
// match: the composite tee-wrapper this feature used to emit as guidance
// (`bash -c 'in=$(cat); printf "%s %s" "..." "$(... "<bin>" statusline)"'`)
// tokenizes to a first token of `bash`, so it is FOREIGN — the embedded crafter
// call lives inside a later quoted argument and never forms the standalone
// two-token shape. Anything that wraps or extends the crafter call therefore
// fails the rule and is treated as foreign.
func IsOurStatusLine(command string) bool {
	tokens := tokenizeCommand(command)
	if len(tokens) != 2 {
		return false
	}
	if tokens[1] != crafterStatuslineToken {
		return false
	}
	return path.Base(tokens[0]) == crafterBinBasename
}

// tokenizeCommand splits a command string into whitespace-separated tokens,
// honouring single and double quotes so that a quoted binary path containing
// spaces (or simply the install.sh `"<bin>"` quoting) stays a single token.
// Quote characters are stripped from the emitted tokens. A token boundary
// occurs only at unquoted whitespace; this is sufficient to distinguish the
// plain two-token crafter command from any wrapper that introduces additional
// unquoted tokens (e.g. `bash -c '...'`).
func tokenizeCommand(command string) []string {
	var tokens []string
	var cur strings.Builder
	inToken := false
	var quote rune // 0 when not inside a quote

	flush := func() {
		if inToken {
			tokens = append(tokens, cur.String())
			cur.Reset()
			inToken = false
		}
	}

	for _, r := range command {
		switch {
		case quote != 0:
			if r == quote {
				quote = 0
			} else {
				cur.WriteRune(r)
			}
			inToken = true
		case r == '\'' || r == '"':
			quote = r
			inToken = true
		case r == ' ' || r == '\t' || r == '\n' || r == '\r':
			flush()
		default:
			cur.WriteRune(r)
			inToken = true
		}
	}
	flush()
	return tokens
}

// ClassifyStatusLine inspects the statusLine value held in s and reports
// whether it is absent, ours, or foreign.
func ClassifyStatusLine(s *Settings) StatusLineClass {
	raw, present := s.Get(statusLineKey)
	cmd, ok := statusLineCommand(raw, present)
	if !ok {
		return StatusLineAbsent
	}
	if IsOurStatusLine(cmd) {
		return StatusLineOurs
	}
	return StatusLineForeign
}

// ReconcileStatusLine applies the two NON-PROMPTING rungs of the statusLine
// decision tree against the freshly-computed crafter command and reports the
// action the caller should take. It is a pure function: it does NOT read or
// write any file and does NOT mutate s.
//
//   - absent          -> StatusLineSet       (write crafterCmd)
//   - ours, identical -> StatusLineNoop      (write nothing)
//   - ours, differs   -> StatusLineUpdated   (write crafterCmd)
//   - foreign         -> StatusLineNeedsDecision (write nothing; later phase decides)
//
// The returned value-to-write (a `{ "type": "command", "command": crafterCmd }`
// object) is only meaningful for StatusLineSet and StatusLineUpdated; it is the
// zero value for StatusLineNoop and StatusLineNeedsDecision.
func ReconcileStatusLine(s *Settings, crafterCmd string) (StatusLineAction, StatusLineValue) {
	raw, present := s.Get(statusLineKey)
	cmd, ok := statusLineCommand(raw, present)

	switch {
	case !ok:
		// Absent (or non-command-shaped): install fresh.
		return StatusLineSet, newStatusLineValue(crafterCmd)
	case IsOurStatusLine(cmd):
		if cmd == crafterCmd {
			// Byte-identical to the fresh command: no write at all.
			return StatusLineNoop, StatusLineValue{}
		}
		// Ours but stale (e.g. moved binary path): refresh it.
		return StatusLineUpdated, newStatusLineValue(crafterCmd)
	default:
		// Foreign: this step does not act; a later phase decides.
		return StatusLineNeedsDecision, StatusLineValue{}
	}
}

// StatusLineValue is the object a caller writes back into settings under the
// statusLine key for the StatusLineSet / StatusLineUpdated actions.
type StatusLineValue struct {
	Type    string `json:"type"`
	Command string `json:"command"`
}

// newStatusLineValue builds the canonical crafter statusLine value shape,
// matching install.sh's `{ type: "command", command: <cmd> }`.
func newStatusLineValue(command string) StatusLineValue {
	return StatusLineValue{Type: "command", Command: command}
}
