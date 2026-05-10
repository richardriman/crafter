package cmd

import (
	"github.com/spf13/cobra"

	"github.com/richardriman/crafter/cli/internal/buffer"
)

var (
	gapRunDir    string
	gapTitle     string
	gapSource    string
	gapCreatedBy string
	gapTaskID    string
	gapDetail    string
	gapFollowup  string
)

var bufferGapCmd = &cobra.Command{
	Use:          "gap",
	Short:        "Append a Gap (tech-debt / deferred follow-up) entry to gaps-buffer.jsonl",
	SilenceUsage: true,
	RunE:         runBufferGap,
}

func init() {
	bufferGapCmd.Flags().StringVar(&gapRunDir, "run-dir", "", "path to the per-run directory (required)")
	bufferGapCmd.Flags().StringVar(&gapTitle, "title", "", "short human-readable headline (required)")
	bufferGapCmd.Flags().StringVar(&gapSource, "source", "", "originating site, e.g. file:line or phase:step (required)")
	bufferGapCmd.Flags().StringVar(&gapCreatedBy, "created-by", "", "identity of the calling agent (required)")
	bufferGapCmd.Flags().StringVar(&gapTaskID, "task-id", "", "originating task ID, e.g. task-file basename without extension (required)")
	bufferGapCmd.Flags().StringVar(&gapDetail, "detail", "", "description of the gap or tech debt (required)")
	bufferGapCmd.Flags().StringVar(&gapFollowup, "followup", "", "recommended action to close the gap (required)")

	_ = bufferGapCmd.MarkFlagRequired("run-dir")
	_ = bufferGapCmd.MarkFlagRequired("title")
	_ = bufferGapCmd.MarkFlagRequired("source")
	_ = bufferGapCmd.MarkFlagRequired("created-by")
	_ = bufferGapCmd.MarkFlagRequired("task-id")
	_ = bufferGapCmd.MarkFlagRequired("detail")
	_ = bufferGapCmd.MarkFlagRequired("followup")

	bufferCmd.AddCommand(bufferGapCmd)
}

func runBufferGap(cmd *cobra.Command, args []string) error {
	id, err := buffer.NewID()
	if err != nil {
		return err
	}

	entry := buffer.GapEntry{
		ID:        id,
		Kind:      "gap",
		CreatedAt: buffer.NowUTC(),
		CreatedBy: gapCreatedBy,
		TaskID:    gapTaskID,
		Title:     gapTitle,
		Source:    gapSource,
		Detail:    gapDetail,
		Followup:  gapFollowup,
	}

	return buffer.AppendGap(gapRunDir, entry)
}
