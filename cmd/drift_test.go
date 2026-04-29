package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/cobra"

	"github.com/yourusername/vaultpull/internal/audit"
)

func newDriftCmd() (*cobra.Command, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	cmd := &cobra.Command{Use: "drift", RunE: runDrift}
	cmd.Flags().Float64("max-age", 48, "")
	cmd.Flags().Int("min-syncs", 1, "")
	cmd.Flags().String("log", "", "")
	cmd.SetOut(buf)
	return cmd, buf
}

func writeDriftEntries(t *testing.T, path string, entries []audit.Entry) {
	t.Helper()
	logger, err := audit.NewLogger(path)
	if err != nil {
		t.Fatalf("NewLogger: %v", err)
	}
	for _, e := range entries {
		if err := logger.Record(e.Path, e.Action, e.Key, e.Value); err != nil {
			t.Fatalf("Record: %v", err)
		}
	}
}

func TestRunDrift_NoEntries(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "audit.log")

	cmd, buf := newDriftCmd()
	cmd.Flags().Set("log", logPath)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !bytes.Contains(buf.Bytes(), []byte("No sync activity")) {
		t.Errorf("expected 'No sync activity' message, got: %s", buf)
	}
}

func TestRunDrift_ShowsFreshPath(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "audit.log")

	f, _ := os.Create(logPath)
	e := audit.Entry{Path: "secret/app", Action: "sync", Key: "K", Value: "V", Timestamp: time.Now().Add(-1 * time.Hour)}
	b, _ := json.Marshal(e)
	f.Write(append(b, '\n'))
	f.Close()

	cmd, buf := newDriftCmd()
	cmd.Flags().Set("log", logPath)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !bytes.Contains(buf.Bytes(), []byte("secret/app")) {
		t.Errorf("expected path in output, got: %s", buf)
	}
	if !bytes.Contains(buf.Bytes(), []byte("OK")) {
		t.Errorf("expected OK status for fresh path, got: %s", buf)
	}
}
