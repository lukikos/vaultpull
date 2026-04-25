package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yourorg/vaultpull/internal/audit"
)

func TestRunSnapshotShow_NotFound(t *testing.T) {
	origDir, _ := os.Getwd()
	dir := t.TempDir()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	cmd := snapshotShowCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := runSnapshotShow(cmd, []string{"nonexistent"})
	if err == nil {
		t.Fatal("expected error for missing snapshot")
	}
	if !strings.Contains(err.Error(), "nonexistent") {
		t.Errorf("error should mention snapshot name, got: %v", err)
	}
}

func TestRunSnapshotShow_DisplaysKeys(t *testing.T) {
	origDir, _ := os.Getwd()
	dir := t.TempDir()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	secrets := map[string]string{"DB_HOST": "localhost", "DB_PORT": "5432"}
	if err := audit.SaveSnapshot(".", "demo", "secret/app", secrets); err != nil {
		t.Fatalf("setup: %v", err)
	}

	cmd := snapshotShowCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	if err := runSnapshotShow(cmd, []string{"demo"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSnapshotFile_CreatedInCurrentDir(t *testing.T) {
	origDir, _ := os.Getwd()
	dir := t.TempDir()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	secrets := map[string]string{"KEY": "value"}
	if err := audit.SaveSnapshot(".", "mysnap", "secret/test", secrets); err != nil {
		t.Fatalf("save: %v", err)
	}

	expected := filepath.Join(dir, ".vaultpull.snapshot.mysnap.json")
	if _, err := os.Stat(expected); err != nil {
		t.Errorf("expected snapshot file at %s: %v", expected, err)
	}
}
