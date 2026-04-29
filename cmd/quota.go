package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/yourusername/vaultpull/internal/audit"
)

func init() {
	saveCmd := &cobra.Command{
		Use:   "save",
		Short: "Save a quota policy",
		RunE:  runQuotaSave,
	}
	saveCmd.Flags().Int("max-syncs-per-hour", 0, "Max syncs allowed per hour (0 = unlimited)")
	saveCmd.Flags().Int("max-keys-per-sync", 0, "Max keys allowed per sync (0 = unlimited)")
	saveCmd.Flags().Int("max-paths-per-day", 0, "Max unique paths allowed per day (0 = unlimited)")

	checkCmd := &cobra.Command{
		Use:   "check",
		Short: "Check current audit log against quota policy",
		RunE:  runQuotaCheck,
	}

	quotaCmd := &cobra.Command{
		Use:   "quota",
		Short: "Manage and enforce sync quota policies",
	}
	quotaCmd.AddCommand(saveCmd, checkCmd)
	rootCmd.AddCommand(quotaCmd)
}

func runQuotaSave(cmd *cobra.Command, _ []string) error {
	dir, _ := os.Getwd()
	p := audit.QuotaPolicy{}
	var err error
	if v, _ := cmd.Flags().GetInt("max-syncs-per-hour"); v != 0 {
		p.MaxSyncsPerHour = v
	}
	if v, _ := cmd.Flags().GetInt("max-keys-per-sync"); v != 0 {
		p.MaxKeysPerSync = v
	}
	if v, _ := cmd.Flags().GetInt("max-paths-per-day"); v != 0 {
		p.MaxPathsPerDay = v
	}
	if err = audit.SaveQuotaPolicy(dir, p); err != nil {
		return fmt.Errorf("saving quota policy: %w", err)
	}
	fmt.Println("Quota policy saved.")
	return nil
}

func runQuotaCheck(cmd *cobra.Command, _ []string) error {
	dir, _ := os.Getwd()
	p, err := audit.LoadQuotaPolicy(dir)
	if err != nil {
		return fmt.Errorf("loading quota policy: %w", err)
	}
	entries, err := audit.ReadAll(dir)
	if err != nil {
		return fmt.Errorf("reading audit log: %w", err)
	}
	result := audit.CheckQuota(entries, p)
	if result.OK() {
		fmt.Println("✓ All quota limits satisfied.")
		return nil
	}
	fmt.Fprintf(os.Stderr, "✗ Quota violations detected (%s):\n",
		strconv.Itoa(len(result.Violations)))
	for _, v := range result.Violations {
		fmt.Fprintf(os.Stderr, "  - %s\n", v)
	}
	os.Exit(1)
	return nil
}
