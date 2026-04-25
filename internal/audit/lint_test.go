package audit

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func writeLintEntries(t *testing.T, dir string, entries []Entry) string {
	t.Helper()
	path := filepath.Join(dir, "audit.log")
	logger, err := NewLogger(path)
	if err != nil {
		t.Fatalf("NewLogger: %v", err)
	}
	for _, e := range entries {
		if err := logger.Record(e.Path, e.Key, e.Action); err != nil {
			t.Fatalf("Record: %v", err)
		}
	}
	return path
}

func TestLint_NoFile(t *testing.T) {
	report, err := Lint(filepath.Join(t.TempDir(), "missing.log"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !report.Clean {
		t.Errorf("expected clean report for missing log")
	}
}

func TestLint_CleanEntries(t *testing.T) {
	dir := t.TempDir()
	path := writeLintEntries(t, dir, []Entry{
		{Path: "secret/app", Key: "DB_PASS", Action: "added", Timestamp: time.Now()},
		{Path: "secret/app", Key: "API_KEY", Action: "updated", Timestamp: time.Now()},
	})

	report, err := Lint(path)
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	if !report.Clean {
		t.Errorf("expected clean, got %d warning(s)", report.Total)
	}
}

func TestLint_DetectsEmptyKey(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "audit.log")

	// Write a raw entry with an empty key directly
	f, _ := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY, 0600)
	f.WriteString(`{"timestamp":"2024-01-01T00:00:00Z","path":"secret/app","key":"","action":"added"}` + "\n")
	f.Close()

	report, err := Lint(logPath)
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	if report.Clean {
		t.Error("expected warning for empty key")
	}
	if report.Warnings[0].Warning != "empty key" {
		t.Errorf("unexpected warning: %s", report.Warnings[0].Warning)
	}
}

func TestLint_DetectsUnknownAction(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "audit.log")

	f, _ := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY, 0600)
	f.WriteString(`{"timestamp":"2024-01-01T00:00:00Z","path":"secret/app","key":"TOKEN","action":"deleted"}` + "\n")
	f.Close()

	report, err := Lint(logPath)
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	if report.Clean {
		t.Error("expected warning for unknown action")
	}
}

func TestLint_DetectsKeyWithWhitespace(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "audit.log")

	f, _ := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY, 0600)
	f.WriteString(`{"timestamp":"2024-01-01T00:00:00Z","path":"secret/app","key":"MY KEY","action":"added"}` + "\n")
	f.Close()

	report, err := Lint(logPath)
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	if report.Total == 0 {
		t.Error("expected warning for key with whitespace")
	}
}
