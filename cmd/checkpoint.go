package cmd

import (
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/yourusername/vaultpull/internal/audit"
)

func init() {
	checkpointCmd := &cobra.Command{
		Use:   "checkpoint",
		Short: "Manage audit log checkpoints",
	}

	saveCmd := &cobra.Command{
		Use:   "save <name> <path> <offset>",
		Short: "Save a named checkpoint for a vault path at a log offset",
		Args:  cobra.ExactArgs(3),
		RunE:  runCheckpointSave,
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all saved checkpoints",
		RunE:  runCheckpointList,
	}

	checkpointCmd.AddCommand(saveCmd, listCmd)
	rootCmd.AddCommand(checkpointCmd)
}

func runCheckpointSave(cmd *cobra.Command, args []string) error {
	name, path := args[0], args[1]
	offset, err := strconv.Atoi(args[2])
	if err != nil {
		return fmt.Errorf("offset must be an integer: %w", err)
	}
	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	if err := audit.SaveCheckpoint(dir, name, path, offset); err != nil {
		return fmt.Errorf("save checkpoint: %w", err)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Checkpoint %q saved for path %q at offset %d\n", name, path, offset)
	return nil
}

func runCheckpointList(cmd *cobra.Command, _ []string) error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	all, err := audit.LoadCheckpoints(dir)
	if err != nil {
		return fmt.Errorf("load checkpoints: %w", err)
	}
	if len(all) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No checkpoints saved.")
		return nil
	}
	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tPATH\tOFFSET\tCREATED")
	for _, c := range all {
		fmt.Fprintf(w, "%s\t%s\t%d\t%s\n", c.Name, c.Path, c.Offset, c.CreatedAt.Format("2006-01-02 15:04:05"))
	}
	return w.Flush()
}
