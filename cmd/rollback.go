package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/yourusername/vaultpull/internal/audit"
	"github.com/yourusername/vaultpull/internal/dotenv"
)

var rollbackCmd = &cobra.Command{
	Use:   "rollback <tag> <path>",
	Short: "Restore .env to the state captured at a named tag",
	Args:  cobra.ExactArgs(2),
	RunE:  runRollback,
}

var rollbackOutput string

func init() {
	rollbackCmd.Flags().StringVarP(&rollbackOutput, "output", "o", ".env", "output .env file path")
	rootCmd.AddCommand(rollbackCmd)
}

func runRollback(cmd *cobra.Command, args []string) error {
	tagName := args[0]
	path := args[1]

	auditDir := "."
	if d := os.Getenv("VAULTPULL_AUDIT_DIR"); d != "" {
		auditDir = d
	}

	result, err := audit.Rollback(auditDir, tagName, path)
	if err != nil {
		return fmt.Errorf("rollback failed: %w", err)
	}

	w, err := dotenv.NewWriter(rollbackOutput)
	if err != nil {
		return fmt.Errorf("open output file: %w", err)
	}

	merged, err := dotenv.Merge(rollbackOutput, result.Restored)
	if err != nil {
		return fmt.Errorf("merge: %w", err)
	}

	if err := w.Write(merged); err != nil {
		return fmt.Errorf("write .env: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(),
		"Rolled back %d key(s) for path %q to tag %q (captured %s)\n",
		len(result.Restored),
		result.Path,
		result.TagName,
		result.Timestamp.Format("2006-01-02 15:04:05"),
	)
	return nil
}
