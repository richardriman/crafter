package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/richardriman/crafter/cli/internal/skillbook"
)

var getAgent string
var getLimit int

var skillbookGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Print learned guidelines for an agent",
	RunE:  runSkillbookGet,
}

func init() {
	skillbookGetCmd.Flags().StringVar(&getAgent, "agent", "", "agent name to retrieve skills for (required)")
	skillbookGetCmd.Flags().IntVar(&getLimit, "limit", 10, "maximum number of skills to return")

	_ = skillbookGetCmd.MarkFlagRequired("agent")

	skillbookCmd.AddCommand(skillbookGetCmd)
}

func runSkillbookGet(cmd *cobra.Command, args []string) error {
	// 1. Check file existence explicitly — if absent the user hasn't opted in.
	if _, err := os.Stat(skillbookFile); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("checking skillbook file: %w", err)
	}

	// 2. Load skillbook.
	sb, err := skillbook.Load(skillbookFile)
	if err != nil {
		return err
	}

	// 3. If disabled, exit 0 silently.
	if sb.Disabled {
		return nil
	}

	// 4. Format and print.
	output := skillbook.FormatMarkdown(sb.Skills, getAgent, getLimit)
	if output == "" {
		return nil
	}
	fmt.Print(output)

	// 5. Increment appliedCount and save.
	skillbook.IncrementApplied(sb, getAgent, getLimit)
	return skillbook.Save(skillbookFile, sb)
}
