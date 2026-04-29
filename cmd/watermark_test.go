package cmd_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/yourusername/vaultpull/internal/audit"
)

func newWatermarkCmd() (*cobra.Command, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	cmd := &cobra.Command{
		Use: "watermark",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SetOut(buf)
			return nil
		},
	}
	cmd.SetOut(buf)
	return cmd, buf
}

func writeWatermarkEntries(t *testing.T, path string, entries []audit.Entry) {
	t.Helper()
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	for _, e := range entries {
		if err := enc.Encode(e); err != nil {
			t.Fatalf("encode: %v", err)
		}
	}
}

func TestRunWatermark_NoEntries(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "audit.log")

	entries, _ := audit.ReadAll(logPath)
	wm := audit.ComputeWatermarks(entries)
	if len(wm) != 0 {
		t.Errorf("expected no watermarks for empty log")
	}
}

func TestRunWatermark_ShowsPath(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "audit.log")
	now := time.Now().UTC()
	writeWatermarkEntries(t, logPath, []audit.Entry{
		{Path: "secret/myapp", Action: "sync", Timestamp: now, Keys: []string{"DB_URL"}},
	})

	entries, err := audit.ReadAll(logPath)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	wm := audit.ComputeWatermarks(entries)
	if len(wm) != 1 {
		t.Fatalf("expected 1 watermark, got %d", len(wm))
	}
	if wm[0].Path != "secret/myapp" {
		t.Errorf("unexpected path: %s", wm[0].Path)
	}
}

func TestRunWatermark_SavePersists(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "audit.log")
	now := time.Now().UTC().Truncate(time.Second)
	writeWatermarkEntries(t, logPath, []audit.Entry{
		{Path: "secret/svc", Action: "sync", Timestamp: now, Keys: []string{"API_KEY"}},
	})

	entries, _ := audit.ReadAll(logPath)
	wm := audit.ComputeWatermarks(entries)
	if err := audit.SaveWatermarks(dir, wm); err != nil {
		t.Fatalf("SaveWatermarks: %v", err)
	}
	loaded, err := audit.LoadWatermarks(dir)
	if err != nil {
		t.Fatalf("LoadWatermarks: %v", err)
	}
	if len(loaded) != 1 {
		t.Fatalf("expected 1 loaded watermark, got %d", len(loaded))
	}
	if !strings.Contains(loaded[0].Path, "svc") {
		t.Errorf("unexpected path in loaded watermark: %s", loaded[0].Path)
	}
}
