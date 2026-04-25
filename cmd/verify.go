package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yourusername/vaultpull/internal/audit"
	"github.com/yourusername/vaultpull/internal/config"
)

func init() {
	verifyCmd := &cobra.Command{
		Use:   "verify",
		Short: "Verify local .env file matches the latest audit log state",
		RunE:  runVerify,
	}
	verifyCmd.Flags().String("env-file", ".env", "Path to the .env file to verify")
	verifyCmd.Flags().String("log-file", "vaultpull-audit.log", "Path to the audit log file")
	rootCmd.AddCommand(verifyCmd)
}

func runVerify(cmd *cobra.Command, _ []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	envFile, _ := cmd.Flags().GetString("env-file")
	logFile, _ := cmd.Flags().GetString("log-file")

	result, err := audit.Verify(logFile, cfg.SecretPath, envFile)
	if err != nil {
		return fmt.Errorf("verify failed: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Secret path : %s\n", result.Path)
	fmt.Fprintf(cmd.OutOrStdout(), "Env file    : %s\n", envFile)
	fmt.Fprintf(cmd.OutOrStdout(), "Status      : %s\n", statusLabel(result.OK))

	if len(result.MatchedKeys) > 0 {
		sort.Strings(result.MatchedKeys)
		fmt.Fprintf(cmd.OutOrStdout(), "Matched     : %s\n", strings.Join(result.MatchedKeys, ", "))
	}
	if len(result.MissingKeys) > 0 {
		sort.Strings(result.MissingKeys)
		fmt.Fprintf(cmd.OutOrStdout(), "Missing     : %s\n", strings.Join(result.MissingKeys, ", "))
	}
	if len(result.ExtraKeys) > 0 {
		sort.Strings(result.ExtraKeys)
		fmt.Fprintf(cmd.OutOrStdout(), "Extra       : %s\n", strings.Join(result.ExtraKeys, ", "))
	}

	if !result.OK {
		os.Exit(1)
	}
	return nil
}

func statusLabel(ok bool) string {
	if ok {
		return "OK"
	}
	return "DRIFT DETECTED"
}
