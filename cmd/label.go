package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/your-org/vaultpull/internal/audit"
)

func init() {
	labelCmd := &cobra.Command{
		Use:   "label",
		Short: "Manage labels attached to audit log entries",
	}

	saveCmd := &cobra.Command{
		Use:   "save <name>",
		Short: "Attach a label to a secret key in the audit log",
		Args:  cobra.ExactArgs(1),
		RunE:  runLabelSave,
	}
	saveCmd.Flags().String("path", "", "Vault secret path (required)")
	saveCmd.Flags().String("key", "", "Secret key name (required)")
	saveCmd.Flags().String("note", "", "Optional human-readable note")
	_ = saveCmd.MarkFlagRequired("path")
	_ = saveCmd.MarkFlagRequired("key")

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all saved labels",
		RunE:  runLabelList,
	}

	labelCmd.AddCommand(saveCmd, listCmd)
	rootCmd.AddCommand(labelCmd)
}

func runLabelSave(cmd *cobra.Command, args []string) error {
	name := args[0]
	path, _ := cmd.Flags().GetString("path")
	key, _ := cmd.Flags().GetString("key")
	note, _ := cmd.Flags().GetString("note")

	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working dir: %w", err)
	}

	if err := audit.SaveLabel(dir, name, path, key, note); err != nil {
		return fmt.Errorf("save label: %w", err)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Label %q saved for %s/%s\n", name, path, key)
	return nil
}

func runLabelList(cmd *cobra.Command, _ []string) error {
	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working dir: %w", err)
	}

	labels, err := audit.LoadLabels(dir)
	if err != nil {
		return fmt.Errorf("load labels: %w", err)
	}
	if len(labels) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No labels found.")
		return nil
	}

	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tPATH\tKEY\tNOTE\tCREATED")
	for _, l := range labels {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			l.Name, l.Path, l.Key, l.Note, l.CreatedAt.Format("2006-01-02 15:04:05"))
	}
	return w.Flush()
}
