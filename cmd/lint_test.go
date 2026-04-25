package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/yourorg/vaultpull/internal/audit"
)

func newLintCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "lint",
		RunE: runLint,
	}
	cmd.Flags().StringVar(&lintLogPath, "log", "vaultpull-audit.log", "")
	return cmd
}

func TestRunLint_CleanLog(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "audit.log")

	logger, err := audit.NewLogger(logPath)
	if err != nil {
		t.Fatalf("NewLogger: %v", err)
	}
	if err := logger.Record("secret/app", "DB_PASS", "added"); err != nil {
		t.Fatalf("Record: %v", err)
	}

	lintLogPath = logPath
	cmd := newLintCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if !strings.Contains(buf.String(), "clean") {
		t.Errorf("expected clean message, got: %s", buf.String())
	}
}

func TestRunLint_NoFile(t *testing.T) {
	lintLogPath = filepath.Join(t.TempDir(), "missing.log")
	cmd := newLintCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !strings.Contains(buf.String(), "clean") {
		t.Errorf("expected clean output for missing log, got: %s", buf.String())
	}
}

func TestRunLint_DetectsIssues(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "audit.log")

	f, _ := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY, 0600)
	f.WriteString(`{"timestamp":"2024-01-01T00:00:00Z","path":"secret/app","key":"","action":"added"}` + "\n")
	f.Close()

	lintLogPath = logPath

	// runLint calls os.Exit(1) on issues; test the report directly instead.
	report, err := audit.Lint(logPath)
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	if report.Clean {
		t.Error("expected issues to be detected")
	}
	if report.Total == 0 {
		t.Error("expected at least one warning")
	}
}
