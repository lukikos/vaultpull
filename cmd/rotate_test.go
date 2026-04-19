package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunRotate_NoFile(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "audit.log")

	buf := &bytes.Buffer{}
	rotateCmd.SetOut(buf)

	rotateCmd.Flags().Set("log", logPath) //nolint

	err := runRotate(rotateCmd, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunRotate_ArchivesLog(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "audit.log")

	if err := os.WriteFile(logPath, []byte("entry"), 0600); err != nil {
		t.Fatal(err)
	}

	rotateCmd.Flags().Set("log", logPath) //nolint

	err := runRotate(rotateCmd, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(logPath); !os.IsNotExist(err) {
		t.Error("original log should be gone after rotate")
	}

	entries, _ := filepath.Glob(logPath + ".*")
	if len(entries) == 0 {
		t.Error("expected at least one archived file")
	}
	for _, e := range entries {
		if !strings.Contains(e, "audit.log.") {
			t.Errorf("unexpected archive path: %s", e)
		}
	}
}
