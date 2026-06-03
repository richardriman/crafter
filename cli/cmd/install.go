package cmd

import "github.com/spf13/cobra"

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Apply Crafter mutations to a Claude Code settings.json",
}

func init() {
	rootCmd.AddCommand(installCmd)
}
