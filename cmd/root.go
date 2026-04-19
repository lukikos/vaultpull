package cmd

import (
	"fmt"
	"os"

	"github.com/example/vaultpull/internal/config"
	"github.com/example/vaultpull/internal/sync"
	"github.com/spf13/cobra"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "vaultpull",
	Short: "Sync HashiCorp Vault secrets into local .env files",
	RunE:  runSync,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: .vaultpull.yaml)")
	rootCmd.Flags().String("output", ".env", "output .env file path")
	rootCmd.Flags().Bool("backup", true, "create a backup before overwriting")
	rootCmd.Flags().Bool("overwrite", false, "overwrite existing keys")
}

func runSync(cmd *cobra.Command, _ []string) error {
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if v, _ := cmd.Flags().GetString("output"); cmd.Flags().Changed("output") {
		cfg.OutputFile = v
	}
	if v, _ := cmd.Flags().GetBool("backup"); cmd.Flags().Changed("backup") {
		cfg.Backup = v
	}
	if v, _ := cmd.Flags().GetBool("overwrite"); cmd.Flags().Changed("overwrite") {
		cfg.Overwrite = v
	}

	s, err := sync.New(cfg)
	if err != nil {
		return err
	}

	result, err := s.Run()
	if err != nil {
		return err
	}

	fmt.Printf("✓ Synced %d secrets to %s", result.SecretsCount, result.OutputPath)
	if result.BackedUp {
		fmt.Printf(" (backup created)")
	}
	fmt.Println()
	return nil
}
