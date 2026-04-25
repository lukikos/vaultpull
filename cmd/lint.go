package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/yourorg/vaultpull/internal/audit"
)

var lintLogPath string

func init() {
	lintCmd := &cobra.Command{
		Use:   "lint",
		Short: "Check audit log for common issues",
		Long:  "Lint inspects the audit log for empty keys, unknown actions, and other anomalies.",
		RunE:  runLint,
	}

	lintCmd.Flags().StringVar(&lintLogPath, "log", "vaultpull-audit.log", "path to audit log file")
	rootCmd.AddCommand(lintCmd)
}

func runLint(cmd *cobra.Command, args []string) error {
	report, err := audit.Lint(lintLogPath)
	if err != nil {
		return fmt.Errorf("lint: %w", err)
	}

	if report.Clean {
		fmt.Fprintln(cmd.OutOrStdout(), "✔ audit log looks clean — no issues found")
		return nil
	}

	fmt.Fprintf(cmd.OutOrStdout(), "⚠ found %d issue(s):\n\n", report.Total)
	for _, w := range report.Warnings {
		fmt.Fprintf(cmd.OutOrStdout(), "  [%d] path=%-30s key=%-20s  %s\n",
			w.Index, w.Path, w.Key, w.Warning)
	}

	// Exit with non-zero status so CI pipelines can detect problems.
	os.Exit(1)
	return nil
}
