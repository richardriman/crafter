package cmd

import (
	"github.com/spf13/cobra"

	"github.com/richardriman/crafter/cli/internal/buffer"
)

var (
	uatRunDir    string
	uatTitle     string
	uatSource    string
	uatCreatedBy string
	uatTaskID    string
	uatVerify    string
	uatWhyManual string
)

var bufferUATCmd = &cobra.Command{
	Use:          "uat",
	Short:        "Append a UAT (manual-QA) entry to uat-buffer.jsonl",
	SilenceUsage: true,
	RunE:         runBufferUAT,
}

func init() {
	bufferUATCmd.Flags().StringVar(&uatRunDir, "run-dir", "", "path to the per-run directory (required)")
	bufferUATCmd.Flags().StringVar(&uatTitle, "title", "", "short human-readable headline (required)")
	bufferUATCmd.Flags().StringVar(&uatSource, "source", "", "originating site, e.g. file:line or phase:step (required)")
	bufferUATCmd.Flags().StringVar(&uatCreatedBy, "created-by", "", "identity of the calling agent (required)")
	bufferUATCmd.Flags().StringVar(&uatTaskID, "task-id", "", "originating task ID, e.g. task-file basename without extension (required)")
	bufferUATCmd.Flags().StringVar(&uatVerify, "verify", "", "what to manually verify (required)")
	bufferUATCmd.Flags().StringVar(&uatWhyManual, "why-manual", "", "why the verification cannot be automated (required)")

	_ = bufferUATCmd.MarkFlagRequired("run-dir")
	_ = bufferUATCmd.MarkFlagRequired("title")
	_ = bufferUATCmd.MarkFlagRequired("source")
	_ = bufferUATCmd.MarkFlagRequired("created-by")
	_ = bufferUATCmd.MarkFlagRequired("task-id")
	_ = bufferUATCmd.MarkFlagRequired("verify")
	_ = bufferUATCmd.MarkFlagRequired("why-manual")

	bufferCmd.AddCommand(bufferUATCmd)
}

func runBufferUAT(cmd *cobra.Command, args []string) error {
	id, err := buffer.NewID()
	if err != nil {
		return err
	}

	entry := buffer.UATEntry{
		ID:        id,
		Kind:      "uat",
		CreatedAt: buffer.NowUTC(),
		CreatedBy: uatCreatedBy,
		TaskID:    uatTaskID,
		Title:     uatTitle,
		Source:    uatSource,
		Verify:    uatVerify,
		WhyManual: uatWhyManual,
	}

	return buffer.AppendUAT(uatRunDir, entry)
}
