package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/yourorg/vaultpull/internal/audit"
)

func init() {
	retentionCmd := &cobra.Command{
		Use:   "retention",
		Short: "Manage audit log retention policy",
	}

	saveCmd := &cobra.Command{
		Use:   "save <max-age-days> <max-entries>",
		Short: "Save a retention policy (use 0 to disable a limit)",
		Args:  cobra.ExactArgs(2),
		RunE:  runRetentionSave,
	}

	enforceCmd := &cobra.Command{
		Use:   "enforce",
		Short: "Apply the saved retention policy to the audit log",
		RunE:  runRetentionEnforce,
	}

	retentionCmd.AddCommand(saveCmd, enforceCmd)
	rootCmd.AddCommand(retentionCmd)
}

func runRetentionSave(cmd *cobra.Command, args []string) error {
	dir, err := auditDir(cmd)
	if err != nil {
		return err
	}
	maxAge, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid max-age-days: %w", err)
	}
	maxEntries, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("invalid max-entries: %w", err)
	}
	policy := audit.RetentionPolicy{MaxAgeDays: maxAge, MaxEntries: maxEntries}
	if err := audit.SaveRetentionPolicy(dir, policy); err != nil {
		return fmt.Errorf("save retention policy: %w", err)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Retention policy saved (max_age_days=%d, max_entries=%d)\n", maxAge, maxEntries)
	return nil
}

func runRetentionEnforce(cmd *cobra.Command, args []string) error {
	dir, err := auditDir(cmd)
	if err != nil {
		return err
	}
	policy, err := audit.LoadRetentionPolicy(dir)
	if err != nil {
		return fmt.Errorf("load retention policy: %w", err)
	}
	logPath := dir + "/audit.log"
	result, err := audit.EnforceRetention(logPath, policy)
	if err != nil {
		return fmt.Errorf("enforce retention: %w", err)
	}
	if result.Removed == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No entries removed; log is within policy.")
		return nil
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Removed %d %s, retained %d.\n",
		result.Removed, pluralSuffix("entry", "entries", result.Removed), result.Retained)
	return nil
}

func auditDir(cmd *cobra.Command) (string, error) {
	dir := os.Getenv("VAULTPULL_AUDIT_DIR")
	if dir == "" {
		dir = "."
	}
	return dir, nil
}
