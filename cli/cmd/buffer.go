package cmd

import "github.com/spf13/cobra"

var bufferCmd = &cobra.Command{
	Use:   "buffer",
	Short: "Append entries to the per-run UAT or Gap buffer",
}

func init() {
	rootCmd.AddCommand(bufferCmd)
}
