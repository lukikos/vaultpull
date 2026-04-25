package audit_test

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLint_AfterWriteIsClean(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "audit.log")

	logger, err := NewLogger(logPath)
	if err != nil {
		t.Fatalf("NewLogger: %v", err)
	}

	keys := []string{"DB_HOST", "DB_PORT", "API_SECRET", "JWT_TOKEN"}
	for _, k := range keys {
		if err := logger.Record("secret/myapp", k, "added"); err != nil {
			t.Fatalf("Record %s: %v", k, err)
		}
	}

	report, err := Lint(logPath)
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	if !report.Clean {
		for _, w := range report.Warnings {
			t.Logf("warning [%d] %s/%s: %s", w.Index, w.Path, w.Key, w.Warning)
		}
		t.Errorf("expected clean log after normal writes, got %d warning(s)", report.Total)
	}
}

func TestLint_MultipleIssuesReported(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "audit.log")

	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	lines := []string{
		`{"timestamp":"2024-01-01T00:00:00Z","path":"","key":"GOOD_KEY","action":"added"}`,
		`{"timestamp":"2024-01-01T00:00:00Z","path":"secret/app","key":"","action":"updated"}`,
		`{"timestamp":"2024-01-01T00:00:00Z","path":"secret/app","key":"BAD KEY","action":"synced"}`,
		`{"timestamp":"2024-01-01T00:00:00Z","path":"secret/app","key":"VALID","action":"removed"}`,
	}
	for _, l := range lines {
		f.WriteString(l + "\n")
	}
	f.Close()

	report, err := Lint(logPath)
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	if report.Total < 3 {
		t.Errorf("expected at least 3 warnings, got %d", report.Total)
	}
}
