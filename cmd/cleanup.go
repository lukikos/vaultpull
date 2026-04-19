package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/your-org/vaultpull/internal/audit"
)

var retentionDays int

var cleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Remove audit log entries older than the retention period",
	RunE:  runCleanup,
}

func init() {
	cleanupCmd.Flags().IntVar(&retentionDays, "retention-days", 30, "Number of days to retain audit log entries")
	rootCmd.AddCommand(cleanupCmd)
}

func runCleanup(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	retention := time.Duration(retentionDays) * 24 * time.Hour

	removed, err := audit.Cleanup(cfg.AuditLogPath, retention)
	if err != nil {
		return fmt.Errorf("cleanup failed: %w", err)
	}

	if removed == 0 {
		fmt.Println("No entries removed.")
	} else {
		fmt.Printf("Removed %d audit log entr%s older than %d day(s).\n",
			removed, pluralSuffix(removed, "y", "ies"), retentionDays)
	}
	return nil
}

func pluralSuffix(n int, singular, plural string) string {
	if n == 1 {
		return singular
	}
	return plural
}
