package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/yourusername/vaultpull/internal/audit"
)

func init() {
	redactCmd := &cobra.Command{
		Use:   "redact",
		Short: "Print audit log with sensitive values masked",
		RunE:  runRedact,
	}
	redactCmd.Flags().StringSlice("patterns", nil, "Additional key patterns to redact (comma-separated)")
	redactCmd.Flags().String("mask", "***", "Replacement string for redacted values")
	redactCmd.Flags().String("log", ".vaultpull-audit.log", "Path to audit log file")
	redactCmd.Flags().String("format", "text", "Output format: text or json")
	rootCmd.AddCommand(redactCmd)
}

func runRedact(cmd *cobra.Command, _ []string) error {
	logPath, _ := cmd.Flags().GetString("log")
	mask, _ := cmd.Flags().GetString("mask")
	patterns, _ := cmd.Flags().GetStringSlice("patterns")
	format, _ := cmd.Flags().GetString("format")

	entries, err := audit.ReadAll(logPath)
	if err != nil {
		return fmt.Errorf("reading audit log: %w", err)
	}

	var opts *audit.RedactOptions
	if len(patterns) > 0 || mask != "***" {
		opts = &audit.RedactOptions{Patterns: patterns, Mask: mask}
	}

	redacted := audit.Redact(entries, opts)

	switch strings.ToLower(format) {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(redacted)
	default:
		if len(redacted) == 0 {
			fmt.Println("No audit entries found.")
			return nil
		}
		for _, e := range redacted {
			fmt.Printf("%s  %-10s  %-30s  %s=%s\n",
				e.Timestamp.Format("2006-01-02T15:04:05Z"),
				e.Action, e.Path, e.Key, e.Value)
		}
	}
	return nil
}
