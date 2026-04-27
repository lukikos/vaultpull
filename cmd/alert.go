package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/yourorg/vaultpull/internal/audit"
)

func init() {
	alertCmd := &cobra.Command{
		Use:   "alert",
		Short: "Check audit log for alerts based on error rate and sync staleness",
		RunE:  runAlert,
	}
	alertCmd.Flags().Float64("max-error-rate", 0.1, "Maximum allowed error rate (0.0–1.0) before a critical alert")
	alertCmd.Flags().Duration("stale-after", audit.DefaultAlertConfig().StaleAfter, "Duration after which a path is considered stale")
	alertCmd.Flags().Duration("max-sync-interval", audit.DefaultAlertConfig().MaxSyncInterval, "Duration after which a warning is raised for no sync")
	rootCmd.AddCommand(alertCmd)
}

func runAlert(cmd *cobra.Command, _ []string) error {
	logPath, _ := cmd.Flags().GetString("log")
	if logPath == "" {
		logPath = ".vaultpull-audit.log"
	}

	entries, err := audit.ReadAll(logPath)
	if err != nil {
		return fmt.Errorf("reading audit log: %w", err)
	}

	maxErrorRate, _ := cmd.Flags().GetFloat64("max-error-rate")
	staleAfter, _ := cmd.Flags().GetDuration("stale-after")
	maxSyncInterval, _ := cmd.Flags().GetDuration("max-sync-interval")

	cfg := audit.AlertConfig{
		MaxErrorRate:    maxErrorRate,
		StaleAfter:      staleAfter,
		MaxSyncInterval: maxSyncInterval,
	}

	alerts := audit.CheckAlerts(entries, cfg)
	if len(alerts) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No alerts — all paths look healthy.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "LEVEL\tPATH\tMESSAGE")
	for _, a := range alerts {
		fmt.Fprintf(w, "%s\t%s\t%s\n", a.Level, a.Path, a.Message)
	}
	w.Flush()

	// Exit with non-zero status if any critical alerts exist.
	for _, a := range alerts {
		if a.Level == audit.AlertCritical {
			os.Exit(1)
		}
	}
	return nil
}
