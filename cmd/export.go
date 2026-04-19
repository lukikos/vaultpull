package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"vaultpull/internal/audit"
)

var exportOutput string

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export audit log to CSV",
	RunE:  runExport,
}

func init() {
	exportCmd.Flags().StringVarP(&exportOutput, "output", "o", "audit.csv", "Output CSV file path")
	rootCmd.AddCommand(exportCmd)
}

func runExport(cmd *cobra.Command, args []string) error {
	entries, err := audit.ReadAll(auditLog)
	if err != nil {
		return fmt.Errorf("reading audit log: %w", err)
	}

	if len(entries) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No audit entries found.")
		return nil
	}

	f, err := os.Create(exportOutput)
	if err != nil {
		return fmt.Errorf("creating output file: %w", err)
	}
	defer f.Close()

	if err := audit.ExportCSV(f, entries); err != nil {
		return fmt.Errorf("exporting CSV: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Exported %d entries to %s\n", len(entries), exportOutput)
	return nil
}
