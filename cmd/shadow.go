package cmd

import (
	"crypto/sha256"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/yourusername/vaultpull/internal/audit"
)

func init() {
	shadowCmd := &cobra.Command{
		Use:   "shadow",
		Short: "Manage secret value shadows for drift detection",
	}

	saveCmd := &cobra.Command{
		Use:   "save <name>",
		Short: "Save a shadow of current secret hashes",
		Args:  cobra.ExactArgs(1),
		RunE:  runShadowSave,
	}
	saveCmd.Flags().String("path", "", "Vault secret path to shadow (required)")
	_ = saveCmd.MarkFlagRequired("path")
	saveCmd.Flags().String("log", ".vaultpull.log", "Audit log file")

	checkCmd := &cobra.Command{
		Use:   "check <name>",
		Short: "Compare current secrets against a saved shadow",
		Args:  cobra.ExactArgs(1),
		RunE:  runShadowCheck,
	}
	checkCmd.Flags().String("path", "", "Vault secret path to check (required)")
	_ = checkCmd.MarkFlagRequired("path")
	checkCmd.Flags().String("log", ".vaultpull.log", "Audit log file")

	shadowCmd.AddCommand(saveCmd, checkCmd)
	rootCmd.AddCommand(shadowCmd)
}

func runShadowSave(cmd *cobra.Command, args []string) error {
	name := args[0]
	path, _ := cmd.Flags().GetString("path")
	logFile, _ := cmd.Flags().GetString("log")

	entries, err := audit.ReadAll(logFile)
	if err != nil {
		return fmt.Errorf("read log: %w", err)
	}

	state := audit.Replay(entries, path, time.Now())
	if len(state) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "no secrets found for path", path)
		return nil
	}

	var shadows []audit.ShadowEntry
	for k, v := range state {
		hash := fmt.Sprintf("%x", sha256.Sum256([]byte(v)))
		shadows = append(shadows, audit.ShadowEntry{
			Path:       path,
			Key:        k,
			ValueHash:  hash,
			RecordedAt: time.Now().UTC(),
		})
	}

	dir := "."
	if err := audit.SaveShadow(dir, name, shadows); err != nil {
		return fmt.Errorf("save shadow: %w", err)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "shadow %q saved (%d keys)\n", name, len(shadows))
	return nil
}

func runShadowCheck(cmd *cobra.Command, args []string) error {
	name := args[0]
	path, _ := cmd.Flags().GetString("path")
	logFile, _ := cmd.Flags().GetString("log")

	shadow, err := audit.LoadShadow(".", name)
	if err != nil {
		return fmt.Errorf("load shadow: %w", err)
	}

	entries, err := audit.ReadAll(logFile)
	if err != nil {
		return fmt.Errorf("read log: %w", err)
	}

	state := audit.Replay(entries, path, time.Now())
	current := make(map[string]string, len(state))
	for k, v := range state {
		current[k] = fmt.Sprintf("%x", sha256.Sum256([]byte(v)))
	}

	reports := audit.CompareShadow(path, current, shadow)
	if len(reports) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "no shadow entries for path", path)
		return nil
	}

	drifted := 0
	for _, r := range reports {
		if r.Status != "match" {
			fmt.Fprintf(cmd.OutOrStdout(), "[%s] %s — %s\n", r.Status, r.Key, r.Message)
			drifted++
		}
	}
	if drifted == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "all keys match shadow")
	} else {
		fmt.Fprintf(cmd.OutOrStdout(), "%d key(s) drifted from shadow\n", drifted)
		os.Exit(1)
	}
	return nil
}
