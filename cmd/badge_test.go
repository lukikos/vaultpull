package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"

	"github.com/yourusername/vaultpull/internal/audit"
)

func newBadgeCmd(t *testing.T, dir string) (*cobra.Command, *bytes.Buffer) {
	t.Helper()
	out := &bytes.Buffer{}
	cmd := &cobra.Command{
		Use: "badge",
		RunE: func(cmd *cobra.Command, args []string) error {
			entries, err := audit.ReadAll(dir)
			if err != nil {
				return err
			}
			badges := audit.GenerateBadges(entries, audit.DefaultBadgeConfig())
			if len(badges) == 0 {
				cmd.Println("No audit entries found.")
				return nil
			}
			for _, b := range badges {
				cmd.Printf("%s %s %s\n", b.Path, b.Status, b.Message)
			}
			return nil
		},
	}
	cmd.SetOut(out)
	return cmd, out
}

func writeBadgeEntries(t *testing.T, dir string, entries []audit.Entry) {
	t.Helper()
	logger, err := audit.NewLogger(dir)
	if err != nil {
		t.Fatalf("NewLogger: %v", err)
	}
	for _, e := range entries {
		if err := logger.Record(e.Path, e.Action, e.Keys); err != nil {
			t.Fatalf("Record: %v", err)
		}
	}
}

func TestRunBadge_NoEntries(t *testing.T) {
	dir := t.TempDir()
	cmd, out := newBadgeCmd(t, dir)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !strings.Contains(out.String(), "No audit entries found") {
		t.Errorf("expected no-entries message, got: %s", out.String())
	}
}

func TestRunBadge_ShowsPath(t *testing.T) {
	dir := t.TempDir()
	logger, _ := audit.NewLogger(dir)
	_ = logger.Record("secret/app", "sync", []string{"KEY"})

	cmd, out := newBadgeCmd(t, dir)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !strings.Contains(out.String(), "secret/app") {
		t.Errorf("expected path in output, got: %s", out.String())
	}
}

func TestRunBadge_JSONOutput(t *testing.T) {
	dir := t.TempDir()
	logger, _ := audit.NewLogger(dir)
	_ = logger.Record("secret/app", "sync", []string{"KEY"})

	entries, _ := audit.ReadAll(dir)
	badges := audit.GenerateBadges(entries, audit.DefaultBadgeConfig())

	data, err := json.Marshal(badges)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(data), "secret/app") {
		t.Errorf("expected path in JSON, got: %s", string(data))
	}
}

func TestRunBadge_StaleBadgeWritten(t *testing.T) {
	dir := t.TempDir()
	// Write a raw entry with an old timestamp directly
	old := time.Now().Add(-72 * time.Hour).Format(time.RFC3339)
	line := `{"timestamp":"` + old + `","path":"secret/old","action":"sync","keys":["X"]}` + "\n"
	_ = os.WriteFile(filepath.Join(dir, "audit.log"), []byte(line), 0600)

	entries, _ := audit.ReadAll(dir)
	badges := audit.GenerateBadges(entries, audit.DefaultBadgeConfig())
	if len(badges) != 1 {
		t.Fatalf("expected 1 badge, got %d", len(badges))
	}
	if badges[0].Status != audit.BadgeStatusWarning {
		t.Errorf("expected warning for stale path, got %s", badges[0].Status)
	}
}
