package audit_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/yourusername/vaultpull/internal/audit"
)

func writeRewindEntries(t *testing.T, logPath string, entries []audit.Entry) {
	t.Helper()
	logger, err := audit.NewLogger(logPath)
	if err != nil {
		t.Fatalf("NewLogger: %v", err)
	}
	for _, e := range entries {
		if err := logger.Record(e.Action, e.Path, e.Keys); err != nil {
			t.Fatalf("Record: %v", err)
		}
	}
}

func TestRewind_EmptyPath(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "audit.log")
	_, err := audit.Rewind(logPath, "", time.Now())
	if err == nil {
		t.Fatal("expected error for empty path")
	}
}

func TestRewind_ZeroTime(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "audit.log")
	_, err := audit.Rewind(logPath, "secret/app", time.Time{})
	if err == nil {
		t.Fatal("expected error for zero time")
	}
}

func TestRewind_NoEntries(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "audit.log")
	// no log file written
	result, err := audit.Rewind(logPath, "secret/app", time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.State) != 0 {
		t.Errorf("expected empty state, got %v", result.State)
	}
}

func TestRewind_ReconstructsState(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "audit.log")

	now := time.Now()
	logger, err := audit.NewLogger(logPath)
	if err != nil {
		t.Fatalf("NewLogger: %v", err)
	}
	if err := logger.Record("sync", "secret/app", map[string]string{"DB_PASS": "abc", "API_KEY": "xyz"}); err != nil {
		t.Fatalf("Record: %v", err)
	}

	result, err := audit.Rewind(logPath, "secret/app", now.Add(time.Minute))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.State["DB_PASS"] != "abc" {
		t.Errorf("expected DB_PASS=abc, got %q", result.State["DB_PASS"])
	}
	if result.State["API_KEY"] != "xyz" {
		t.Errorf("expected API_KEY=xyz, got %q", result.State["API_KEY"])
	}
}

func TestRewind_CutoffExcludesFutureEntries(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "audit.log")

	past := time.Now().Add(-2 * time.Hour)

	logger, err := audit.NewLogger(logPath)
	if err != nil {
		t.Fatalf("NewLogger: %v", err)
	}
	if err := logger.Record("sync", "secret/app", map[string]string{"TOKEN": "newval"}); err != nil {
		t.Fatalf("Record: %v", err)
	}

	// Rewind to before the entry was written
	result, err := audit.Rewind(logPath, "secret/app", past)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := result.State["TOKEN"]; ok {
		t.Error("expected TOKEN to be absent for past rewind")
	}
}

func TestRewind_FiltersOtherPaths(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "audit.log")

	logger, err := audit.NewLogger(logPath)
	if err != nil {
		t.Fatalf("NewLogger: %v", err)
	}
	if err := logger.Record("sync", "secret/other", map[string]string{"FOO": "bar"}); err != nil {
		t.Fatalf("Record: %v", err)
	}

	result, err := audit.Rewind(logPath, "secret/app", time.Now().Add(time.Minute))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.State) != 0 {
		t.Errorf("expected empty state for unrelated path, got %v", result.State)
	}
	_ = os.Remove(logPath)
}
