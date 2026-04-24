package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/yourusername/vaultpull/internal/audit"
)

var tagNote string

func init() {
	tagCmd := &cobra.Command{
		Use:   "tag <name>",
		Short: "Tag the current point in the audit log",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runTag,
	}
	tagCmd.Flags().StringVar(&tagNote, "note", "", "Optional note to attach to the tag")

	listTagsCmd := &cobra.Command{
		Use:   "tags",
		Short: "List all audit log tags",
		RunE:  runListTags,
	}

	rootCmd.AddCommand(tagCmd)
	rootCmd.AddCommand(listTagsCmd)
}

func runTag(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("tag name is required")
	}
	logPath, _ := cmd.Root().PersistentFlags().GetString("log")
	if logPath == "" {
		logPath = ".vaultpull.log"
	}
	name := args[0]
	if err := audit.SaveTag(logPath, name, tagNote); err != nil {
		return fmt.Errorf("save tag: %w", err)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Tag %q saved.\n", name)
	return nil
}

func runListTags(cmd *cobra.Command, args []string) error {
	logPath, _ := cmd.Root().PersistentFlags().GetString("log")
	if logPath == "" {
		logPath = ".vaultpull.log"
	}
	tags, err := audit.LoadTags(logPath)
	if err != nil {
		return fmt.Errorf("load tags: %w", err)
	}
	if len(tags) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No tags found.")
		return nil
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tCREATED\tNOTE")
	for _, tag := range tags {
		fmt.Fprintf(w, "%s\t%s\t%s\n", tag.Name, tag.CreatedAt.Format(time.RFC3339), tag.Note)
	}
	return w.Flush()
}
