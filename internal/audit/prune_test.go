package audit

import (
	"encoding/json"
	"os"
	"testing"
	"time"
)

func writeEntries(t *testing.T, path string, entries []Entry) {
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

func TestPrune_NoFile(t *testing.T) {
	removed, err := Prune("/tmp/nonexistent_prune_audit.log", PruneOptions{KeepTopN: 5})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if removed != 0 {
		t.Errorf("expected 0 removed, got %d", removed)
	}
}

func TestPrune_KeepsTopN(t *testing.T) {
	tmp := t.TempDir()
	logPath := tmp + "/audit.log"

	now := time.Now()
	var entries []Entry
	for i := 0; i < 5; i++ {
		entries = append(entries, Entry{
			Timestamp:  now.Add(time.Duration(i) * time.Minute),
			SecretPath: "secret/app",
			Action:     "sync",
		})
	}
	writeEntries(t, logPath, entries)

	removed, err := Prune(logPath, PruneOptions{KeepTopN: 3})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if removed != 2 {
		t.Errorf("expected 2 removed, got %d", removed)
	}

	remaining, _ := ReadAll(logPath)
	if len(remaining) != 3 {
		t.Errorf("expected 3 remaining, got %d", len(remaining))
	}
}

func TestPrune_MultiplePathsIndependent(t *testing.T) {
	tmp := t.TempDir()
	logPath := tmp + "/audit.log"

	now := time.Now()
	var entries []Entry
	for i := 0; i < 4; i++ {
		entries = append(entries, Entry{Timestamp: now.Add(time.Duration(i) * time.Minute), SecretPath: "secret/a", Action: "sync"})
		entries = append(entries, Entry{Timestamp: now.Add(time.Duration(i) * time.Minute), SecretPath: "secret/b", Action: "sync"})
	}
	writeEntries(t, logPath, entries)

	removed, err := Prune(logPath, PruneOptions{KeepTopN: 2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if removed != 4 {
		t.Errorf("expected 4 removed, got %d", removed)
	}

	remaining, _ := ReadAll(logPath)
	if len(remaining) != 4 {
		t.Errorf("expected 4 remaining, got %d", len(remaining))
	}
}

func TestPrune_KeepsNewestEntries(t *testing.T) {
	tmp := t.TempDir()
	logPath := tmp + "/audit.log"

	now := time.Now()
	entries := []Entry{
		{Timestamp: now.Add(-3 * time.Hour), SecretPath: "secret/x", Action: "sync"},
		{Timestamp: now.Add(-2 * time.Hour), SecretPath: "secret/x", Action: "sync"},
		{Timestamp: now.Add(-1 * time.Hour), SecretPath: "secret/x", Action: "sync"},
	}
	writeEntries(t, logPath, entries)

	_, err := Prune(logPath, PruneOptions{KeepTopN: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	remaining, _ := ReadAll(logPath)
	if len(remaining) != 1 {
		t.Fatalf("expected 1, got %d", len(remaining))
	}
	if remaining[0].Timestamp.Before(now.Add(-90 * time.Minute)) {
		t.Error("expected newest entry to be kept")
	}
}
