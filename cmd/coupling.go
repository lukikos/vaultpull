package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/your-org/vaultpull/internal/audit"
)

func init() {
	var minSupport float64
	var logPath string

	cmd := &cobra.Command{
		Use:   "coupling",
		Short: "Show secret paths that are frequently synced together",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCoupling(logPath, minSupport)
		},
	}

	cmd.Flags().Float64Var(&minSupport, "min-support", 0.5,
		"Minimum fraction of sync batches where both paths must appear (0–1)")
	cmd.Flags().StringVar(&logPath, "log", ".vaultpull-audit.log",
		"Path to the audit log file")

	rootCmd.AddCommand(cmd)
}

func runCoupling(logPath string, minSupport float64) error {
	entries, err := audit.ReadAll(logPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("reading audit log: %w", err)
	}

	results := audit.DetectCoupling(entries, minSupport)
	if len(results) == 0 {
		fmt.Println("No coupled paths found at the given support threshold.")
		return nil
	}

	fmt.Printf("%-30s  %-30s  %10s  %8s\n", "PATH A", "PATH B", "CO-OCCURS", "SUPPORT")
	fmt.Printf("%-30s  %-30s  %10s  %8s\n",
		"------------------------------",
		"------------------------------",
		"----------",
		"--------")
	for _, r := range results {
		fmt.Printf("%-30s  %-30s  %10d  %7.1f%%\n",
			r.PathA, r.PathB, r.CoOccurs, r.Support*100)
	}
	return nil
}
