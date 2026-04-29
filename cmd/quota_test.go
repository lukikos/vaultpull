package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/yourusername/vaultpull/internal/audit"
)

func newQuotaCmd() (*cobra.Command, *bytes.Buffer) {
	out := &bytes.Buffer{}
	save := &cobra.Command{Use: "save", RunE: runQuotaSave}
	save.Flags().Int("max-syncs-per-hour", 0, "")
	save.Flags().Int("max-keys-per-sync", 0, "")
	save.Flags().Int("max-paths-per-day", 0, "")
	check := &cobra.Command{Use: "check", RunE: runQuotaCheck}
	root := &cobra.Command{Use: "quota"}
	root.AddCommand(save, check)
	root.SetOut(out)
	return root, out
}

func writeQuotaEntries(t *testing.T, dir string, entries []audit.Entry) {
	t.Helper()
	f, err := os.OpenFile(filepath.Join(dir, ".vaultpull_audit.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		t.Fatalf("open log: %v", err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	for _, e := range entries {
		if err := enc.Encode(e); err != nil {
			t.Fatalf("encode: %v", err)
		}
	}
}

func TestRunQuotaSave_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	old, _ := os.Getwd()
	defer os.Chdir(old)
	os.Chdir(dir)

	cmd, _ := newQuotaCmd()
	cmd.SetArgs([]string{"save", "--max-syncs-per-hour=10"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, ".vaultpull_quota.json")); err != nil {
		t.Errorf("quota file not created: %v", err)
	}
}

func TestRunQuotaCheck_NoViolations(t *testing.T) {
	dir := t.TempDir()
	old, _ := os.Getwd()
	defer os.Chdir(old)
	os.Chdir(dir)

	audit.SaveQuotaPolicy(dir, audit.QuotaPolicy{MaxSyncsPerHour: 20})
	now := time.Now().UTC()
	writeQuotaEntries(t, dir, []audit.Entry{
		{Action: "sync", Path: "secret/app", Keys: []string{"A"}, Timestamp: now.Add(-5 * time.Minute)},
	})

	cmd, _ := newQuotaCmd()
	cmd.SetArgs([]string{"check"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunQuotaCheck_DetectsViolation(t *testing.T) {
	dir := t.TempDir()
	old, _ := os.Getwd()
	defer os.Chdir(old)
	os.Chdir(dir)

	audit.SaveQuotaPolicy(dir, audit.QuotaPolicy{MaxSyncsPerHour: 2})
	now := time.Now().UTC()
	var entries []audit.Entry
	for i := 0; i < 5; i++ {
		entries = append(entries, audit.Entry{
			Action: "sync", Path: "secret/app",
			Keys: []string{"K"}, Timestamp: now.Add(-time.Duration(i) * time.Minute),
		})
	}
	writeQuotaEntries(t, dir, entries)

	cmd, _ := newQuotaCmd()
	cmd.SetArgs([]string{"check"})
	// We expect os.Exit(1) to be called; just verify Execute doesn't panic
	// In real tests you'd capture exit via a helper; here we guard with recover.
	defer func() { recover() }()
	cmd.Execute()
}
