package cmd

import (
	"encoding/json"

	"github.com/spf13/cobra"

	"github.com/richardriman/crafter/cli/internal/claudesettings"
)

const (
	hooksKey        = "hooks"
	sessionStartKey = "SessionStart"
)

var (
	installHookSettings string
	installHookCommand  string
)

// hookEntry mirrors an install.sh SessionStart entry: an object carrying a
// hooks array of command descriptors.
type hookEntry struct {
	Hooks []hookCommand `json:"hooks"`
}

// hookCommand is a single `{ "type": "command", "command": "<cmd>" }` descriptor.
type hookCommand struct {
	Type    string `json:"type"`
	Command string `json:"command"`
}

var installHookCmd = &cobra.Command{
	Use:   "hook",
	Short: "Register the Crafter SessionStart hook into a Claude Code settings.json",
	Long: "Register a SessionStart hook command into a settings.json. This is a " +
		"basic registration; the behaviour-preserving, idempotent port is a later " +
		"phase.",
	SilenceUsage: true,
	RunE:         runInstallHook,
}

func init() {
	installHookCmd.Flags().StringVar(&installHookSettings, "settings", "", "path to the target settings.json (required)")
	installHookCmd.Flags().StringVar(&installHookCommand, "command", "", "the hook command to register, e.g. node \"<hook_dest>\" (required)")

	_ = installHookCmd.MarkFlagRequired("settings")
	_ = installHookCmd.MarkFlagRequired("command")

	installCmd.AddCommand(installHookCmd)
}

func runInstallHook(cmd *cobra.Command, args []string) error {
	settings, err := claudesettings.Load(installHookSettings)
	if err != nil {
		return err
	}

	// Hold the hooks object as a raw-value map, mirroring how claudesettings
	// holds top-level settings: every sibling event array (PreToolUse, Stop, ...)
	// and any unknown nested key round-trips verbatim, and only SessionStart is
	// mutated below.
	hooks := map[string]json.RawMessage{}
	if raw, ok := settings.Get(hooksKey); ok {
		// Ignore a malformed hooks value: treat it as empty, mirroring the
		// tolerant posture of the install.sh node block.
		_ = json.Unmarshal(raw, &hooks)
	}

	// Decode the existing SessionStart array (missing/invalid -> empty), append
	// the new entry, and re-encode it back under the same key. install.sh dedup
	// (skip-if-already-registered) is a later phase; for now we always append.
	var sessionStart []hookEntry
	if raw, ok := hooks[sessionStartKey]; ok {
		_ = json.Unmarshal(raw, &sessionStart)
	}
	sessionStart = append(sessionStart, hookEntry{
		Hooks: []hookCommand{{Type: "command", Command: installHookCommand}},
	})

	encoded, err := json.Marshal(sessionStart)
	if err != nil {
		return err
	}
	hooks[sessionStartKey] = encoded

	if err := settings.Set(hooksKey, hooks); err != nil {
		return err
	}
	return claudesettings.Save(installHookSettings, settings)
}
