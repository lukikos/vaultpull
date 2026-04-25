package audit_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/yourusername/vaultpull/internal/audit"
)

func writeVerifyEntries(t *testing.T, logPath string, entries []audit.Entry) {
	t.Helper()
	l, err := audit.NewLogger(logPath)
	if err != nil {
		t.Fatalf("NewLogger: %v", err)
	}
	for _, e := range entries {
		if err := l.Record(e.SecretPath, e.Key, e.Action); err != nil {
			t.Fatalf("Record: %v", err)
		}
	}
}

func TestVerify_MissingEnvFile(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "audit.log")

	writeVerifyEntries(t, logPath, []audit.Entry{
		{SecretPath: "secret/app", Key: "DB_HOST", Action: "added", Timestamp: time.Now()},
	})

	result, err := audit.Verify(logPath, "secret/app", filepath.Join(dir, "nonexistent.env"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.OK {
		t.Error("expected OK=false for missing env file")
	}
	if len(result.MissingKeys) == 0 {
		t.Error("expected missing keys to be populated")
	}
}

func TestVerify_AllKeysPresent(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "audit.log")
	envPath := filepath.Join(dir, ".env")

	writeVerifyEntries(t, logPath, []audit.Entry{
		{SecretPath: "secret/app", Key: "API_KEY", Action: "added", Timestamp: time.Now()},
		{SecretPath: "secret/app", Key: "DB_PASS", Action: "added", Timestamp: time.Now()},
	})

	_ = os.WriteFile(envPath, []byte("API_KEY=abc\nDB_PASS=xyz\n"), 0600)

	result, err := audit.Verify(logPath, "secret/app", envPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.OK {
		t.Errorf("expected OK=true, missing: %v", result.MissingKeys)
	}
	if len(result.MatchedKeys) != 2 {
		t.Errorf("expected 2 matched keys, got %d", len(result.MatchedKeys))
	}
}

func TestVerify_ExtraKeysDetected(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "audit.log")
	envPath := filepath.Join(dir, ".env")

	writeVerifyEntries(t, logPath, []audit.Entry{
		{SecretPath: "secret/app", Key: "API_KEY", Action: "added", Timestamp: time.Now()},
	})

	_ = os.WriteFile(envPath, []byte("API_KEY=abc\nEXTRA_KEY=oops\n"), 0600)

	result, err := audit.Verify(logPath, "secret/app", envPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.ExtraKeys) == 0 {
		t.Error("expected extra keys to be detected")
	}
}
