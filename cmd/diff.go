package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/yourorg/vaultpull/internal/audit"
	"github.com/yourorg/vaultpull/internal/dotenv"
)

var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Show what would change if secrets were synced now",
	RunE:  runDiff,
}

func init() {
	rootCmd.AddCommand(diffCmd)
}

func runDiff(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	// Parse the current .env file as the "previous" state.
	previous, err := dotenv.ParseFile(cfg.OutputFile)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("reading existing env file: %w", err)
	}

	// Fetch live secrets from Vault.
	incoming, err := fetchSecrets(cfg)
	if err != nil {
		return fmt.Errorf("fetching secrets: %w", err)
	}

	changes := audit.Diff(previous, incoming)
	summary := audit.SummarizeDiff(changes)

	for _, c := range changes {
		switch c.Action {
		case "added":
			fmt.Fprintf(cmd.OutOrStdout(), "+ %s\n", c.Key)
		case "updated":
			fmt.Fprintf(cmd.OutOrStdout(), "~ %s\n", c.Key)
		case "unchanged":
			fmt.Fprintf(cmd.OutOrStdout(), "  %s\n", c.Key)
		}
	}

	fmt.Fprintf(cmd.OutOrStdout(), "\n%d added, %d updated, %d unchanged\n",
		summary.Added, summary.Updated, summary.Unchanged)

	return nil
}
