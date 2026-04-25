package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/yourusername/vaultpull/internal/audit"
)

var compareCmd = &cobra.Command{
	Use:   "compare <from-snapshot> <to-snapshot>",
	Short: "Compare two named snapshots and show key-level differences",
	Args:  cobra.ExactArgs(2),
	RunE:  runCompare,
}

func init() {
	rootCmd.AddCommand(compareCmd)
}

func runCompare(cmd *cobra.Command, args []string) error {
	fromName := args[0]
	toName := args[1]

	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	cmp, err := audit.CompareSnapshots(dir, fromName, toName)
	if err != nil {
		return fmt.Errorf("compare snapshots: %w", err)
	}

	fmt.Fprintln(cmd.OutOrStdout(), cmp.Summary())

	if len(cmp.Diffs) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No differences found.")
		return nil
	}

	fmt.Fprintf(cmd.OutOrStdout(), "\n%-30s %s\n", "KEY", "ACTION")
	fmt.Fprintf(cmd.OutOrStdout(), "%-30s %s\n", "---", "------")
	for _, d := range cmp.Diffs {
		if d.Action == "unchanged" {
			continue
		}
		fmt.Fprintf(cmd.OutOrStdout(), "%-30s %s\n", d.Key, d.Action)
	}

	return nil
}
