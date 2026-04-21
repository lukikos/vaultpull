package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/yourorg/vaultpull/internal/audit"
)

func writeTestEntries(t *testing.T, path string, entries []audit.Entry) {
	t.Helper()
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	for _, e := range entries {
		if err := enc.Encode(e); err != nil {
			t.Fatalf("encode: %v", err)
		}
	}
}

func TestRunPrune_NothingToPrune(t *testing.T) {
	tmp := t.TempDir()
	logPath := tmp + "/audit.log"

	now := time.Now()
	writeTestEntries(t, logPath, []audit.Entry{
		{Timestamp: now, SecretPath: "secret/app", Action: "sync"},
	})

	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)

	pruneCmd.Flags().Set("keep", "5")
	err := runPrune(pruneCmd, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunPrune_RemovesEntries(t *testing.T) {
	tmp := t.TempDir()
	logPath := tmp + "/audit.log"

	now := time.Now()
	var entries []audit.Entry
	for i := 0; i < 6; i++ {
		entries = append(entries, audit.Entry{
			Timestamp:  now.Add(time.Duration(i) * time.Minute),
			SecretPath: "secret/app",
			Action:     "sync",
		})
	}
	writeTestEntries(t, logPath, entries)

	opts := audit.PruneOptions{KeepTopN: 3}
	removed, err := audit.Prune(logPath, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if removed != 3 {
		t.Errorf("expected 3 removed, got %d", removed)
	}

	remaining, _ := audit.ReadAll(logPath)
	if len(remaining) != 3 {
		t.Errorf("expected 3 remaining entries, got %d", len(remaining))
	}
}
