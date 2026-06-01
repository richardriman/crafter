package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/richardriman/crafter/cli/internal/statusline"
)

// statuslineInput mirrors the subset of the Claude Code statusLine JSON payload
// that this command needs. Unknown fields are ignored.
type statuslineInput struct {
	Workspace struct {
		CurrentDir string `json:"current_dir"`
	} `json:"workspace"`
}

var statuslineCmd = &cobra.Command{
	Use:          "statusline",
	Short:        "Render the current Crafter plan position as a status-bar segment",
	SilenceUsage: true,
	RunE:         runStatusline,
}

func init() {
	rootCmd.AddCommand(statuslineCmd)
}

func runStatusline(cmd *cobra.Command, args []string) error {
	raw, err := io.ReadAll(os.Stdin)
	if err != nil {
		// Silent-fail: never break the status bar on read errors.
		return nil
	}

	var input statuslineInput
	if len(raw) > 0 {
		// Ignore unmarshal errors — malformed or missing payload is acceptable.
		_ = json.Unmarshal(raw, &input)
	}

	workdir := input.Workspace.CurrentDir
	if workdir == "" {
		workdir, _ = os.Getwd()
	}

	segment := statusline.Render(workdir)
	if segment != "" {
		fmt.Println(segment)
	}
	return nil
}
