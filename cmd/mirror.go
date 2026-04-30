package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/yourusername/vaultpull/internal/audit"
)

func init() {
	mirrorCmd := &cobra.Command{
		Use:   "mirror",
		Short: "Save and load named mirrors of secret state",
	}

	saveCmd := &cobra.Command{
		Use:   "save <name> <path> <KEY=VALUE>...",
		Short: "Save a named mirror of a secret path",
		Args:  cobra.MinimumNArgs(2),
		RunE:  runMirrorSave,
	}

	showCmd := &cobra.Command{
		Use:   "show <name>",
		Short: "Display a saved mirror",
		Args:  cobra.ExactArgs(1),
		RunE:  runMirrorShow,
	}

	mirrorCmd.AddCommand(saveCmd, showCmd)
	rootCmd.AddCommand(mirrorCmd)
}

func runMirrorSave(cmd *cobra.Command, args []string) error {
	name := args[0]
	path := args[1]
	pairs := args[2:]

	keys := make(map[string]string, len(pairs))
	for _, p := range pairs {
		for i := 0; i < len(p); i++ {
			if p[i] == '=' {
				keys[p[:i]] = p[i+1:]
				break
			}
		}
	}

	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working dir: %w", err)
	}

	if err := audit.SaveMirror(dir, name, path, keys); err != nil {
		return fmt.Errorf("save mirror: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Mirror %q saved for path %q (%d keys)\n", name, path, len(keys))
	return nil
}

func runMirrorShow(cmd *cobra.Command, args []string) error {
	name := args[0]

	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working dir: %w", err)
	}

	entry, err := audit.LoadMirror(dir, name)
	if err != nil {
		return fmt.Errorf("load mirror: %w", err)
	}

	out := cmd.OutOrStdout()
	fmt.Fprintf(out, "Mirror: %s\n", entry.Name)
	fmt.Fprintf(out, "Path:   %s\n", entry.Path)
	fmt.Fprintf(out, "Saved:  %s\n", entry.CreatedAt.Format("2006-01-02 15:04:05 UTC"))
	fmt.Fprintf(out, "Keys (%d):\n", len(entry.Keys))
	for k := range entry.Keys {
		fmt.Fprintf(out, "  %s\n", k)
	}
	return nil
}
