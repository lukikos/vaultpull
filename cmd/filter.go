package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/yourorg/vaultpull/internal/audit"
)

var (
	filterPath   string
	filterAction string
	filterSince  string
	filterUntil  string
)

var filterCmd = &cobra.Command{
	Use:   "filter",
	Short: "Filter audit log entries by path, action, or time range",
	RunE:  runFilter,
}

func init() {
	filterCmd.Flags().StringVar(&filterPath, "path", "", "filter by secret path")
	filterCmd.Flags().StringVar(&filterAction, "action", "", "filter by action (write, skip, …)")
	filterCmd.Flags().StringVar(&filterSince, "since", "", "include entries at or after this RFC3339 time")
	filterCmd.Flags().StringVar(&filterUntil, "until", "", "include entries at or before this RFC3339 time")
	rootCmd.AddCommand(filterCmd)
}

func runFilter(cmd *cobra.Command, _ []string) error {
	logFile, _ := cmd.Root().PersistentFlags().GetString("log")

	entries, err := audit.ReadAll(logFile)
	if err != nil {
		return fmt.Errorf("reading audit log: %w", err)
	}

	opts := audit.FilterOptions{
		Path:   filterPath,
		Action: filterAction,
	}

	if filterSince != "" {
		t, err := time.Parse(time.RFC3339, filterSince)
		if err != nil {
			return fmt.Errorf("invalid --since value: %w", err)
		}
		opts.Since = t
	}
	if filterUntil != "" {
		t, err := time.Parse(time.RFC3339, filterUntil)
		if err != nil {
			return fmt.Errorf("invalid --until value: %w", err)
		}
		opts.Until = t
	}

	matched := audit.Filter(entries, opts)
	if len(matched) == 0 {
		fmt.Fprintln(os.Stderr, "no entries matched the given filters")
		return nil
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(matched)
}
