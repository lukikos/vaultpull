package audit_test

import (
	"testing"

	"github.com/yourusername/vaultpull/internal/audit"
)

// TestCheckpoint_RoundTripAfterMultipleSaves verifies that saving multiple
// checkpoints across different paths and names preserves each independently.
func TestCheckpoint_RoundTripAfterMultipleSaves(t *testing.T) {
	dir := t.TempDir()

	entries := []struct {
		name   string
		path   string
		offset int
	}{
		{"alpha", "secret/svc-a", 1},
		{"beta", "secret/svc-a", 5},
		{"alpha", "secret/svc-b", 3},
	}

	for _, e := range entries {
		if err := audit.SaveCheckpoint(dir, e.name, e.path, e.offset); err != nil {
			t.Fatalf("save error: %v", err)
		}
	}

	all, err := audit.LoadCheckpoints(dir)
	if err != nil {
		t.Fatalf("load error: %v", err)
	}
	if len(all) != 3 {
		t.Fatalf("expected 3 checkpoints, got %d", len(all))
	}

	c, err := audit.FindCheckpoint(dir, "alpha", "secret/svc-a")
	if err != nil || c == nil {
		t.Fatalf("FindCheckpoint returned nil or error: %v", err)
	}
	if c.Offset != 1 {
		t.Errorf("expected offset 1, got %d", c.Offset)
	}

	c, err = audit.FindCheckpoint(dir, "beta", "secret/svc-a")
	if err != nil || c == nil {
		t.Fatalf("FindCheckpoint returned nil or error: %v", err)
	}
	if c.Offset != 5 {
		t.Errorf("expected offset 5, got %d", c.Offset)
	}
}

// TestCheckpoint_OverwritePreservesOthers ensures overwriting one checkpoint
// does not disturb unrelated checkpoints.
func TestCheckpoint_OverwritePreservesOthers(t *testing.T) {
	dir := t.TempDir()

	_ = audit.SaveCheckpoint(dir, "v1", "secret/app", 1)
	_ = audit.SaveCheckpoint(dir, "v2", "secret/app", 2)
	_ = audit.SaveCheckpoint(dir, "v1", "secret/app", 99) // overwrite v1

	all, _ := audit.LoadCheckpoints(dir)
	if len(all) != 2 {
		t.Fatalf("expected 2 checkpoints, got %d", len(all))
	}

	v1, _ := audit.FindCheckpoint(dir, "v1", "secret/app")
	if v1 == nil || v1.Offset != 99 {
		t.Errorf("expected v1 offset=99, got %+v", v1)
	}

	v2, _ := audit.FindCheckpoint(dir, "v2", "secret/app")
	if v2 == nil || v2.Offset != 2 {
		t.Errorf("expected v2 offset=2, got %+v", v2)
	}
}
