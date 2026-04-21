package audit_test

import (
	"testing"
	"time"
)

func TestPrune_AfterWritePreservesNewest(t *testing.T) {
	tmp := t.TempDir()
	logPath := tmp + "/audit.log"

	logger, err := NewLogger(logPath)
	if err != nil {
		t.Fatalf("NewLogger: %v", err)
	}

	now := time.Now()
	paths := []string{"secret/alpha", "secret/beta"}
	for _, p := range paths {
		for i := 0; i < 5; i++ {
			if err := logger.Record(Entry{
				Timestamp:  now.Add(time.Duration(i) * time.Minute),
				SecretPath: p,
				Action:     "sync",
			}); err != nil {
				t.Fatalf("Record: %v", err)
			}
		}
	}

	all, err := ReadAll(logPath)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if len(all) != 10 {
		t.Fatalf("expected 10 entries before prune, got %d", len(all))
	}

	removed, err := Prune(logPath, PruneOptions{KeepTopN: 2})
	if err != nil {
		t.Fatalf("Prune: %v", err)
	}
	if removed != 6 {
		t.Errorf("expected 6 removed, got %d", removed)
	}

	remaining, err := ReadAll(logPath)
	if err != nil {
		t.Fatalf("ReadAll after prune: %v", err)
	}
	if len(remaining) != 4 {
		t.Errorf("expected 4 remaining, got %d", len(remaining))
	}

	// Verify the newest entries per path were retained.
	counts := make(map[string]int)
	for _, e := range remaining {
		counts[e.SecretPath]++
	}
	for _, p := range paths {
		if counts[p] != 2 {
			t.Errorf("expected 2 entries for %s, got %d", p, counts[p])
		}
	}
}
