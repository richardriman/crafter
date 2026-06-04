package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/richardriman/crafter/cli/internal/claudesettings"
)

// printLines writes each line followed by a newline to cmd's stdout.
func printLines(cmd *cobra.Command, lines []string) {
	for _, l := range lines {
		fmt.Fprintln(cmd.OutOrStdout(), l)
	}
}

const (
	onForeignKeep      = "keep"
	onForeignOverwrite = "overwrite"
)

// rungTokens are the stdout tokens printed by --classify mode.
const (
	rungAbsent  = "absent"
	rungOurs    = "ours"
	rungForeign = "foreign"
)

var (
	installStatuslineSettings  string
	installStatuslineCommand   string
	installStatuslineOnForeign string
	installStatuslineClassify  bool
)

var installStatuslineCmd = &cobra.Command{
	Use:   "statusline",
	Short: "Reconcile the Crafter statusLine into a Claude Code settings.json",
	Long: "Non-interactively reconcile the Crafter status-line command into a " +
		"settings.json. The caller (install.sh) computes the command and, when a " +
		"foreign statusLine is present, resolves the keep-vs-overwrite decision and " +
		"passes it via --on-foreign; this command never prompts or reads a TTY.\n\n" +
		"With --classify, prints exactly one of 'absent', 'ours', or 'foreign' and " +
		"exits 0 without writing anything.",
	SilenceUsage: true,
	RunE:         runInstallStatusline,
}

func init() {
	installStatuslineCmd.Flags().StringVar(&installStatuslineSettings, "settings", "", "path to the target settings.json (required)")
	installStatuslineCmd.Flags().StringVar(&installStatuslineCommand, "command", "", "the computed Crafter statusline command to install (required)")
	installStatuslineCmd.Flags().StringVar(&installStatuslineOnForeign, "on-foreign", onForeignKeep, "what to do when a foreign statusLine is present: keep|overwrite")
	installStatuslineCmd.Flags().BoolVar(&installStatuslineClassify, "classify", false, "print the current statusLine rung (absent|ours|foreign) and exit without writing")

	_ = installStatuslineCmd.MarkFlagRequired("settings")
	_ = installStatuslineCmd.MarkFlagRequired("command")

	installCmd.AddCommand(installStatuslineCmd)
}

func runInstallStatusline(cmd *cobra.Command, args []string) error {
	// --classify mode: read the settings file, classify the statusLine rung using
	// the existing ClassifyStatusLine function, print one token to stdout, and exit
	// without writing anything. This lets install.sh learn the rung before deciding
	// whether to prompt.
	if installStatuslineClassify {
		settings, err := claudesettings.Load(installStatuslineSettings)
		if err != nil {
			return err
		}
		switch claudesettings.ClassifyStatusLine(settings) {
		case claudesettings.StatusLineAbsent:
			fmt.Fprintln(cmd.OutOrStdout(), rungAbsent)
		case claudesettings.StatusLineOurs:
			fmt.Fprintln(cmd.OutOrStdout(), rungOurs)
		default:
			fmt.Fprintln(cmd.OutOrStdout(), rungForeign)
		}
		return nil
	}

	if installStatuslineOnForeign != onForeignKeep && installStatuslineOnForeign != onForeignOverwrite {
		return fmt.Errorf("--on-foreign must be %q or %q, got %q", onForeignKeep, onForeignOverwrite, installStatuslineOnForeign)
	}

	// Read the raw on-disk bytes before Load so that the foreign-overwrite backup
	// captures the original file content verbatim — even if Load would treat it as
	// empty (garbled JSON). A nil rawBytes means the file did not exist.
	rawBytes, err := os.ReadFile(installStatuslineSettings)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("reading settings for backup: %w", err)
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
		switch installStatuslineOnForeign {
		case onForeignKeep:
			// Non-destructive: leave the foreign value untouched and print
			// the compose-wrapper / manual-merge guidance for the user.
			raw, _ := settings.Get("statusLine")
			lines, err := claudesettings.ForeignKeepGuidanceLines(installStatuslineSettings, raw, installStatuslineCommand)
			if err != nil {
				return err
			}
			printLines(cmd, lines)
			return nil
		case onForeignOverwrite:
			return runForeignOverwrite(cmd, settings, rawBytes)
		}
	}
	return nil
}

// runForeignOverwrite handles the --on-foreign=overwrite path for a foreign
// statusLine. It backs up the raw on-disk bytes (addressing Phase-1 findings #2
// and #3), echoes the old command to stdout, then overwrites statusLine with the
// freshly-computed crafter command.
//
// rawBytes is the content of the settings file read before Load was called; it
// may be nil if the file did not exist (in which case there is no foreign value
// to overwrite and this function is a no-op — this path is only reached when
// ReconcileStatusLine returned StatusLineNeedsDecision, which requires the key
// to be present).
func runForeignOverwrite(cmd *cobra.Command, settings *claudesettings.Settings, rawBytes []byte) error {
	// Retrieve the foreign statusLine raw value before we mutate settings.
	raw, _ := settings.Get("statusLine")

	// Back up the original on-disk bytes BEFORE any write, using the
	// content-aware scheme that addresses findings #2 and #3.
	if rawBytes != nil {
		if err := claudesettings.BackupForOverwrite(installStatuslineSettings, rawBytes); err != nil {
			return err
		}
	}

	// Echo the old command to stdout so it is recoverable by the caller.
	oldCmd := claudesettings.StatusLineCommandString(raw)
	fmt.Fprintln(cmd.OutOrStdout(), oldCmd)

	// Overwrite statusLine with the freshly-computed crafter command.
	newValue := claudesettings.NewStatusLineValue(installStatuslineCommand)
	if err := settings.Set("statusLine", newValue); err != nil {
		return err
	}
	return claudesettings.Save(installStatuslineSettings, settings)
}
