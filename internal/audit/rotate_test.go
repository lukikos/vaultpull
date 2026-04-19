package audit

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRotate_NoFile(t *testing.T) {
	dir := t.TempDir()
	archived, err := Rotate(filepath.Join(dir, "audit.log"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if archived != "" {
		t.Errorf("expected empty archive path, got %q", archived)
	}
}

func TestRotate_RenamesFile(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "audit.log")

	if err := os.WriteFile(logPath, []byte("data"), 0600); err != nil {
		t.Fatal(err)
	}

	archived, err := Rotate(logPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if archived == "" {
		t.Fatal("expected non-empty archive path")
	}

	if _, err := os.Stat(logPath); !os.IsNotExist(err) {
		t.Error("original log file should no longer exist")
	}

	if _, err := os.Stat(archived); err != nil {
		t.Errorf("archived file not found: %v", err)
	}

	if !strings.HasPrefix(filepath.Base(archived), "audit.log.") {
		t.Errorf("unexpected archive name: %s", archived)
	}
}

func TestRotate_OriginalDataPreserved(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "audit.log")
	original := []byte("important audit data")

	if err := os.WriteFile(logPath, original, 0600); err != nil {
		t.Fatal(err)
	}

	archived, err := Rotate(logPath)
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(archived)
	if err != nil {
		t.Fatal(err)
	}

	if string(data) != string(original) {
		t.Errorf("data mismatch: got %q want %q", data, original)
	}
}
