package cmd

import "github.com/spf13/cobra"

var skillbookFile string

var skillbookCmd = &cobra.Command{
	Use:   "skillbook",
	Short: "Manage skillbook entries",
}

func init() {
	skillbookCmd.PersistentFlags().StringVar(&skillbookFile, "file", "", "path to skillbook.json (required)")
	skillbookCmd.MarkPersistentFlagRequired("file")

	rootCmd.AddCommand(skillbookCmd)
}
