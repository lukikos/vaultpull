package audit_test

import (
	"testing"
	"time"

	"github.com/yourusername/vaultpull/internal/audit"
)

func TestRollback_EmptyTag(t *testing.T) {
	dir := t.TempDir()
	_, err := audit.Rollback(dir, "", "secret/app")
	if err == nil {
		t.Fatal("expected error for empty tag name")
	}
}

func TestRollback_EmptyPath(t *testing.T) {
	dir := t.TempDir()
	_, err := audit.Rollback(dir, "v1", "")
	if err == nil {
		t.Fatal("expected error for empty path")
	}
}

func TestRollback_TagNotFound(t *testing.T) {
	dir := t.TempDir()
	_, err := audit.Rollback(dir, "nonexistent", "secret/app")
	if err == nil {
		t.Fatal("expected error when tag does not exist")
	}
}

func TestRollback_RestoresState(t *testing.T) {
	dir := t.TempDir()

	// Save a snapshot under tag "v1".
	snap := audit.Snapshot{
		"secret/app": {"DB_HOST": "localhost", "DB_PORT": "5432"},
	}
	if err := audit.SaveSnapshot(dir, "v1", snap); err != nil {
		t.Fatalf("SaveSnapshot: %v", err)
	}

	// Save the corresponding tag.
	if err := audit.SaveTag(dir, "v1"); err != nil {
		t.Fatalf("SaveTag: %v", err)
	}

	result, err := audit.Rollback(dir, "v1", "secret/app")
	if err != nil {
		t.Fatalf("Rollback: %v", err)
	}

	if result.TagName != "v1" {
		t.Errorf("TagName = %q, want %q", result.TagName, "v1")
	}
	if result.Path != "secret/app" {
		t.Errorf("Path = %q, want %q", result.Path, "secret/app")
	}
	if result.Restored["DB_HOST"] != "localhost" {
		t.Errorf("DB_HOST = %q, want %q", result.Restored["DB_HOST"], "localhost")
	}
	if result.Restored["DB_PORT"] != "5432" {
		t.Errorf("DB_PORT = %q, want %q", result.Restored["DB_PORT"], "5432")
	}
	if result.Timestamp.IsZero() {
		t.Error("expected non-zero timestamp")
	}
}

func TestRollback_PathMissingInSnapshot(t *testing.T) {
	dir := t.TempDir()

	snap := audit.Snapshot{
		"secret/other": {"KEY": "val"},
	}
	if err := audit.SaveSnapshot(dir, "v2", snap); err != nil {
		t.Fatalf("SaveSnapshot: %v", err)
	}
	if err := audit.SaveTag(dir, "v2"); err != nil {
		t.Fatalf("SaveTag: %v", err)
	}

	_, err := audit.Rollback(dir, "v2", "secret/app")
	if err == nil {
		t.Fatal("expected error when path not in snapshot")
	}
	_ = time.Now() // ensure time import used
}
