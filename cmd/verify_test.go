package cmd_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/yourusername/vaultpull/internal/audit"
)

func newVerifyCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "verify", RunE: runVerify}
	cmd.Flags().String("env-file", ".env", "")
	cmd.Flags().String("log-file", "vaultpull-audit.log", "")
	return cmd
}

func writeVerifyCmdEntries(t *testing.T, logPath string) {
	t.Helper()
	l, err := audit.NewLogger(logPath)
	if err != nil {
		t.Fatalf("NewLogger: %v", err)
	}
	e := audit.Entry{SecretPath: "secret/app", Key: "TOKEN", Action: "added", Timestamp: time.Now()}
	if err := l.Record(e.SecretPath, e.Key, e.Action); err != nil {
		t.Fatalf("Record: %v", err)
	}
}

func TestRunVerify_OK(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "audit.log")
	envPath := filepath.Join(dir, ".env")

	writeVerifyCmdEntries(t, logPath)
	_ = os.WriteFile(envPath, []byte("TOKEN=secret\n"), 0600)

	cmd := newVerifyCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	_ = cmd.Flags().Set("env-file", envPath)
	_ = cmd.Flags().Set("log-file", logPath)

	// Inject secret path via env for config.Load
	t.Setenv("VAULT_SECRET_PATH", "secret/app")
	t.Setenv("VAULT_TOKEN", "test-token")

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !contains(out, "OK") {
		t.Errorf("expected OK in output, got: %s", out)
	}
}

func TestRunVerify_DriftDetected(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "audit.log")
	envPath := filepath.Join(dir, ".env")

	writeVerifyCmdEntries(t, logPath)
	// Write .env without the expected key
	_ = os.WriteFile(envPath, []byte("UNRELATED=value\n"), 0600)

	cmd := newVerifyCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	_ = cmd.Flags().Set("env-file", envPath)
	_ = cmd.Flags().Set("log-file", logPath)

	t.Setenv("VAULT_SECRET_PATH", "secret/app")
	t.Setenv("VAULT_TOKEN", "test-token")

	// We expect a non-zero exit but the command itself should not error
	_ = cmd.Execute()
	out := buf.String()
	if !contains(out, "DRIFT") {
		t.Errorf("expected DRIFT in output, got: %s", out)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 &&
		(func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		})())
}
