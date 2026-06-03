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
//
// Pointer types are used where null/absent must be distinguishable from a real
// zero value: context_window.used_percentage is number|null (nil → omit the ctx
// section, not "0%"), and cost.total_cost_usd is absent before the first API
// call (nil and *0 both omit the cost section; a positive value renders).
type statuslineInput struct {
	Model struct {
		DisplayName string `json:"display_name"`
	} `json:"model"`
	Effort struct {
		Level string `json:"level"`
	} `json:"effort"`
	ContextWindow struct {
		UsedPercentage    *float64 `json:"used_percentage"`
		ContextWindowSize int      `json:"context_window_size"`
	} `json:"context_window"`
	Cost struct {
		TotalCostUSD      *float64 `json:"total_cost_usd"`
		TotalLinesAdded   int      `json:"total_lines_added"`
		TotalLinesRemoved int      `json:"total_lines_removed"`
	} `json:"cost"`
	Workspace struct {
		CurrentDir string `json:"current_dir"`
		ProjectDir string `json:"project_dir"`
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

	segment := statusline.RenderPanel(statusline.Payload{
		Workdir:           workdir,
		ModelDisplayName:  input.Model.DisplayName,
		EffortLevel:       input.Effort.Level,
		UsedPercentage:    input.ContextWindow.UsedPercentage,
		ContextWindowSize: input.ContextWindow.ContextWindowSize,
		TotalCostUSD:      input.Cost.TotalCostUSD,
		TotalLinesAdded:   input.Cost.TotalLinesAdded,
		TotalLinesRemoved: input.Cost.TotalLinesRemoved,
		ProjectDir:        input.Workspace.ProjectDir,
	})
	if segment != "" {
		fmt.Println(segment)
	}
	return nil
}
