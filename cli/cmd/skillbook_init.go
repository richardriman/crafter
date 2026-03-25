package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/richardriman/crafter/cli/internal/skillbook"
)

var skillbookInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Create an empty skillbook file",
	RunE:  runSkillbookInit,
}

func init() {
	skillbookCmd.AddCommand(skillbookInitCmd)
}

func runSkillbookInit(cmd *cobra.Command, args []string) error {
	// 1. If file already exists, report and exit 0.
	if _, err := os.Stat(skillbookFile); err == nil {
		fmt.Fprintf(os.Stderr, "Skillbook already exists at %s\n", skillbookFile)
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("checking skillbook file: %w", err)
	}

	// 2. Create parent directories if needed.
	dir := filepath.Dir(skillbookFile)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating directories: %w", err)
	}

	// 3. Create new empty skillbook and save.
	sb := skillbook.NewSkillbook()
	if err := skillbook.Save(skillbookFile, sb); err != nil {
		return err
	}

	// 4. Confirm.
	fmt.Printf("Created empty skillbook at %s\n", skillbookFile)
	return nil
}
