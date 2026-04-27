package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"

	"github.com/yourorg/vaultpull/internal/audit"
)

func newTrendCmd() (*cobra.Command, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	cmd := &cobra.Command{
		Use:  "trend",
		RunE: runTrend,
	}
	cmd.Flags().IntVar(&trendDays, "days", 30, "")
	cmd.SetOut(buf)
	return cmd, buf
}

func writeTrendEntries(t *testing.T, dir string, entries []audit.Entry) {
	t.Helper()
	logPath := filepath.Join(dir, "audit.log")
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	for _, e := range entries {
		if err := enc.Encode(e); err != nil {
			t.Fatal(err)
		}
	}
}

func TestRunTrend_NoEntries(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	cmd, buf := newTrendCmd()
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "No sync activity") {
		t.Errorf("expected no-activity message, got: %s", buf.String())
	}
}

func TestRunTrend_ShowsPath(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	now := time.Now().UTC()
	entries := []audit.Entry{
		{Path: "secret/myapp", Action: "sync", Key: "DB_PASS", Timestamp: now},
		{Path: "secret/myapp", Action: "sync", Key: "API_KEY", Timestamp: now},
	}
	writeTrendEntries(t, dir, entries)

	cmd, buf := newTrendCmd()
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "secret/myapp") {
		t.Errorf("expected path in output, got: %s", out)
	}
	if !strings.Contains(out, "avg/day") {
		t.Errorf("expected avg/day label, got: %s", out)
	}
}

func TestRunTrend_DirectionLabel(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	base := time.Now().UTC().AddDate(0, 0, -5)
	var entries []audit.Entry
	for i := 0; i < 4; i++ {
		for j := 0; j <= i; j++ {
			entries = append(entries, audit.Entry{
				Path:      "secret/grow",
				Action:    "sync",
				Key:       "X",
				Timestamp: base.Add(time.Duration(i) * 24 * time.Hour),
			})
		}
	}
	writeTrendEntries(t, dir, entries)

	cmd, buf := newTrendCmd()
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "increasing") {
		t.Errorf("expected increasing trend label, got: %s", buf.String())
	}
}
