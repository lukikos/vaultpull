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
	driftCmd := &cobra.Command{
		Use:   "drift",
		Short: "Detect secret paths that have not been synced recently",
		RunE:  runDrift,
	}
	driftCmd.Flags().Float64("max-age", 48, "Maximum acceptable hours since last sync")
	driftCmd.Flags().Int("min-syncs", 1, "Minimum sync count required")
	driftCmd.Flags().String("log", ".vaultpull-audit.log", "Path to audit log file")
	rootCmd.AddCommand(driftCmd)
}

func runDrift(cmd *cobra.Command, _ []string) error {
	maxAge, _ := cmd.Flags().GetFloat64("max-age")
	minSyncs, _ := cmd.Flags().GetInt("min-syncs")
	logPath, _ := cmd.Flags().GetString("log")

	entries, err := audit.ReadAll(logPath)
	if err != nil {
		return fmt.Errorf("reading audit log: %w", err)
	}

	cfg := audit.DriftConfig{
		MaxAgeHours:  maxAge,
		MinSyncCount: minSyncs,
	}

	results := audit.DetectDrift(entries, cfg, time.Now())
	if len(results) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No sync activity found in audit log.")
		return nil
	}

	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "PATH\tAGE (h)\tSYNCS\tSTATUS\tREASON")
	for _, r := range results {
		status := "OK"
		if r.HasDrift {
			status = "DRIFT"
		}
		fmt.Fprintf(w, "%s\t%.1f\t%d\t%s\t%s\n",
			r.Path, r.AgeHours, r.SyncCount, status, r.Reason)
	}
	w.Flush()

	for _, r := range results {
		if r.HasDrift {
			os.Exit(1)
		}
	}
	return nil
}
