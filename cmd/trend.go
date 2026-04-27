package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/yourorg/vaultpull/internal/audit"
)

var trendDays int

func init() {
	trendCmd := &cobra.Command{
		Use:   "trend",
		Short: "Show sync frequency trends per secret path",
		RunE:  runTrend,
	}
	trendCmd.Flags().IntVar(&trendDays, "days", 30, "Number of past days to analyse")
	rootCmd.AddCommand(trendCmd)
}

func runTrend(cmd *cobra.Command, args []string) error {
	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("could not determine working directory: %w", err)
	}

	entries, err := audit.ReadAll(dir)
	if err != nil {
		return fmt.Errorf("reading audit log: %w", err)
	}

	since := time.Now().UTC().AddDate(0, 0, -trendDays)
	results := audit.Trend(entries, since)

	if len(results) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No sync activity found in the specified period.")
		return nil
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Sync trend (last %d days)\n", trendDays)
	fmt.Fprintln(cmd.OutOrStdout(), "")

	for _, r := range results {
		direction := "→ stable"
		if r.TrendSlope > 0.1 {
			direction = "↑ increasing"
		} else if r.TrendSlope < -0.1 {
			direction = "↓ decreasing"
		}

		fmt.Fprintf(cmd.OutOrStdout(),
			"  %s\n    avg/day: %.2f  peak: %d (%s)  trend: %s\n",
			r.Path,
			r.AvgPerDay,
			r.PeakCount,
			r.PeakDate.Format("2006-01-02"),
			direction,
		)
	}
	return nil
}
