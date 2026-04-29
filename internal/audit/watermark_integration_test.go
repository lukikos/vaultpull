package audit_test

import (
	"testing"
	"time"

	"github.com/yourusername/vaultpull/internal/audit"
)

// TestWatermark_AfterWriteReflectsLatest verifies that after writing multiple
// sync entries via the logger, ComputeWatermarks returns the most recent one.
func TestWatermark_AfterWriteReflectsLatest(t *testing.T) {
	dir := t.TempDir()
	logger, err := audit.NewLogger(dir)
	if err != nil {
		t.Fatalf("NewLogger: %v", err)
	}

	base := time.Now().UTC().Truncate(time.Second)
	for i := 0; i < 3; i++ {
		if err := logger.Record(audit.Entry{
			Path:      "secret/integration",
			Action:    "sync",
			Timestamp: base.Add(time.Duration(i) * time.Minute),
			Keys:      []string{"KEY"},
		}); err != nil {
			t.Fatalf("Record: %v", err)
		}
	}

	entries, err := audit.ReadAll(logger.Path())
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	wm := audit.ComputeWatermarks(entries)
	if len(wm) != 1 {
		t.Fatalf("expected 1 watermark, got %d", len(wm))
	}
	expected := base.Add(2 * time.Minute)
	if !wm[0].LatestAt.Equal(expected) {
		t.Errorf("want latest %v, got %v", expected, wm[0].LatestAt)
	}
	if wm[0].SyncCount != 3 {
		t.Errorf("want sync count 3, got %d", wm[0].SyncCount)
	}
}

// TestWatermark_SaveAndLoadAfterWrite ensures the full pipeline of write →
// compute → save → load preserves data integrity.
func TestWatermark_SaveAndLoadAfterWrite(t *testing.T) {
	dir := t.TempDir()
	logger, err := audit.NewLogger(dir)
	if err != nil {
		t.Fatalf("NewLogger: %v", err)
	}

	now := time.Now().UTC().Truncate(time.Second)
	paths := []string{"secret/alpha", "secret/beta"}
	for _, p := range paths {
		if err := logger.Record(audit.Entry{
			Path:      p,
			Action:    "sync",
			Timestamp: now,
			Keys:      []string{"X"},
		}); err != nil {
			t.Fatalf("Record: %v", err)
		}
	}

	entries, _ := audit.ReadAll(logger.Path())
	wm := audit.ComputeWatermarks(entries)
	if err := audit.SaveWatermarks(dir, wm); err != nil {
		t.Fatalf("SaveWatermarks: %v", err)
	}
	loaded, err := audit.LoadWatermarks(dir)
	if err != nil {
		t.Fatalf("LoadWatermarks: %v", err)
	}
	if len(loaded) != 2 {
		t.Fatalf("expected 2 watermarks, got %d", len(loaded))
	}
}
