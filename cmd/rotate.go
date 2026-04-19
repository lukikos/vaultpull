package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/user/vaultpull/internal/audit"
)

var rotateCmd = &cobra.Command{
	Use:   "rotate",
	Short: "Archive the current audit log and start fresh",
	RunE:  runRotate,
}

func init() {
	rootCmd.AddCommand(rotateCmd)
	rotateCmd.Flags().String("log", ".vaultpull-audit.log", "path to audit log file")
}

func runRotate(cmd *cobra.Command, _ []string) error {
	logPath, err := cmd.Flags().GetString("log")
	if err != nil {
		return err
	}

	archived, err := audit.Rotate(logPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return err
	}

	if archived == "" {
		fmt.Println("No audit log found — nothing to rotate.")
		return nil
	}

	fmt.Printf("Audit log rotated → %s\n", archived)
	return nil
}
