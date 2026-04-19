package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/yourorg/vaultpull/internal/audit"
)

var summaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "Print a summary of past sync operations from the audit log",
	RunE:  runSummary,
}

func init() {
	rootCmd.AddCommand(summaryCmd)
	summaryCmd.Flags().String("audit-log", ".vaultpull-audit.log", "Path to the audit log file")
}

func runSummary(cmd *cobra.Command, _ []string) error {
	logPath, err := cmd.Flags().GetString("audit-log")
	if err != nil {
		return err
	}

	entries, err := audit.ReadAll(logPath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No audit log found.")
			return nil
		}
		return fmt.Errorf("reading audit log: %w", err)
	}

	s := audit.Summarize(entries)
	fmt.Println(s)
	return nil
}
