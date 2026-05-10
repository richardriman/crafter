package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/richardriman/crafter/cli/internal/prbody"
)

var (
	prBodyRunDir   string
	prBodyTaskFile string
)

var prBodyCmd = &cobra.Command{
	Use:          "pr-body",
	Short:        "Assemble the appended-sections Markdown block for a pull-request body",
	SilenceUsage: true,
	RunE:         runPRBody,
}

func init() {
	prBodyCmd.Flags().StringVar(&prBodyRunDir, "run-dir", "", "path to the per-run directory containing buffer files (required)")
	prBodyCmd.Flags().StringVar(&prBodyTaskFile, "task-file", "", "path to the task Markdown file containing the Decisions section (required)")

	_ = prBodyCmd.MarkFlagRequired("run-dir")
	_ = prBodyCmd.MarkFlagRequired("task-file")

	rootCmd.AddCommand(prBodyCmd)
}

func runPRBody(cmd *cobra.Command, args []string) error {
	out, err := prbody.Assemble(prBodyRunDir, prBodyTaskFile)
	if err != nil {
		return err
	}
	if out != "" {
		fmt.Println(out)
	}
	return nil
}
