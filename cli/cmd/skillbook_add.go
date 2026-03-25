package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/richardriman/crafter/cli/internal/skillbook"
)

var (
	addAgent     string
	addRule      string
	addRationale string
	addTask      string
)

var skillbookAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add or merge a skill into the skillbook",
	RunE:  runSkillbookAdd,
}

func init() {
	skillbookAddCmd.Flags().StringVar(&addAgent, "agent", "", "agent name (e.g. implementer)")
	skillbookAddCmd.Flags().StringVar(&addRule, "rule", "", "learned guideline text")
	skillbookAddCmd.Flags().StringVar(&addRationale, "rationale", "", "why this was observed")
	skillbookAddCmd.Flags().StringVar(&addTask, "task", "", "source task filename")

	_ = skillbookAddCmd.MarkFlagRequired("agent")
	_ = skillbookAddCmd.MarkFlagRequired("rule")
	_ = skillbookAddCmd.MarkFlagRequired("rationale")
	_ = skillbookAddCmd.MarkFlagRequired("task")

	skillbookCmd.AddCommand(skillbookAddCmd)
}

func runSkillbookAdd(cmd *cobra.Command, args []string) error {
	// 1. Load skillbook (create if not exists).
	sb, err := skillbook.Load(skillbookFile)
	if err != nil {
		return err
	}

	// 2. If disabled, print message and exit 0.
	if sb.Disabled {
		fmt.Fprintln(os.Stderr, "skillbook is disabled")
		return nil
	}

	// 3. Find duplicate with threshold 0.6.
	idx, found := skillbook.FindDuplicate(sb.Skills, addAgent, addRule, 0.6)

	if found {
		// 4. Merge: append rationale, promote confidence.
		s := &sb.Skills[idx]
		s.Rationale = s.Rationale + "; " + addRationale
		switch s.Confidence {
		case "low":
			s.Confidence = "medium"
		case "medium":
			s.Confidence = "high"
		// "high" stays high
		}
		fmt.Println("Merged with existing skill:", s.ID)
	} else {
		// 5. Create new skill.
		id, err := skillbook.NewID()
		if err != nil {
			return err
		}
		s := skillbook.Skill{
			ID:           id,
			Agent:        addAgent,
			Rule:         addRule,
			Rationale:    addRationale,
			SourceTask:   addTask,
			Confidence:   "low",
			AddedAt:      time.Now().UTC().Format(time.RFC3339),
			AppliedCount: 0,
		}
		sb.Skills = append(sb.Skills, s)
		fmt.Println("Added new skill:", id)
	}

	// 6. Atomic save.
	return skillbook.Save(skillbookFile, sb)
}
