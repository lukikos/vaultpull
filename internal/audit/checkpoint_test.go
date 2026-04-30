package audit_test

import (
	"testing"
	"time"

	"github.com/yourusername/vaultpull/internal/audit"
)

func TestSaveCheckpoint_EmptyName(t *testing.T) {
	dir := t.TempDir()
	err := audit.SaveCheckpoint(dir, "", "secret/app", 5)
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestSaveCheckpoint_EmptyPath(t *testing.T) {
	dir := t.TempDir()
	err := audit.SaveCheckpoint(dir, "v1", "", 5)
	if err == nil {
		t.Fatal("expected error for empty path")
	}
}

func TestSaveCheckpoint_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	if err := audit.SaveCheckpoint(dir, "v1", "secret/app", 3); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	all, err := audit.LoadCheckpoints(dir)
	if err != nil {
		t.Fatalf("load error: %v", err)
	}
	if len(all) != 1 {
		t.Fatalf("expected 1 checkpoint, got %d", len(all))
	}
	if all[0].Name != "v1" || all[0].Path != "secret/app" || all[0].Offset != 3 {
		t.Errorf("unexpected checkpoint: %+v", all[0])
	}
}

func TestSaveCheckpoint_OverwritesExisting(t *testing.T) {
	dir := t.TempDir()
	_ = audit.SaveCheckpoint(dir, "v1", "secret/app", 2)
	_ = audit.SaveCheckpoint(dir, "v1", "secret/app", 9)

	all, _ := audit.LoadCheckpoints(dir)
	if len(all) != 1 {
		t.Fatalf("expected 1 checkpoint after overwrite, got %d", len(all))
	}
	if all[0].Offset != 9 {
		t.Errorf("expected offset 9, got %d", all[0].Offset)
	}
}

func TestSaveCheckpoint_MultipleNamesKeptSeparate(t *testing.T) {
	dir := t.TempDir()
	_ = audit.SaveCheckpoint(dir, "v1", "secret/app", 1)
	_ = audit.SaveCheckpoint(dir, "v2", "secret/app", 5)

	all, _ := audit.LoadCheckpoints(dir)
	if len(all) != 2 {
		t.Fatalf("expected 2 checkpoints, got %d", len(all))
	}
}

func TestLoadCheckpoints_NoFile(t *testing.T) {
	dir := t.TempDir()
	all, err := audit.LoadCheckpoints(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(all) != 0 {
		t.Errorf("expected empty slice, got %d entries", len(all))
	}
}

func TestFindCheckpoint_NotFound(t *testing.T) {
	dir := t.TempDir()
	c, err := audit.FindCheckpoint(dir, "missing", "secret/app")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c != nil {
		t.Errorf("expected nil, got %+v", c)
	}
}

func TestFindCheckpoint_ReturnsMatch(t *testing.T) {
	dir := t.TempDir()
	_ = audit.SaveCheckpoint(dir, "release", "secret/prod", 42)
	c, err := audit.FindCheckpoint(dir, "release", "secret/prod")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c == nil {
		t.Fatal("expected checkpoint, got nil")
	}
	if c.Offset != 42 {
		t.Errorf("expected offset 42, got %d", c.Offset)
	}
	if c.CreatedAt.IsZero() {
		t.Error("expected non-zero CreatedAt")
	}
	if time.Since(c.CreatedAt) > 5*time.Second {
		t.Error("CreatedAt seems stale")
	}
}
