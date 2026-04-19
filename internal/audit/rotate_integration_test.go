package audit_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/user/vaultpull/internal/audit"
)

func TestRotate_AfterWritePreservesEntries(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "audit.log")

	logger, err := audit.NewLogger(logPath)
	if err != nil {
		t.Fatal(err)
	}

	entry := audit.Entry{
		Timestamp:  time.Now().UTC(),
		SecretPath: "secret/data/app",
		Keys:       []string{"DB_PASS"},
		Status:     "ok",
	}
	if err := logger.Record(entry); err != nil {
		t.Fatal(err)
	}

	archived, err := audit.Rotate(logPath)
	if err != nil {
		t.Fatalf("rotate failed: %v", err)
	}

	if archived == "" {
		t.Fatal("expected archive path")
	}

	data, err := os.ReadFile(archived)
	if err != nil {
		t.Fatal(err)
	}

	var got audit.Entry
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("invalid JSON in archive: %v", err)
	}

	if got.SecretPath != entry.SecretPath {
		t.Errorf("path mismatch: got %q want %q", got.SecretPath, entry.SecretPath)
	}
}
