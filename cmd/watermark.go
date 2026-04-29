package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"github.com/yourusername/vaultpull/internal/audit"
)

func init() {
	watermarkCmd := &cobra.Command{
		Use:   "watermark",
		Short: "Show the latest sync timestamp (high-water mark) for each secret path",
		RunE:  runWatermark,
	}
	watermarkCmd.Flags().String("log", ".vaultpull_audit.log", "Path to audit log file")
	watermarkCmd.Flags().Bool("save", false, "Persist computed watermarks to disk")
	RootCmd.AddCommand(watermarkCmd)
}

func runWatermark(cmd *cobra.Command, _ []string) error {
	logPath, _ := cmd.Flags().GetString("log")
	save, _ := cmd.Flags().GetBool("save")

	entries, err := audit.ReadAll(logPath)
	if err != nil {
		return fmt.Errorf("reading audit log: %w", err)
	}

	watermarks := audit.ComputeWatermarks(entries)
	if len(watermarks) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No sync entries found.")
		return nil
	}

	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "PATH\tLATEST SYNC\tSYNC COUNT")
	for _, wm := range watermarks {
		age := time.Since(wm.LatestAt).Truncate(time.Second)
		fmt.Fprintf(w, "%s\t%s (%s ago)\t%d\n",
			wm.Path,
			wm.LatestAt.Format(time.RFC3339),
			age,
			wm.SyncCount,
		)
	}
	w.Flush()

	if save {
		dir := "."
		if err := audit.SaveWatermarks(dir, watermarks); err != nil {
			return fmt.Errorf("saving watermarks: %w", err)
		}
		fmt.Fprintln(cmd.OutOrStdout(), "Watermarks saved.")
	}

	_ = os.Stderr // satisfy import if needed
	return nil
}
