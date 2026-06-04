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

// hookEntryProbe is a minimal struct used ONLY to read inner hooks[].command
// values from a raw SessionStart entry during the idempotency check. It is
// never re-marshalled back; the raw entry bytes are always preserved verbatim.
type hookEntryProbe struct {
	Hooks []struct {
		Command string `json:"command"`
	} `json:"hooks"`
}

var installHookCmd = &cobra.Command{
	Use:   "hook",
	Short: "Register the Crafter SessionStart hook into a Claude Code settings.json",
	Long: "Register a SessionStart hook command into a settings.json. " +
		"The registration is idempotent: if the command is already present in any " +
		"existing SessionStart entry the file is left untouched. Pre-existing " +
		"SessionStart entries (including any extra fields such as matcher or " +
		"inner command timeouts) are preserved verbatim.",
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

	// Hold the hooks object as a raw-value map so every sibling event array
	// (PreToolUse, Stop, ...) and unknown nested key round-trips verbatim.
	// Only the SessionStart array is mutated below.
	hooks := map[string]json.RawMessage{}
	if raw, ok := settings.Get(hooksKey); ok {
		// Ignore a malformed hooks value: treat it as empty, mirroring the
		// tolerant posture of the install.sh node block.
		_ = json.Unmarshal(raw, &hooks)
	}

	// Model the existing SessionStart array as []json.RawMessage so each entry
	// is kept byte-for-byte intact (preserving entry-level fields like matcher
	// and inner-command fields like timeout that are not in our struct).
	// Missing or invalid value degrades to an empty slice.
	var sessionStart []json.RawMessage
	if raw, ok := hooks[sessionStartKey]; ok {
		_ = json.Unmarshal(raw, &sessionStart)
	}

	// Idempotency check: scan each raw entry using a minimal probe struct that
	// reads only hooks[].command — we never re-marshal the probe, so no fields
	// of the original entry are ever discarded.
	// Mirrors the install.sh node block's alreadyRegistered check:
	//   settings.hooks.SessionStart.some(e => e.hooks && e.hooks.some(h => h.command === hookCommand))
	alreadyRegistered := false
	for _, rawEntry := range sessionStart {
		var probe hookEntryProbe
		if err := json.Unmarshal(rawEntry, &probe); err != nil {
			// Malformed entry: skip without altering.
			continue
		}
		for _, h := range probe.Hooks {
			if h.Command == installHookCommand {
				alreadyRegistered = true
				break
			}
		}
		if alreadyRegistered {
			break
		}
	}

	if !alreadyRegistered {
		// Build the new crafter-owned entry and append it as a raw JSON value.
		newEntry, err := json.Marshal(map[string]interface{}{
			"hooks": []map[string]string{
				{"type": "command", "command": installHookCommand},
			},
		})
		if err != nil {
			return err
		}
		sessionStart = append(sessionStart, json.RawMessage(newEntry))

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

	return nil
}
