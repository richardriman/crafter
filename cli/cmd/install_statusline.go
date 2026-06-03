package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/richardriman/crafter/cli/internal/claudesettings"
)

const (
	onForeignKeep      = "keep"
	onForeignOverwrite = "overwrite"
)

var (
	installStatuslineSettings  string
	installStatuslineCommand   string
	installStatuslineOnForeign string
)

var installStatuslineCmd = &cobra.Command{
	Use:   "statusline",
	Short: "Reconcile the Crafter statusLine into a Claude Code settings.json",
	Long: "Non-interactively reconcile the Crafter status-line command into a " +
		"settings.json. The caller (install.sh) computes the command and, when a " +
		"foreign statusLine is present, resolves the keep-vs-overwrite decision and " +
		"passes it via --on-foreign; this command never prompts or reads a TTY.",
	SilenceUsage: true,
	RunE:         runInstallStatusline,
}

func init() {
	installStatuslineCmd.Flags().StringVar(&installStatuslineSettings, "settings", "", "path to the target settings.json (required)")
	installStatuslineCmd.Flags().StringVar(&installStatuslineCommand, "command", "", "the computed Crafter statusline command to install (required)")
	installStatuslineCmd.Flags().StringVar(&installStatuslineOnForeign, "on-foreign", onForeignKeep, "what to do when a foreign statusLine is present: keep|overwrite")

	_ = installStatuslineCmd.MarkFlagRequired("settings")
	_ = installStatuslineCmd.MarkFlagRequired("command")

	installCmd.AddCommand(installStatuslineCmd)
}

func runInstallStatusline(cmd *cobra.Command, args []string) error {
	if installStatuslineOnForeign != onForeignKeep && installStatuslineOnForeign != onForeignOverwrite {
		return fmt.Errorf("--on-foreign must be %q or %q, got %q", onForeignKeep, onForeignOverwrite, installStatuslineOnForeign)
	}

	settings, err := claudesettings.Load(installStatuslineSettings)
	if err != nil {
		return err
	}

	action, value := claudesettings.ReconcileStatusLine(settings, installStatuslineCommand)

	switch action {
	case claudesettings.StatusLineSet, claudesettings.StatusLineUpdated:
		if err := settings.Set("statusLine", value); err != nil {
			return err
		}
		return claudesettings.Save(installStatuslineSettings, settings)
	case claudesettings.StatusLineNoop:
		// Ours and byte-identical: nothing to write.
		return nil
	case claudesettings.StatusLineNeedsDecision:
		// A foreign statusLine is present. The destructive overwrite (with .bak)
		// is Phase 4 and the full compose/manual guidance text is Phase 2; this
		// step only wires the resolved decision flag through.
		switch installStatuslineOnForeign {
		case onForeignKeep:
			// Non-destructive: leave the foreign value untouched.
			return nil
		case onForeignOverwrite:
			// Phase 4 implements the real .bak + overwrite. Stub for now: do not
			// perform a destructive write here.
			return nil
		}
	}
	return nil
}
