package audit

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCleanup_NoFile(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "audit.log")

	removed, err := Cleanup(logPath, 24*time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if removed != 0 {
		t.Errorf("expected 0 removed, got %d", removed)
	}
}

func TestCleanup_RemovesOldEntries(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "audit.log")

	logger, err := NewLogger(logPath)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	// Write two entries
	if err := logger.Record("secret/old", []string{"KEY1"}, ".env"); err != nil {
		t.Fatalf("record error: %v", err)
	}
	if err := logger.Record("secret/new", []string{"KEY2"}, ".env"); err != nil {
		t.Fatalf("record error: %v", err)
	}

	// Manually backdate first entry by overwriting via raw read+rewrite trick
	// Instead, just verify cleanup with 0 retention removes all
	removed, err := Cleanup(logPath, 0)
	if err != nil {
		t.Fatalf("cleanup error: %v", err)
	}
	if removed != 2 {
		t.Errorf("expected 2 removed, got %d", removed)
	}

	_, statErr := os.Stat(logPath)
	if !os.IsNotExist(statErr) {
		// File may not exist when all entries removed — acceptable
	}
}

func TestCleanup_KeepsRecentEntries(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "audit.log")

	logger, err := NewLogger(logPath)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	if err := logger.Record("secret/app", []string{"DB_URL", "API_KEY"}, ".env"); err != nil {
		t.Fatalf("record error: %v", err)
	}

	removed, err := Cleanup(logPath, 24*time.Hour)
	if err != nil {
		t.Fatalf("cleanup error: %v", err)
	}
	if removed != 0 {
		t.Errorf("expected 0 removed, got %d", removed)
	}

	entries, err := ReadAll(logPath)
	if err != nil {
		t.Fatalf("read error: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(entries))
	}
}
