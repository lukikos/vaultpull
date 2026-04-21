package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/yourorg/vaultpull/internal/audit"
)

var pruneKeepTopN int

var pruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "Remove excess audit log entries, keeping the N most recent per secret path",
	RunE:  runPrune,
}

func init() {
	pruneCmd.Flags().IntVarP(
		&pruneKeepTopN, "keep", "k", 10,
		"Number of most recent entries to keep per secret path",
	)
	rootCmd.AddCommand(pruneCmd)
}

func runPrune(cmd *cobra.Command, args []string) error {
	logPath, err := cmd.Flags().GetString("log")
	if err != nil || logPath == "" {
		logPath = ".vaultpull-audit.log"
	}

	opts := audit.PruneOptions{KeepTopN: pruneKeepTopN}
	removed, err := audit.Prune(logPath, opts)
	if err != nil {
		return fmt.Errorf("prune failed: %w", err)
	}

	if removed == 0 {
		fmt.Println("Nothing to prune.")
		return nil
	}

	fmt.Printf(
		"Pruned %d %s from audit log (keeping top %d per path).\n",
		removed,
		pluralSuffix(removed, "entry", "entries"),
		pruneKeepTopN,
	)
	return nil
}
