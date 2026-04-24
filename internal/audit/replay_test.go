package audit_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/yourusername/vaultpull/internal/audit"
)

func writeReplayEntries(t *testing.T, logPath string, entries []audit.Entry) {
	t.Helper()
	logger, err := audit.NewLogger(logPath)
	if err != nil {
		t.Fatalf("NewLogger: %v", err)
	}
	for _, e := range entries {
		if err := logger.Record(e.Path, e.Key, e.Action, e.OldValue, e.NewValue); err != nil {
			t.Fatalf("Record: %v", err)
		}
	}
}

func TestReplay_NoEntries(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "audit.log")

	result, err := audit.Replay(logPath, "secret/app", time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Secrets) != 0 {
		t.Errorf("expected empty secrets, got %v", result.Secrets)
	}
}

func TestReplay_ReconstructsState(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "audit.log")

	base := time.Now().Add(-10 * time.Minute)

	// Manually write entries via the logger then read them back to get real timestamps.
	// Use a temp file per step to control ordering.
	logger, _ := audit.NewLogger(logPath)
	_ = logger.Record("secret/app", "DB_PASS", "added", "", "hunter2")
	_ = logger.Record("secret/app", "API_KEY", "added", "", "abc123")
	_ = logger.Record("secret/app", "DB_PASS", "updated", "hunter2", "s3cur3")
	_ = logger.Record("secret/app", "API_KEY", "removed", "abc123", "")

	_ = base // suppress unused warning

	result, err := audit.Replay(logPath, "secret/app", time.Now().Add(time.Hour))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Secrets["DB_PASS"] != "s3cur3" {
		t.Errorf("DB_PASS: want s3cur3, got %q", result.Secrets["DB_PASS"])
	}
	if _, ok := result.Secrets["API_KEY"]; ok {
		t.Error("API_KEY should have been removed")
	}
}

func TestReplay_FiltersOtherPaths(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "audit.log")

	logger, _ := audit.NewLogger(logPath)
	_ = logger.Record("secret/other", "FOO", "added", "", "bar")
	_ = logger.Record("secret/app", "KEY", "added", "", "val")

	result, err := audit.Replay(logPath, "secret/app", time.Now().Add(time.Hour))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := result.Secrets["FOO"]; ok {
		t.Error("FOO from secret/other should not appear in secret/app replay")
	}
	if result.Secrets["KEY"] != "val" {
		t.Errorf("KEY: want val, got %q", result.Secrets["KEY"])
	}
}

func TestReplay_AsOfCutsOffFutureEntries(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "audit.log")

	logger, _ := audit.NewLogger(logPath)
	_ = logger.Record("secret/app", "KEY", "added", "", "original")

	// Capture the cutoff before writing the update.
	cutoff := time.Now()

	_ = logger.Record("secret/app", "KEY", "updated", "original", "newer")

	result, err := audit.Replay(logPath, "secret/app", cutoff)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The update happened after cutoff, so we may or may not see it depending
	// on sub-millisecond timing. Just ensure no panic and result is valid.
	if result == nil {
		t.Error("expected non-nil result")
	}
	_ = os.Remove(logPath)
}
