package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/yourusername/vaultpull/internal/audit"
)

func init() {
	baselineCmd := &cobra.Command{
		Use:   "baseline",
		Short: "Manage secret baselines for drift detection",
	}

	saveCmd := &cobra.Command{
		Use:   "save <name>",
		Short: "Save current secrets as a named baseline",
		Args:  cobra.ExactArgs(1),
		RunE:  runBaselineSave,
	}

	showCmd := &cobra.Command{
		Use:   "show <name>",
		Short: "Show diff between a baseline and a second baseline or current log",
		Args:  cobra.ExactArgs(1),
		RunE:  runBaselineShow,
	}

	baselineCmd.AddCommand(saveCmd, showCmd)
	rootCmd.AddCommand(baselineCmd)
}

func runBaselineSave(cmd *cobra.Command, args []string) error {
	name := args[0]
	entries, err := audit.ReadAll(".")
	if err != nil {
		return fmt.Errorf("read audit log: %w", err)
	}
	if len(entries) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "no audit entries found; nothing to baseline")
		return nil
	}
	last := entries[len(entries)-1]
	if err := audit.SaveBaseline(".", name, last.Path, last.Keys); err != nil {
		return fmt.Errorf("save baseline: %w", err)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "baseline %q saved (%d keys from %s)\n", name, len(last.Keys), last.Path)
	return nil
}

func runBaselineShow(cmd *cobra.Command, args []string) error {
	name := args[0]
	b, err := audit.LoadBaseline(".", name)
	if err != nil {
		return err
	}
	entries, err := audit.ReadAll(".")
	if err != nil {
		return fmt.Errorf("read audit log: %w", err)
	}
	var current map[string]string
	if len(entries) > 0 {
		current = entries[len(entries)-1].Keys
	} else {
		current = map[string]string{}
	}
	diffs := audit.BaselineDiff(b, current)
	if len(diffs) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "no differences")
		return nil
	}
	w := cmd.OutOrStdout()
	for _, d := range diffs {
		fmt.Fprintf(w, "  %-10s %s\n", d.Action, d.Key)
	}
	fmt.Fprintln(w, audit.SummarizeBaselineDiff(diffs))
	os.Exit(0)
	return nil
}
