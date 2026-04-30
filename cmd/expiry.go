package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/your-org/vaultpull/internal/audit"
)

func init() {
	var warnDays int

	cmd := &cobra.Command{
		Use:   "expiry",
		Short: "Check for secrets that are expired or expiring soon",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runExpiry(warnDays)
		},
	}

	cmd.Flags().IntVar(&warnDays, "warn-days", 30, "warn if a secret expires within this many days")
	rootCmd.AddCommand(cmd)
}

func runExpiry(warnDays int) error {
	entries, err := audit.ReadAll(".vaultpull-audit.log")
	if err != nil {
		return fmt.Errorf("reading audit log: %w", err)
	}

	cfg := audit.ExpiryConfig{WarnWithinDays: warnDays}
	results := audit.CheckExpiry(entries, cfg)

	if len(results) == 0 {
		fmt.Println("✓ No expiring or expired secrets detected.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "STATUS\tPATH\tKEY\tEXPIRES AT\tDAYS LEFT")

	for _, r := range results {
		status := "⚠ expiring"
		if r.Expired {
			status = "✗ expired"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\n",
			status,
			r.Path,
			r.Key,
			r.ExpiresAt.Format("2006-01-02 15:04 UTC"),
			r.DaysLeft,
		)
	}

	w.Flush()
	return nil
}
