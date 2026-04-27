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
	cmd := &cobra.Command{
		Use:   "anomaly",
		Short: "Detect paths with anomalous sync frequency",
		RunE:  runAnomaly,
	}
	rootCmd.AddCommand(cmd)
}

func runAnomaly(cmd *cobra.Command, args []string) error {
	entries, err := audit.ReadAll(auditDir())
	if err != nil {
		return fmt.Errorf("reading audit log: %w", err)
	}

	results := audit.DetectAnomalies(entries)
	if len(results) == 0 {
		fmt.Println("No anomalies detected.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "PATH\tLAST SEEN\tMEAN INTERVAL\tSTATUS\tREASON")

	for _, r := range results {
		status := "ok"
		reason := "-"
		if r.IsAnomaly {
			status = "ANOMALY"
			reason = r.Reason
		}
		meanStr := formatDuration(r.MeanInterval)
		lastStr := r.LastSeen.Format(time.RFC3339)
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", r.Path, lastStr, meanStr, status, reason)
	}
	w.Flush()
	return nil
}

func formatDuration(seconds float64) string {
	d := time.Duration(seconds) * time.Second
	if d < time.Minute {
		return fmt.Sprintf("%.0fs", seconds)
	}
	if d < time.Hour {
		return fmt.Sprintf("%.1fm", d.Minutes())
	}
	return fmt.Sprintf("%.1fh", d.Hours())
}
