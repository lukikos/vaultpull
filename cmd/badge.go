package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/yourusername/vaultpull/internal/audit"
)

var badgeJSONOutput bool

func init() {
	badgeCmd := &cobra.Command{
		Use:   "badge",
		Short: "Show sync health badges for each vault path",
		RunE:  runBadge,
	}
	badgeCmd.Flags().BoolVar(&badgeJSONOutput, "json", false, "Output badges as JSON")
	rootCmd.AddCommand(badgeCmd)
}

func runBadge(cmd *cobra.Command, args []string) error {
	entries, err := audit.ReadAll(auditDir())
	if err != nil {
		return fmt.Errorf("reading audit log: %w", err)
	}

	cfg := audit.DefaultBadgeConfig()
	badges := audit.GenerateBadges(entries, cfg)

	if len(badges) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No audit entries found.")
		return nil
	}

	if badgeJSONOutput {
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(badges)
	}

	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "PATH\tSTATUS\tMESSAGE\tLAST SYNC")
	for _, b := range badges {
		lastSync := "never"
		if b.LastSync != nil {
			lastSync = b.LastSync.Format("2006-01-02 15:04:05")
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", b.Path, b.Status, b.Message, lastSync)
	}
	return w.Flush()
}

func badgeStatusIcon(status audit.BadgeStatus) string {
	switch status {
	case audit.BadgeStatusOK:
		return "✓"
	case audit.BadgeStatusWarning:
		return "!"
	case audit.BadgeStatusError:
		return "✗"
	default:
		return "?"
	}
}

// Ensure badgeStatusIcon is used (suppress unused warning in minimal builds).
var _ = badgeStatusIcon
var _ = os.Stderr
