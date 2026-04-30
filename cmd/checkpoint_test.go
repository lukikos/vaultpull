package cmd_test

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/yourusername/vaultpull/internal/audit"
)

func newCheckpointCmd(dir string) *cobra.Command {
	root := &cobra.Command{Use: "vaultpull"}

	checkpointCmd := &cobra.Command{Use: "checkpoint"}

	saveCmd := &cobra.Command{
		Use:  "save",
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			var offset int
			if _, err := fmt.Sscanf(args[2], "%d", &offset); err != nil {
				return err
			}
			return audit.SaveCheckpoint(dir, args[0], args[1], offset)
		},
	}

	listCmd := &cobra.Command{
		Use: "list",
		RunE: func(cmd *cobra.Command, _ []string) error {
			all, err := audit.LoadCheckpoints(dir)
			if err != nil {
				return err
			}
			if len(all) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No checkpoints saved.")
				return nil
			}
			for _, c := range all {
				fmt.Fprintf(cmd.OutOrStdout(), "%s %s %d\n", c.Name, c.Path, c.Offset)
			}
			return nil
		},
	}

	checkpointCmd.AddCommand(saveCmd, listCmd)
	root.AddCommand(checkpointCmd)
	return root
}

func TestRunCheckpointSave_CreatesEntry(t *testing.T) {
	dir := t.TempDir()
	if err := os.Chdir(dir); err != nil {
		t.Skip("cannot chdir:", err)
	}

	root := newCheckpointCmd(dir)
	root.SetArgs([]string{"checkpoint", "save", "v1", "secret/app", "7"})
	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	all, _ := audit.LoadCheckpoints(dir)
	if len(all) != 1 {
		t.Fatalf("expected 1 checkpoint, got %d", len(all))
	}
	if all[0].Offset != 7 {
		t.Errorf("expected offset 7, got %d", all[0].Offset)
	}
}

func TestRunCheckpointList_NoEntries(t *testing.T) {
	dir := t.TempDir()
	var buf bytes.Buffer
	root := newCheckpointCmd(dir)
	root.SetOut(&buf)
	root.SetArgs([]string{"checkpoint", "list"})
	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "No checkpoints") {
		t.Errorf("expected 'No checkpoints' message, got: %q", buf.String())
	}
}

func TestRunCheckpointList_ShowsEntries(t *testing.T) {
	dir := t.TempDir()
	_ = audit.SaveCheckpoint(dir, "rel", "secret/prod", 10)

	var buf bytes.Buffer
	root := newCheckpointCmd(dir)
	root.SetOut(&buf)
	root.SetArgs([]string{"checkpoint", "list"})
	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "rel") || !strings.Contains(out, "secret/prod") {
		t.Errorf("expected checkpoint details in output, got: %q", out)
	}
}
