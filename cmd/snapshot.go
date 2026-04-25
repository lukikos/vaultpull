package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/yourorg/vaultpull/internal/audit"
	"github.com/yourorg/vaultpull/internal/config"
	"github.com/yourorg/vaultpull/internal/vault"
)

var snapshotCmd = &cobra.Command{
	Use:   "snapshot",
	Short: "Manage named snapshots of Vault secrets",
}

var snapshotSaveCmd = &cobra.Command{
	Use:   "save <name>",
	Short: "Save current Vault secrets as a named snapshot",
	Args:  cobra.ExactArgs(1),
	RunE:  runSnapshotSave,
}

var snapshotShowCmd = &cobra.Command{
	Use:   "show <name>",
	Short: "Display secrets stored in a named snapshot",
	Args:  cobra.ExactArgs(1),
	RunE:  runSnapshotShow,
}

func init() {
	snapshotCmd.AddCommand(snapshotSaveCmd)
	snapshotCmd.AddCommand(snapshotShowCmd)
	rootCmd.AddCommand(snapshotCmd)
}

func runSnapshotSave(cmd *cobra.Command, args []string) error {
	name := args[0]
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	client, err := vault.NewClient(cfg)
	if err != nil {
		return fmt.Errorf("vault client: %w", err)
	}
	secrets, err := client.GetSecrets()
	if err != nil {
		return fmt.Errorf("fetch secrets: %w", err)
	}
	dir := "."
	if err := audit.SaveSnapshot(dir, name, cfg.SecretPath, secrets); err != nil {
		return fmt.Errorf("save snapshot: %w", err)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Snapshot %q saved (%d keys)\n", name, len(secrets))
	return nil
}

func runSnapshotShow(cmd *cobra.Command, args []string) error {
	name := args[0]
	dir := "."
	s, err := audit.LoadSnapshot(dir, name)
	if err != nil {
		return err
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "Snapshot:\t%s\n", s.Name)
	fmt.Fprintf(w, "Path:\t%s\n", s.Path)
	fmt.Fprintf(w, "Created:\t%s\n", s.CreatedAt.Format("2006-01-02 15:04:05 UTC"))
	fmt.Fprintf(w, "Keys:\t%d\n", len(s.Secrets))
	w.Flush()
	fmt.Fprintln(os.Stdout)
	for k := range s.Secrets {
		fmt.Fprintf(os.Stdout, "  %s\n", k)
	}
	return nil
}
